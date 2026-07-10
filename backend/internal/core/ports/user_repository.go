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

	CreateApplicantUser(
		ctx context.Context,
		email string,
		keycloakSub string,
	) (*domain.User, error)

	UpdateUserRole(ctx context.Context, userID string, role string) error
	UpdateUserStatus(ctx context.Context, userID string, status string) error

	GetProfile(ctx context.Context, userID string) (*domain.Profile, *domain.Student, error)
	GetProfileByProfileID(ctx context.Context, profileID string) (*domain.Profile, error)
	GetProfileByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID string, displayName, bio, profilePictureURL *string, displayNameLastChangedAt *time.Time) error
	UpdateStudent(ctx context.Context, userID string, major *string, year *int) error
	UpdateDeviceToken(ctx context.Context, userID string, token string) error
	CreateProfileForUser(ctx context.Context, userID string, displayName string) error

	GetAdminUserProfileStats(ctx context.Context, userID string) (int, int, int, error)
}
