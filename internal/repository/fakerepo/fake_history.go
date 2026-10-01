package fakerepo

import (
	"context"

	"example/internal/repository"
)

type FakeHistoryRepo struct {
	Entries map[string][]*repository.HistoryEntry
	Err     error
}

func NewFakeHistoryRepo() *FakeHistoryRepo {
	return &FakeHistoryRepo{Entries: map[string][]*repository.HistoryEntry{}}
}

func (h *FakeHistoryRepo) Record(_ context.Context, e *repository.HistoryEntry) error {
	if h.Err != nil {
		return h.Err
	}
	cp := *e
	h.Entries[e.UserID] = append(h.Entries[e.UserID], &cp)
	return nil
}

func (h *FakeHistoryRepo) GetHistory(_ context.Context, userID string) ([]*repository.HistoryEntry, error) {
	if h.Err != nil {
		return nil, h.Err
	}
	return h.Entries[userID], nil
}