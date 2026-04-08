package ports

import (
	"context"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IUserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByKeycloakSub(ctx context.Context, sub string) (*domain.User, error)

	CreateStudent(
		ctx context.Context,
		displayName string,
		email string,
		keycloakSub string,
		major *string,
		year *int,
		universityID string,
		studentID string,
	) (*domain.User, error)

	GetProfile(ctx context.Context, userID string) (*domain.Profile, *domain.Student, error)
	UpdateProfile(ctx context.Context, userID string, displayName, bio, profilePictureURL *string, displayNameLastChangedAt *time.Time) error
	UpdateStudent(ctx context.Context, userID string, major *string, year *int) error
}