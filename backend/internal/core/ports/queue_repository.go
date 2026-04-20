package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IQueueRepository interface {
	JoinQueue(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error)
	GetQueueByItem(ctx context.Context, itemID string) ([]domain.QueueEntry, error)
	GetQueuePosition(ctx context.Context, itemID, entryID string) (int, error)
	GetFrontOfQueue(ctx context.Context, itemID string) (*domain.QueueEntry, error)
	UpdateEntryStatus(ctx context.Context, entryID string, status domain.QueueEntryStatus) error
	SetTurnStarted(ctx context.Context, entryID string) error
	GetExpiredTurns(ctx context.Context) ([]domain.QueueEntry, error)
	RemoveFromQueue(ctx context.Context, entryID string) error
	GetQueueEntriesByUser(ctx context.Context, userID string) ([]domain.QueueEntry, error)
	CountQueueEntries(ctx context.Context, itemID string) (int, error)
	LeaveQueue(ctx context.Context, itemID, userID string) error
}
