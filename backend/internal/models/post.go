package models

import "time"

// PostStatus represents the lifecycle state of a post.
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusSent      PostStatus = "sent"
	PostStatusFailed    PostStatus = "failed"
)

// MediaType represents the type of media attached to a post.
type MediaType string

const (
	MediaTypeText      MediaType = "text"
	MediaTypePhoto     MediaType = "photo"
	MediaTypeVideo     MediaType = "video"
	MediaTypeVideoNote MediaType = "video_note"
)

// Post represents a channel post with optional media and scheduled time.
type Post struct {
	ID           int64      `json:"id"`
	ChannelID    int64      `json:"channel_id"`
	UserID       int64      `json:"user_id"`
	MediaType    MediaType  `json:"media_type"`
	FileID       string     `json:"file_id,omitempty"`
	Caption      string     `json:"caption,omitempty"`
	Status       PostStatus `json:"status"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Populated separately
	Buttons []Button `json:"buttons,omitempty"`
}
