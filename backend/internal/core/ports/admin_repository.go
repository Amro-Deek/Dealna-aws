package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IAdminRepository interface {
	GetDashboardMetrics(ctx context.Context, universityID string) (*domain.DashboardMetrics, error)
	GetUsers(ctx context.Context, search string, roleFilter string, statusFilter string, limit int, offset int) ([]domain.AdminUserSnapshot, int, error)
	GetVerifications(ctx context.Context, status string) ([]domain.AdminProviderVerification, error)
	ApproveVerification(ctx context.Context, applicationID string, adminID string) error
	RejectVerification(ctx context.Context, applicationID string, adminID string, comment string) error
	WarnUser(ctx context.Context, adminID string, targetUserID string, reason string) (int, error)
	BanUser(ctx context.Context, targetUserID string) error
	GetVerificationDocuments(ctx context.Context, applicationID string) ([]domain.AdminProviderDocument, error)
	GetApplicantEmail(ctx context.Context, applicationID string) (string, error)
	GetKeycloakSub(ctx context.Context, userID string) (string, error)
	GetAdminUserProfileStats(ctx context.Context, userID string) (int, int, int, error)
}
