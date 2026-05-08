package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/google/uuid"
)

var (
	ErrQueueFull      = errors.New("queue is full")
	ErrAlreadyInQueue = errors.New("user already in active queue")
	ErrCooldown       = errors.New("user is in cooldown period after canceling or expiring")
)

type QueueService struct {
	repo     ports.IQueueRepository
	notifs   *NotificationService
	itemRepo ports.ItemRepository
}

func NewQueueService(repo ports.IQueueRepository, notifs *NotificationService, itemRepo ports.ItemRepository) *QueueService {
	return &QueueService{repo: repo, notifs: notifs, itemRepo: itemRepo}
}

func (s *QueueService) JoinQueue(ctx context.Context, itemID, userID string) (*domain.QueueEntry, error) {
	// Pre-flight checks for user-friendly errors
	activeCount, alreadyInQueue, inCooldown, err := s.repo.GetJoinEligibility(ctx, itemID, userID)
	if err != nil {
		return nil, err
	}
	if activeCount >= 10 {
		return nil, ErrQueueFull
	}
	if alreadyInQueue {
		return nil, ErrAlreadyInQueue
	}
	if inCooldown {
		return nil, ErrCooldown
	}

	// Atomic insert handles race conditions
	entry, err := s.repo.JoinQueueAtomic(ctx, itemID, userID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("failed to join queue due to concurrent limit or cooldown")
	}
	
	// If it's the first in queue, promote them immediately
	if activeCount == 0 {
		s.promoteNext(ctx, itemID)
		entry.EntryStatus = domain.QueueStatusReserved
	}

	return entry, nil
}

func (s *QueueService) LeaveQueue(ctx context.Context, itemID, userID string) error {
	// First get the active entry to see if we need to promote next
	entry, err := s.repo.GetActiveEntryByItemAndUser(ctx, itemID, userID)
	if err != nil || entry == nil {
		// If no active entry, just call leave
		return s.repo.LeaveQueue(ctx, itemID, userID)
	}
	
	err = s.repo.UpdateEntryStatus(ctx, entry.EntryID, domain.QueueStatusCancelled)
	if err != nil {
		return err
	}

	// If the user was RESERVED or CONFIRMED, promote the next person
	if entry.EntryStatus == domain.QueueStatusReserved || entry.EntryStatus == domain.QueueStatusConfirmed {
		s.promoteNext(ctx, itemID)
	}
	return nil
}

func (s *QueueService) GetPosition(ctx context.Context, itemID, entryID string) (int, error) {
	return s.repo.GetQueuePosition(ctx, itemID, entryID)
}

func (s *QueueService) GetQueueEntriesByUser(ctx context.Context, userID string) ([]domain.QueuePosition, error) {
	entries, err := s.repo.GetQueueEntriesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []domain.QueuePosition
	for i := range entries {
		entry := entries[i]
		pos, _ := s.repo.GetQueuePosition(ctx, entry.ItemID, entry.EntryID)
		total, _ := s.repo.CountQueueEntries(ctx, entry.ItemID)
		results = append(results, domain.QueuePosition{
			Entry:    &entry,
			Position: pos,
			Total:    total,
		})
	}
	return results, nil
}

func (s *QueueService) ExpireStaleEntries(ctx context.Context) error {
	// 1. Expire 1-hour RESERVED entries
	expiredReserved, err := s.repo.ExpireReservedEntries(ctx)
	if err == nil {
		for _, e := range expiredReserved {
			s.promoteNext(ctx, e.ItemID)
		}
	}

	// 2. Expire 7-day CONFIRMED entries
	expiredConfirmed, err := s.repo.ExpireConfirmedEntries(ctx)
	if err == nil {
		for _, e := range expiredConfirmed {
			s.promoteNext(ctx, e.ItemID)
		}
	}

	// 3. Auto-complete 24-hour HANDED_OFF entries
	autoCompleted, err := s.repo.AutoCompleteHandedOffEntries(ctx)
	if err == nil {
		for _, e := range autoCompleted {
			// Cancel all other entries since item is given away
			s.repo.CancelAllQueueEntries(ctx, e.ItemID)
		}
	}

	return nil
}

func (s *QueueService) promoteNext(ctx context.Context, itemID string) {
	next, _ := s.repo.GetFrontOfQueue(ctx, itemID)
	if next != nil {
		s.repo.SetTurnStarted(ctx, next.EntryID)
		sendQueueNotif(s, ctx, next.UserID, itemID, next.EntryID, domain.NotifTypeTurnStarted)
	}
}

func sendQueueNotif(s *QueueService, ctx context.Context, userID, itemID, entryID string, typ domain.NotificationType) {
	if s.notifs == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"item_id":  itemID,
		"entry_id": entryID,
	})
	_ = s.notifs.CreateNotification(ctx, userID, typ, payload)
}

func (s *QueueService) CancelGiveaway(ctx context.Context, itemID string) error {
	return s.repo.CancelAllQueueEntries(ctx, itemID)
}

func (s *QueueService) AcceptTurn(ctx context.Context, itemID, entryID, callerID string) error {
	ownerID, err := s.repo.GetItemOwner(ctx, itemID)
	if err != nil {
		return err
	}
	if ownerID != callerID {
		return errors.New("unauthorized: only item owner can accept turn")
	}

	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.EntryStatus != domain.QueueStatusReserved {
		return errors.New("entry is not in RESERVED state")
	}

	err = s.repo.UpdateEntryStatus(ctx, entryID, domain.QueueStatusConfirmed)
	if err == nil {
		sendQueueNotif(s, ctx, entry.UserID, itemID, entryID, domain.NotifTypeTurnAccepted)
	}
	return err
}

func (s *QueueService) RejectTurn(ctx context.Context, itemID, entryID, callerID string) error {
	ownerID, err := s.repo.GetItemOwner(ctx, itemID)
	if err != nil {
		return err
	}
	if ownerID != callerID {
		return errors.New("unauthorized: only item owner can reject turn")
	}

	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.EntryStatus != domain.QueueStatusReserved {
		return errors.New("entry is not in RESERVED state")
	}

	err = s.repo.UpdateEntryStatus(ctx, entryID, domain.QueueStatusExpired)
	if err != nil {
		return err
	}

	sendQueueNotif(s, ctx, entry.UserID, itemID, entryID, domain.NotifTypeTurnExpired)
	s.promoteNext(ctx, itemID)
	return nil
}

func (s *QueueService) InitiateHandoff(ctx context.Context, itemID, entryID, callerID string) error {
	ownerID, err := s.repo.GetItemOwner(ctx, itemID)
	if err != nil {
		return err
	}
	if ownerID != callerID {
		return errors.New("unauthorized: only item owner can initiate handoff")
	}

	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.EntryStatus != domain.QueueStatusConfirmed {
		return errors.New("entry is not in CONFIRMED state")
	}

	err = s.repo.UpdateEntryStatus(ctx, entryID, domain.QueueStatusHandedOff)
	if err == nil {
		sendQueueNotif(s, ctx, entry.UserID, itemID, entryID, domain.NotifTypeHandoffInitiated)
	}
	return err
}

func (s *QueueService) ConfirmHandoff(ctx context.Context, itemID, entryID, callerID string) error {
	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.UserID != callerID {
		return errors.New("unauthorized: only the receiver can confirm handoff")
	}
	if entry.EntryStatus != domain.QueueStatusHandedOff {
		return errors.New("entry is not in HANDED_OFF state")
	}

	err = s.repo.UpdateEntryStatus(ctx, entryID, domain.QueueStatusCompleted)
	if err != nil {
		return err
	}

	// Cancel all other entries
	s.repo.CancelAllQueueEntries(ctx, itemID)

	// Update item status to SOLD
	parsedItemID, _ := uuid.Parse(itemID)
	err = s.itemRepo.UpdateItemStatus(ctx, parsedItemID, domain.ItemStatusSold)
	
	if err == nil {
		// Notify owner that receiver confirmed it
		sendQueueNotif(s, ctx, callerID, itemID, entryID, domain.NotifTypeGiveawayCompleted)
	}
	return err
}
