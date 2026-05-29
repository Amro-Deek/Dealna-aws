package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IProviderRepository interface {
	CreateProviderApplication(ctx context.Context, applicantID, universityID, businessName string, phoneNumber, businessType, address *string, status string) (*domain.ProviderApplication, error)
	GetProviderApplicationByApplicantID(ctx context.Context, applicantID string) (*domain.ProviderApplication, error)
	UpdateProviderApplicationStatus(ctx context.Context, applicationID, status string) error
	UpdateProviderApplicationReview(ctx context.Context, applicationID, status, adminID, comment string) error
	CreateProviderApplicationDocument(ctx context.Context, applicationID, filePath, documentType, originalFilename, contentType string, sizeBytes int64, uploadStatus string) (*domain.ProviderApplicationDocument, error)
	GetProviderApplicationDocuments(ctx context.Context, applicationID string) ([]domain.ProviderApplicationDocument, error)
	CreateProvider(ctx context.Context, userID, businessName string, phoneNumber, businessType, address *string) (*domain.Provider, error)
}
