package ports

import (
	"context"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IStudentPreRegistrationRepository interface {
	Create(ctx context.Context, pre *domain.StudentPreRegistration) error
	GetByToken(ctx context.Context, token string) (*domain.StudentPreRegistration, error)
	GetByEmail(ctx context.Context, email string) (*domain.StudentPreRegistration, error)
	Update(ctx context.Context, pre *domain.StudentPreRegistration) error
}
