package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IUserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateStudent(
	ctx context.Context,
	displayName string,
	email string,
	passwordHash string,
	major *string,
	year *int,
	universityID string,
	studentID string,
) (*domain.User, error)

}
