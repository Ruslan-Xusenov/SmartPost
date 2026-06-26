package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	mdl "github.com/smartpost/backend/internal/models"
)

type createPostReq struct {
	ChannelID int64  `json:"channel_id"`
	MediaType string `json:"media_type"`
	FileID    string `json:"file_id"`
	Caption   string `json:"caption"`
	Buttons   []struct {
		Text      string `json:"text"`
		URL       string `json:"url"`
		ColorCode string `json:"color_code"`
		RowIndex  int    `json:"row_index"`
	} `json:"buttons"`
}

type scheduleReq struct {
	ScheduledAt string `json:"scheduled_at"` // ISO 8601 format
}

// createPost creates a new post in draft status.
func (s *Server) createPost(w http.ResponseWriter, r *http.Request) {
	telegramID := getTelegramID(r)

	var req createPostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user DB ID
	var userDBID int64
	err := s.db.QueryRow(r.Context(),
		`SELECT id FROM users WHERE telegram_id = $1`, telegramID,
	).Scan(&userDBID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Insert post
	var postID int64
	err = s.db.QueryRow(r.Context(),
		`INSERT INTO posts (channel_id, user_id, media_type, file_id, caption, status)
		 VALUES ($1, $2, $3, $4, $5, 'draft') RETURNING id`,
		req.ChannelID, userDBID, req.MediaType, req.FileID, req.Caption,
	).Scan(&postID)
	if err != nil {
		log.Printf("❌ Post insert error: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Insert buttons
	for _, btn := range req.Buttons {
		_, err := s.db.Exec(r.Context(),
			`INSERT INTO buttons (post_id, text, url, color_code, row_index)
			 VALUES ($1, $2, $3, $4, $5)`,
			postID, btn.Text, btn.URL, btn.ColorCode, btn.RowIndex,
		)
		if err != nil {
			log.Printf("❌ Button insert error: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": postID})
}

// listPosts returns posts filtered by channel and/or status.
func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	telegramID := getTelegramID(r)
	status := r.URL.Query().Get("status")
	channelID := r.URL.Query().Get("channel_id")

	query := `SELECT p.id, p.channel_id, p.media_type, p.file_id, p.caption,
	           p.status, p.scheduled_at, p.sent_at, p.error_message, p.created_at
	          FROM posts p
	          JOIN users u ON u.id = p.user_id
	          WHERE u.telegram_id = $1`
	args := []interface{}{telegramID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if channelID != "" {
		query += fmt.Sprintf(" AND p.channel_id = $%d", argIdx)
		args = append(args, channelID)
		argIdx++
	}
	query += " ORDER BY p.created_at DESC LIMIT 50"

	rows, err := s.db.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []mdl.Post
	for rows.Next() {
		var p mdl.Post
		err := rows.Scan(&p.ID, &p.ChannelID, &p.MediaType, &p.FileID, &p.Caption,
			&p.Status, &p.ScheduledAt, &p.SentAt, &p.ErrorMessage, &p.CreatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []mdl.Post{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// getPost returns a single post by ID with its buttons.
func (s *Server) getPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var p mdl.Post
	err = s.db.QueryRow(r.Context(),
		`SELECT p.id, p.channel_id, p.media_type, p.file_id, p.caption,
		        p.status, p.scheduled_at, p.sent_at, p.error_message, p.created_at
		 FROM posts p WHERE p.id = $1`, postID,
	).Scan(&p.ID, &p.ChannelID, &p.MediaType, &p.FileID, &p.Caption,
		&p.Status, &p.ScheduledAt, &p.SentAt, &p.ErrorMessage, &p.CreatedAt)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Fetch buttons
	rows, err := s.db.Query(r.Context(),
		`SELECT id, post_id, text, url, color_code, row_index, col_index
		 FROM buttons WHERE post_id = $1 ORDER BY row_index, col_index`, postID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var btn mdl.Button
			if err := rows.Scan(&btn.ID, &btn.PostID, &btn.Text, &btn.URL,
				&btn.ColorCode, &btn.RowIndex, &btn.ColIndex); err != nil {
				continue
			}
			p.Buttons = append(p.Buttons, btn)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// updatePost updates a draft post.
func (s *Server) updatePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req createPostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(r.Context(),
		`UPDATE posts SET media_type = $1, file_id = $2, caption = $3, updated_at = NOW()
		 WHERE id = $4 AND status = 'draft'`,
		req.MediaType, req.FileID, req.Caption, postID,
	)
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	// Replace buttons
	s.db.Exec(r.Context(), `DELETE FROM buttons WHERE post_id = $1`, postID)
	for _, btn := range req.Buttons {
		s.db.Exec(r.Context(),
			`INSERT INTO buttons (post_id, text, url, color_code, row_index)
			 VALUES ($1, $2, $3, $4, $5)`,
			postID, btn.Text, btn.URL, btn.ColorCode, btn.RowIndex,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// deletePost removes a draft post.
func (s *Server) deletePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(r.Context(),
		`DELETE FROM posts WHERE id = $1 AND status = 'draft'`, postID)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// sendPost triggers immediate sending of a post.
func (s *Server) sendPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Update status to scheduled
	_, err = s.db.Exec(r.Context(),
		`UPDATE posts SET status = 'scheduled', updated_at = NOW() WHERE id = $1`, postID)
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	// Enqueue for immediate sending
	if err := s.publisher.EnqueuePostSend(postID); err != nil {
		log.Printf("❌ Failed to enqueue post: %v", err)
		http.Error(w, "Failed to enqueue post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sending"})
}

// schedulePost schedules a post for a future time.
func (s *Server) schedulePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req scheduleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		// Try alternative format
		scheduledAt, err = time.Parse("2006-01-02 15:04", req.ScheduledAt)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
	}

	_, err = s.db.Exec(r.Context(),
		`UPDATE posts SET status = 'scheduled', scheduled_at = $1, updated_at = NOW()
		 WHERE id = $2`, scheduledAt, postID)
	if err != nil {
		http.Error(w, "Failed to schedule post", http.StatusInternalServerError)
		return
	}

	if err := s.publisher.SchedulePostSend(postID, scheduledAt); err != nil {
		log.Printf("❌ Failed to schedule post: %v", err)
		http.Error(w, "Failed to schedule post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "scheduled",
		"scheduled_at": scheduledAt.Format(time.RFC3339),
	})
}
