package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IAdminService interface {
	GetDashboardMetrics(ctx context.Context, universityID string) (*domain.DashboardMetrics, error)
	GetUsers(ctx context.Context, search string, roleFilter string, statusFilter string, limit int, offset int) ([]domain.AdminUserSnapshot, int, error)
	GetVerifications(ctx context.Context, status string) ([]domain.AdminProviderVerification, error)
	ApproveVerification(ctx context.Context, applicationID string, adminID string) error
	RejectVerification(ctx context.Context, applicationID string, adminID string, comment string) error
	WarnUser(ctx context.Context, adminID string, targetUserID string, reason string) error
	WarnItemOwner(ctx context.Context, adminID string, itemIDStr string, reason string) error
	BanUser(ctx context.Context, adminID string, targetUserID string, reason string) error
	GetVerificationDocuments(ctx context.Context, applicationID string) ([]domain.AdminProviderDocument, error)
	GetAdminUserProfileStats(ctx context.Context, userID string) (int, int, int, error)
	DeleteItemAdmin(ctx context.Context, adminID string, itemID string, reason string) error
}
