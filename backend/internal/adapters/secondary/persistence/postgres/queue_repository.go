package postgres

import (
	"context"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
)

type QueueRepository struct {
	q *generated.Queries
}

func NewQueueRepository(conn *pgxpool.Pool) *QueueRepository {
	return &QueueRepository{
		q: generated.New(conn),
	}
}

func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}

func mapQueueEntry(entry generated.QueueEntry) *domain.QueueEntry {
	var turnStartedAt *time.Time
	if entry.TurnStartedAt.Valid {
		turnStartedAt = &entry.TurnStartedAt.Time
	}
	return &domain.QueueEntry{
		EntryID:       uuidToString(entry.EntryID),
		ItemID:        uuidToString(entry.ItemID),
		UserID:        uuidToString(entry.UserID),
		JoinedAt:      entry.JoinedAt.Time,
		EntryStatus:   domain.QueueEntryStatus(entry.EntryStatus),
		TurnStartedAt: turnStartedAt,
		UpdatedAt:     entry.UpdatedAt.Time,
	}
}

func (r *QueueRepository) JoinQueueAtomic(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error) {
	entry, err := r.q.JoinQueueAtomic(ctx, generated.JoinQueueAtomicParams{
		ItemID: toUUID(itemID),
		UserID: toUUID(userID),
	})
	if err != nil {
		return nil, err
	}
	return mapQueueEntry(entry), nil
}

func (r *QueueRepository) GetJoinEligibility(ctx context.Context, itemID, userID string) (int, bool, bool, error) {
	row, err := r.q.GetJoinEligibility(ctx, generated.GetJoinEligibilityParams{
		ItemID: toUUID(itemID),
		UserID: toUUID(userID),
	})
	if err != nil {
		return 0, false, false, err
	}
	return int(row.ActiveCount), row.AlreadyInQueue, row.InCooldown, nil
}

func (r *QueueRepository) GetQueueByItem(ctx context.Context, itemID string) ([]domain.QueueEntry, error) {
	entries, err := r.q.GetQueueByItemID(ctx, toUUID(itemID))
	if err != nil {
		return nil, err
	}
	var res []domain.QueueEntry
	for _, e := range entries {
		res = append(res, *mapQueueEntry(e))
	}
	return res, nil
}

func (r *QueueRepository) GetQueuePosition(ctx context.Context, itemID, entryID string) (int, error) {
	pos, err := r.q.GetQueuePosition(ctx, generated.GetQueuePositionParams{
		ItemID:  toUUID(itemID),
		EntryID: toUUID(entryID),
	})
	return int(pos), err
}

func (r *QueueRepository) GetFrontOfQueue(ctx context.Context, itemID string) (*domain.QueueEntry, error) {
	entry, err := r.q.GetFrontOfQueue(ctx, toUUID(itemID))
	if err != nil {
		return nil, err
	}
	return mapQueueEntry(entry), nil
}

func (r *QueueRepository) GetActiveEntryByItemAndUser(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error) {
	entry, err := r.q.GetActiveEntryByItemAndUser(ctx, generated.GetActiveEntryByItemAndUserParams{
		ItemID: toUUID(itemID),
		UserID: toUUID(userID),
	})
	if err != nil {
		return nil, err
	}
	return mapQueueEntry(entry), nil
}

func (r *QueueRepository) UpdateEntryStatus(ctx context.Context, entryID string, status domain.QueueEntryStatus) error {
	return r.q.UpdateEntryStatus(ctx, generated.UpdateEntryStatusParams{
		EntryID:     toUUID(entryID),
		EntryStatus: string(status),
	})
}

func (r *QueueRepository) SetTurnStarted(ctx context.Context, entryID string) error {
	return r.q.SetTurnStarted(ctx, toUUID(entryID))
}

func (r *QueueRepository) ExpireReservedEntries(ctx context.Context) ([]domain.QueueEntry, error) {
	entries, err := r.q.ExpireReservedEntries(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.QueueEntry
	for _, e := range entries {
		res = append(res, *mapQueueEntry(e))
	}
	return res, nil
}

func (r *QueueRepository) ExpireConfirmedEntries(ctx context.Context) ([]domain.QueueEntry, error) {
	entries, err := r.q.ExpireConfirmedEntries(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.QueueEntry
	for _, e := range entries {
		res = append(res, *mapQueueEntry(e))
	}
	return res, nil
}

func (r *QueueRepository) AutoCompleteHandedOffEntries(ctx context.Context) ([]domain.QueueEntry, error) {
	entries, err := r.q.AutoCompleteHandedOffEntries(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.QueueEntry
	for _, e := range entries {
		res = append(res, *mapQueueEntry(e))
	}
	return res, nil
}

func (r *QueueRepository) CancelAllQueueEntries(ctx context.Context, itemID string) error {
	return r.q.CancelAllQueueEntries(ctx, toUUID(itemID))
}

func (r *QueueRepository) RemoveFromQueue(ctx context.Context, entryID string) error {
	return r.q.RemoveFromQueue(ctx, toUUID(entryID))
}

func (r *QueueRepository) GetQueueEntriesByUser(ctx context.Context, userID string) ([]domain.QueueEntry, error) {
	entries, err := r.q.GetQueueEntriesByUser(ctx, toUUID(userID))
	if err != nil {
		return nil, err
	}
	var res []domain.QueueEntry
	for _, e := range entries {
		res = append(res, *mapQueueEntry(e))
	}
	return res, nil
}

func (r *QueueRepository) CountQueueEntries(ctx context.Context, itemID string) (int, error) {
	count, err := r.q.CountQueueEntries(ctx, toUUID(itemID))
	return int(count), err
}

func (r *QueueRepository) LeaveQueue(ctx context.Context, itemID, userID string) error {
	return r.q.LeaveQueue(ctx, generated.LeaveQueueParams{
		ItemID: toUUID(itemID),
		UserID: toUUID(userID),
	})
}

func (r *QueueRepository) GetEntryByID(ctx context.Context, entryID string) (*domain.QueueEntry, error) {
	entry, err := r.q.GetEntryByID(ctx, toUUID(entryID))
	if err != nil {
		return nil, err
	}
	return mapQueueEntry(entry), nil
}

func (r *QueueRepository) GetItemOwner(ctx context.Context, itemID string) (string, error) {
	return r.q.GetItemOwner(ctx, toUUID(itemID))
}
