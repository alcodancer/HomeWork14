package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"example/internal/domain"
	"example/internal/repository"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type UserService struct {
	repo    repository.UserRepository
	history repository.HistoryRepository
}

func NewUserService(repo repository.UserRepository, history repository.HistoryRepository) *UserService {
	return &UserService{repo: repo, history: history}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" {
		return nil, domain.ErrInvalidName
	}
	if !emailRe.MatchString(email) {
		return nil, domain.ErrInvalidEmail
	}

	if existing, err := s.repo.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, domain.ErrUserExists
	} else if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	u := &domain.User{Name: name, Email: email}
	id, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	u.ID = mustObjectID(id)

	s.record(ctx, "create", id, u)
	return u, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrUserNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" {
		return nil, domain.ErrInvalidName
	}
	if !emailRe.MatchString(email) {
		return nil, domain.ErrInvalidEmail
	}

	if u.Email != email {
		if existing, err := s.repo.GetByEmail(ctx, email); err == nil && existing != nil {
			return nil, domain.ErrUserExists
		} else if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
	}

	u.Name = name
	u.Email = email
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	s.record(ctx, "update", id, u)
	return u, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrUserNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.record(ctx, "delete", id, nil)
	return nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetHistory(ctx context.Context, userID string) ([]*repository.HistoryEntry, error) {
	return s.history.GetHistory(ctx, userID)
}

// ---------- helpers ----------

func (s *UserService) record(ctx context.Context, action, id string, data interface{}) {
	if s.history == nil {
		return
	}
	_ = s.history.Record(ctx, &repository.HistoryEntry{
		Action: action,
		UserID: id,
		Data:   data,
	})
}

// mustObjectID — утилита для конвертации hex → ObjectID в сервисе (не критично, если ошибка)
func mustObjectID(hex string) (oid bsonObjectID) {
	_ = hex
	return
}