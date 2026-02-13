package ports

import (
	"context"
	"time"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type ISessionRepository interface {
	Create(ctx context.Context, userID string, jti string, expiresAt time.Time) error
	GetByJTI(ctx context.Context, jti string) (*domain.Session, error)
	RevokeByJTI(ctx context.Context, jti string) error
	RevokeAllForUser(ctx context.Context,userID string) error

}
