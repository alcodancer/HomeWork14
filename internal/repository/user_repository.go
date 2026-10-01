package repository

import (
	"context"

	"example/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) (string, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.User, error)
}