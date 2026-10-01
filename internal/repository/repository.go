// internal/repository/user_repository.go
package repository

import (
    "context"
    "example/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, u *domain.User) (int64, error)
    GetByID(ctx context.Context, id int64) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, u *domain.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context) ([]*domain.User, error)
}