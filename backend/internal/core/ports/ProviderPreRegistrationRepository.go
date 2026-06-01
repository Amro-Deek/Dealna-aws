package ports

import (
	"context"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IProviderPreRegistrationRepository interface {
	Create(ctx context.Context, pre *domain.ProviderPreRegistration) error
	GetByToken(ctx context.Context, token string) (*domain.ProviderPreRegistration, error)
	GetByEmail(ctx context.Context, email string) (*domain.ProviderPreRegistration, error)
	Update(ctx context.Context, pre *domain.ProviderPreRegistration) error
}
