package postgres

import (
	"context"
	"encoding/json"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	q *generated.Queries
}

func NewNotificationRepository(conn *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{
		q: generated.New(conn),
	}
}

func mapNotification(n generated.Notification) *domain.Notification {
	return &domain.Notification{
		NotificationID: uuidToString(n.NotificationID),
		UserID:         uuidToString(n.UserID),
		Type:           domain.NotificationType(n.Type),
		Payload:        json.RawMessage(n.Payload),
		IsRead:         n.IsRead,
		CreatedAt:      n.CreatedAt.Time,
	}
}

func (r *NotificationRepository) CreateNotification(ctx context.Context, notif *domain.Notification) (*domain.Notification, error) {
	n, err := r.q.CreateNotification(ctx, generated.CreateNotificationParams{
		UserID:  toUUID(notif.UserID),
		Type:    string(notif.Type),
		Payload: []byte(notif.Payload),
	})
	if err != nil {
		return nil, err
	}
	return mapNotification(n), nil
}

func (r *NotificationRepository) GetNotificationsForUser(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	notifs, err := r.q.GetNotificationsForUser(ctx, generated.GetNotificationsForUserParams{
		UserID: toUUID(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.Notification, len(notifs))
	for i, n := range notifs {
		res[i] = *mapNotification(n)
	}
	return res, nil
}

func (r *NotificationRepository) MarkNotificationRead(ctx context.Context, notificationID, userID string) error {
	return r.q.MarkNotificationRead(ctx, generated.MarkNotificationReadParams{
		NotificationID: toUUID(notificationID),
		UserID:         toUUID(userID),
	})
}

func (r *NotificationRepository) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	count, err := r.q.CountUnreadNotifications(ctx, toUUID(userID))
	return int(count), err
}
