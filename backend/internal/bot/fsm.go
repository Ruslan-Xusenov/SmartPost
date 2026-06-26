package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StateIdle          = ""
	StateSelectChannel = "select_channel"
	StateWaitMedia     = "wait_media"
	StateWaitCaption   = "wait_caption"
	StateWaitButtons   = "wait_buttons"
	StateWaitSchedule  = "wait_schedule"
	StateConfirm       = "confirm"
)

type FSMData struct {
	State     string `json:"state"`
	ChannelID int64  `json:"channel_id,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	FileID    string `json:"file_id,omitempty"`
	Caption   string `json:"caption,omitempty"`
	Buttons   string `json:"buttons,omitempty"`
	Schedule  string `json:"schedule,omitempty"`
}

type FSM struct {
	rdb *redis.Client
}

func NewFSM(rdb *redis.Client) *FSM {
	return &FSM{rdb: rdb}
}

func fsmKey(telegramID int64) string {
	return fmt.Sprintf("fsm:%d", telegramID)
}

func (f *FSM) Get(ctx context.Context, telegramID int64) (*FSMData, error) {
	data, err := f.rdb.Get(ctx, fsmKey(telegramID)).Result()
	if err == redis.Nil {
		return &FSMData{State: StateIdle}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fsm get error: %w", err)
	}
	var fsmData FSMData
	if err := json.Unmarshal([]byte(data), &fsmData); err != nil {
		return nil, fmt.Errorf("fsm unmarshal error: %w", err)
	}
	return &fsmData, nil
}

func (f *FSM) Set(ctx context.Context, telegramID int64, data *FSMData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("fsm marshal error: %w", err)
	}
	return f.rdb.Set(ctx, fsmKey(telegramID), string(jsonData), 30*time.Minute).Err()
}

func (f *FSM) Clear(ctx context.Context, telegramID int64) error {
	return f.rdb.Del(ctx, fsmKey(telegramID)).Err()
}

func (f *FSM) SetState(ctx context.Context, telegramID int64, state string) error {
	data, err := f.Get(ctx, telegramID)
	if err != nil {
		return err
	}
	data.State = state
	return f.Set(ctx, telegramID, data)
}
