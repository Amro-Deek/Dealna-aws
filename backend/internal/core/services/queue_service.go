package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type QueueService struct {
	repo   ports.IQueueRepository
	notifs *NotificationService
}

func NewQueueService(repo ports.IQueueRepository, notifs *NotificationService) *QueueService {
	return &QueueService{repo: repo, notifs: notifs}
}

func (s *QueueService) JoinQueue(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error) {
	return s.repo.JoinQueue(ctx, itemID, userID)
}

func (s *QueueService) LeaveQueue(ctx context.Context, itemID, userID string) error {
	return s.repo.LeaveQueue(ctx, itemID, userID)
}

func (s *QueueService) GetPosition(ctx context.Context, itemID, entryID string) (int, error) {
	return s.repo.GetQueuePosition(ctx, itemID, entryID)
}

func (s *QueueService) ExpireStaleEntries(ctx context.Context) error {
	entries, err := s.repo.GetExpiredTurns(ctx)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s.repo.UpdateEntryStatus(ctx, e.EntryID, domain.QueueStatusExpired)
		// Auto promote next person:
		s.promoteNext(ctx, e.ItemID)
	}
	return nil
}

func (s *QueueService) promoteNext(ctx context.Context, itemID string) {
	next, _ := s.repo.GetFrontOfQueue(ctx, itemID)
	if next != nil {
		s.repo.SetTurnStarted(ctx, next.EntryID)
	}
}

func (s *QueueService) CancelGiveaway(ctx context.Context, itemID string) error {
	// Cancel logic
	return nil
}
