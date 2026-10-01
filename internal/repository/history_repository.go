package repository

import "context"

type HistoryEntry struct {
	Action    string      `json:"action"`
	UserID    string      `json:"user_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type HistoryRepository interface {
	Record(ctx context.Context, entry *HistoryEntry) error
	GetHistory(ctx context.Context, userID string) ([]*HistoryEntry, error)
}