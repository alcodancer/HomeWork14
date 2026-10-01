package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"example/internal/repository"
)

type HistoryRepo struct {
	client redis.UniversalClient
	ttl    time.Duration
}

func NewHistoryRepo(client redis.UniversalClient, ttl time.Duration) *HistoryRepo {
	return &HistoryRepo{client: client, ttl: ttl}
}

func (h *HistoryRepo) Record(ctx context.Context, entry *repository.HistoryEntry) error {
	entry.Timestamp = time.Now().UnixNano()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("history:%s:%d", entry.UserID, entry.Timestamp)

	// SET с TTL — ключ истечёт через h.ttl
	return h.client.Set(ctx, key, data, h.ttl).Err()
}

func (h *HistoryRepo) GetHistory(ctx context.Context, userID string) ([]*repository.HistoryEntry, error) {
	pattern := fmt.Sprintf("history:%s:*", userID)

	var entries []*repository.HistoryEntry
	iter := h.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		raw, err := h.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue // ключ мог истечь между SCAN и GET
		}
		var entry repository.HistoryEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}