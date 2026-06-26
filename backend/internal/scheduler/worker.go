package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	mdl "github.com/smartpost/backend/internal/models"
	"github.com/smartpost/backend/internal/telegram"
)

// Worker processes scheduled post-send tasks.
type Worker struct {
	db     *pgxpool.Pool
	sender *telegram.Sender
}

// NewWorker creates a new task worker.
func NewWorker(db *pgxpool.Pool, botAPI *bot.Bot) *Worker {
	return &Worker{
		db:     db,
		sender: telegram.NewSender(botAPI, db),
	}
}

// Start launches the Asynq worker server.
func (w *Worker) Start(redisAddr string) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"posts":   6,
				"default": 3,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n*n) * time.Second // Exponential backoff
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypePostSend, w.handlePostSend)

	log.Println("🚀 Asynq worker started")
	return srv.Start(mux)
}

// handlePostSend processes a post:send task.
func (w *Worker) handlePostSend(ctx context.Context, t *asynq.Task) error {
	var payload PostSendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("📤 Processing post:send for post ID: %d", payload.PostID)

	// Fetch post from database
	post, err := w.getPost(ctx, payload.PostID)
	if err != nil {
		w.markFailed(ctx, payload.PostID, err.Error())
		return err
	}

	// Send the post via Telegram API
	if err := w.sender.SendPost(ctx, post); err != nil {
		w.markFailed(ctx, payload.PostID, err.Error())
		return err
	}

	// Mark as sent
	_, err = w.db.Exec(ctx,
		`UPDATE posts SET status = 'sent', sent_at = NOW(), updated_at = NOW() WHERE id = $1`,
		payload.PostID,
	)
	if err != nil {
		log.Printf("❌ Failed to update post status: %v", err)
	}

	log.Printf("✅ Post %d sent successfully", payload.PostID)
	return nil
}

// getPost retrieves a post with its buttons from the database.
func (w *Worker) getPost(ctx context.Context, postID int64) (*mdl.Post, error) {
	post := &mdl.Post{}
	err := w.db.QueryRow(ctx,
		`SELECT id, channel_id, user_id, media_type, file_id, caption, status
		 FROM posts WHERE id = $1`,
		postID,
	).Scan(&post.ID, &post.ChannelID, &post.UserID, &post.MediaType,
		&post.FileID, &post.Caption, &post.Status)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Fetch buttons
	rows, err := w.db.Query(ctx,
		`SELECT id, post_id, text, url, color_code, row_index, col_index
		 FROM buttons WHERE post_id = $1 ORDER BY row_index, col_index`,
		postID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buttons: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var btn mdl.Button
		if err := rows.Scan(&btn.ID, &btn.PostID, &btn.Text, &btn.URL,
			&btn.ColorCode, &btn.RowIndex, &btn.ColIndex); err != nil {
			continue
		}
		post.Buttons = append(post.Buttons, btn)
	}

	return post, nil
}

// markFailed updates a post's status to 'failed' with an error message.
func (w *Worker) markFailed(ctx context.Context, postID int64, errMsg string) {
	_, err := w.db.Exec(ctx,
		`UPDATE posts SET status = 'failed', error_message = $1, updated_at = NOW() WHERE id = $2`,
		errMsg, postID,
	)
	if err != nil {
		log.Printf("❌ Failed to mark post %d as failed: %v", postID, err)
	}
}
