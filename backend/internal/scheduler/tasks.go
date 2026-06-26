package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Task type constants.
const (
	TypePostSend = "post:send"
)

// PostSendPayload contains the data needed to send a post.
type PostSendPayload struct {
	PostID int64 `json:"post_id"`
}

// NewPostSendTask creates a new task for sending a post.
func NewPostSendTask(postID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(PostSendPayload{PostID: postID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}
	return asynq.NewTask(TypePostSend, payload), nil
}
