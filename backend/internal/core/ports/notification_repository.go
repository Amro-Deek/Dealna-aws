package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type INotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) (*domain.Notification, error)
	GetNotificationsForUser(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error)
	MarkNotificationRead(ctx context.Context, notificationID, userID string) error
	CountUnreadNotifications(ctx context.Context, userID string) (int, error)
}
