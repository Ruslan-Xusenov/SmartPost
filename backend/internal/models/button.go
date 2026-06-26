package models

import (
	"strings"
	"time"
)

// Button represents an inline keyboard button attached to a post.
type Button struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	Text      string    `json:"text"`
	URL       string    `json:"url"`
	ColorCode string    `json:"color_code"`
	RowIndex  int       `json:"row_index"`
	ColIndex  int       `json:"col_index"`
	CreatedAt time.Time `json:"created_at"`
}

// colorEmojiMap maps color names to their emoji representation.
var colorEmojiMap = map[string]string{
	"green":  "🟢",
	"red":    "🔴",
	"blue":   "🔵",
	"yellow": "🟡",
	"orange": "🟠",
	"purple": "🟣",
	"white":  "⚪",
	"black":  "⚫",
}

// DisplayText returns the button text with an optional color emoji prefix.
func (b *Button) DisplayText() string {
	emoji, ok := colorEmojiMap[strings.ToLower(b.ColorCode)]
	if ok {
		return emoji + " " + b.Text
	}
	return b.Text
}

// ParseButtonLine parses a single line in the format:
// [Tugma matni] - [Havola] - [Rang]
// Returns a Button struct with parsed values.
func ParseButtonLine(line string, rowIndex int) (*Button, error) {
	parts := strings.Split(line, " - ")
	if len(parts) < 2 {
		// Try splitting by " – " (em dash)
		parts = strings.Split(line, " – ")
	}

	btn := &Button{
		RowIndex: rowIndex,
		ColIndex: 0,
	}

	if len(parts) >= 1 {
		btn.Text = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		btn.URL = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		btn.ColorCode = strings.ToLower(strings.TrimSpace(parts[2]))
	} else {
		btn.ColorCode = "default"
	}

	return btn, nil
}

// ParseButtons parses multiple lines of button definitions.
// Each line format: [Tugma matni] - [Havola] - [Rang]
func ParseButtons(text string) []*Button {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var buttons []*Button

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		btn, err := ParseButtonLine(line, i)
		if err != nil {
			continue
		}
		buttons = append(buttons, btn)
	}

	return buttons
}
