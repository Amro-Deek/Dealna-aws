package ports

import (
    "context"
    "github.com/Amro-Deek/Dealna-aws/internal/core/domain"
)

type IUserRepository interface {
    GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}
