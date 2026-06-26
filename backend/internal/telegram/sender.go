package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	mdl "github.com/smartpost/backend/internal/models"
)

// Sender handles sending posts to Telegram channels.
type Sender struct {
	api     *bot.Bot
	db      *pgxpool.Pool
	limiter *RateLimiter
}

// NewSender creates a new Sender instance.
func NewSender(api *bot.Bot, db *pgxpool.Pool) *Sender {
	return &Sender{
		api:     api,
		db:      db,
		limiter: NewRateLimiter(),
	}
}

// SendPost sends a post to its target channel.
func (s *Sender) SendPost(ctx context.Context, post *mdl.Post) error {
	// Get channel chat_id
	var chatID int64
	err := s.db.QueryRow(ctx,
		`SELECT chat_id FROM channels WHERE id = $1 AND is_active = true`,
		post.ChannelID,
	).Scan(&chatID)
	if err != nil {
		return fmt.Errorf("channel not found or inactive: %w", err)
	}

	// Wait for rate limiter
	if err := s.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Build inline keyboard from buttons
	var keyboard *models.InlineKeyboardMarkup
	if len(post.Buttons) > 0 {
		keyboard = s.buildKeyboard(post.Buttons)
	}

	// Send based on media type
	switch mdl.MediaType(post.MediaType) {
	case mdl.MediaTypeText:
		return s.sendText(ctx, chatID, post, keyboard)
	case mdl.MediaTypePhoto:
		return s.sendPhoto(ctx, chatID, post, keyboard)
	case mdl.MediaTypeVideo:
		return s.sendVideo(ctx, chatID, post, keyboard)
	case mdl.MediaTypeVideoNote:
		return s.sendVideoNote(ctx, chatID, post, keyboard)
	default:
		return fmt.Errorf("unknown media type: %s", post.MediaType)
	}
}

func (s *Sender) sendText(ctx context.Context, chatID int64, post *mdl.Post, keyboard *models.InlineKeyboardMarkup) error {
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      post.Caption,
		ParseMode: models.ParseModeHTML,
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, err := s.api.SendMessage(ctx, params)
	if err != nil {
		return fmt.Errorf("sendMessage failed: %w", err)
	}
	log.Printf("✅ Text post sent to channel %d", chatID)
	return nil
}

func (s *Sender) sendPhoto(ctx context.Context, chatID int64, post *mdl.Post, keyboard *models.InlineKeyboardMarkup) error {
	params := &bot.SendPhotoParams{
		ChatID:    chatID,
		Photo:     &models.InputFileString{Data: post.FileID},
		Caption:   post.Caption,
		ParseMode: models.ParseModeHTML,
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, err := s.api.SendPhoto(ctx, params)
	if err != nil {
		return fmt.Errorf("sendPhoto failed: %w", err)
	}
	log.Printf("✅ Photo post sent to channel %d", chatID)
	return nil
}

func (s *Sender) sendVideo(ctx context.Context, chatID int64, post *mdl.Post, keyboard *models.InlineKeyboardMarkup) error {
	params := &bot.SendVideoParams{
		ChatID:    chatID,
		Video:     &models.InputFileString{Data: post.FileID},
		Caption:   post.Caption,
		ParseMode: models.ParseModeHTML,
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, err := s.api.SendVideo(ctx, params)
	if err != nil {
		return fmt.Errorf("sendVideo failed: %w", err)
	}
	log.Printf("✅ Video post sent to channel %d", chatID)
	return nil
}

func (s *Sender) sendVideoNote(ctx context.Context, chatID int64, post *mdl.Post, keyboard *models.InlineKeyboardMarkup) error {
	params := &bot.SendVideoNoteParams{
		ChatID:    chatID,
		VideoNote: &models.InputFileString{Data: post.FileID},
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, err := s.api.SendVideoNote(ctx, params)
	if err != nil {
		return fmt.Errorf("sendVideoNote failed: %w", err)
	}
	log.Printf("✅ VideoNote sent to channel %d", chatID)
	return nil
}

// buildKeyboard creates an InlineKeyboardMarkup from button models.
func (s *Sender) buildKeyboard(buttons []mdl.Button) *models.InlineKeyboardMarkup {
	// Group buttons by row_index
	rowMap := make(map[int][]models.InlineKeyboardButton)
	maxRow := 0
	for _, btn := range buttons {
		rowMap[btn.RowIndex] = append(rowMap[btn.RowIndex], models.InlineKeyboardButton{
			Text: btn.DisplayText(),
			URL:  btn.URL,
		})
		if btn.RowIndex > maxRow {
			maxRow = btn.RowIndex
		}
	}

	var rows [][]models.InlineKeyboardButton
	for i := 0; i <= maxRow; i++ {
		if row, ok := rowMap[i]; ok {
			rows = append(rows, row)
		}
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}
