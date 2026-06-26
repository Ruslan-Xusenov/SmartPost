package models

import "time"

// User represents a registered bot user.
type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username,omitempty"`
	FirstName  string    `json:"first_name,omitempty"`
	Timezone   string    `json:"timezone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
