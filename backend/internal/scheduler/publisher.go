package scheduler

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Publisher enqueues tasks to the Asynq task queue.
type Publisher struct {
	client *asynq.Client
}

// NewPublisher creates a new task publisher.
func NewPublisher(redisAddr string) *Publisher {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Publisher{client: client}
}

// EnqueuePostSend enqueues a post for immediate sending.
func (p *Publisher) EnqueuePostSend(postID int64) error {
	task, err := NewPostSendTask(postID)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(task,
		asynq.MaxRetry(3),
		asynq.Queue("posts"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue post send: %w", err)
	}
	return nil
}

// SchedulePostSend enqueues a post for sending at a specific time.
func (p *Publisher) SchedulePostSend(postID int64, sendAt time.Time) error {
	task, err := NewPostSendTask(postID)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(task,
		asynq.ProcessAt(sendAt),
		asynq.MaxRetry(3),
		asynq.Queue("posts"),
	)
	if err != nil {
		return fmt.Errorf("failed to schedule post send: %w", err)
	}
	return nil
}

// Close closes the publisher client.
func (p *Publisher) Close() error {
	return p.client.Close()
}
