package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type NotificationService struct {
	repo ports.INotificationRepository
}

func NewNotificationService(repo ports.INotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) CreateNotification(ctx context.Context, userID string, typ domain.NotificationType, payload []byte) error {
	_, err := s.repo.CreateNotification(ctx, &domain.Notification{
		UserID:  userID,
		Type:    typ,
		Payload: payload,
	})
	return err
}

func (s *NotificationService) GetNotificationsForUser(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkNotificationRead(ctx context.Context, notificationID, userID string) error {
	return s.repo.MarkNotificationRead(ctx, notificationID, userID)
}

func (s *NotificationService) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnreadNotifications(ctx, userID)
}
