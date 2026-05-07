package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IQueueRepository interface {
	JoinQueueAtomic(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error)
	GetJoinEligibility(ctx context.Context, itemID, userID string) (int, bool, bool, error)
	GetQueueByItem(ctx context.Context, itemID string) ([]domain.QueueEntry, error)
	GetQueuePosition(ctx context.Context, itemID, entryID string) (int, error)
	GetFrontOfQueue(ctx context.Context, itemID string) (*domain.QueueEntry, error)
	GetActiveEntryByItemAndUser(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error)
	UpdateEntryStatus(ctx context.Context, entryID string, status domain.QueueEntryStatus) error
	SetTurnStarted(ctx context.Context, entryID string) error
	ExpireReservedEntries(ctx context.Context) ([]domain.QueueEntry, error)
	ExpireConfirmedEntries(ctx context.Context) ([]domain.QueueEntry, error)
	AutoCompleteHandedOffEntries(ctx context.Context) ([]domain.QueueEntry, error)
	CancelAllQueueEntries(ctx context.Context, itemID string) error
	RemoveFromQueue(ctx context.Context, entryID string) error
	GetQueueEntriesByUser(ctx context.Context, userID string) ([]domain.QueueEntry, error)
	CountQueueEntries(ctx context.Context, itemID string) (int, error)
	LeaveQueue(ctx context.Context, itemID, userID string) error
	GetEntryByID(ctx context.Context, entryID string) (*domain.QueueEntry, error)
	GetItemOwner(ctx context.Context, itemID string) (string, error)
}
