package models

import "time"

// Channel represents a Telegram channel linked to a user.
type Channel struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	ChatID    int64     `json:"chat_id"`
	Title     string    `json:"title"`
	Username  string    `json:"username,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
