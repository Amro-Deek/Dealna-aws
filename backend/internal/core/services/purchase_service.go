package services

import (
	"context"
	"errors"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/google/uuid"
)

type PurchaseService struct {
	repo     ports.IPurchaseRequestRepository
	notifs   *NotificationService
	itemRepo ports.ItemRepository
	txRepo   ports.ITransactionRepository
}

func NewPurchaseService(repo ports.IPurchaseRequestRepository, notifs *NotificationService, itemRepo ports.ItemRepository, txRepo ports.ITransactionRepository) *PurchaseService {
	return &PurchaseService{repo: repo, notifs: notifs, itemRepo: itemRepo, txRepo: txRepo}
}

func (s *PurchaseService) SendRequest(ctx context.Context, itemID, buyerID string) (*domain.PurchaseRequest, error) {
	parsedItemID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, err
	}
	item, err := s.itemRepo.GetItemDetail(ctx, parsedItemID)
	if err != nil {
		return nil, err
	}
	if item.Status != domain.ItemStatusAvailable {
		return nil, errors.New("item is no longer available for purchase")
	}
	if item.OwnerID.String() == buyerID {
		return nil, errors.New("you cannot purchase your own item")
	}
	
	// Check for existing requests to prevent spam
	existingReqs, err := s.repo.GetPurchaseRequestsByBuyer(ctx, buyerID)
	if err == nil {
		for _, req := range existingReqs {
			if req.ItemID == itemID && (req.Status == domain.PurchaseRequestPending || req.Status == domain.PurchaseRequestAccepted) {
				return nil, errors.New("you already have an active purchase request for this item")
			}
		}
	}

	req, err := s.repo.CreatePurchaseRequest(ctx, itemID, buyerID)
	if err == nil {
		sendPurchaseNotif(s, ctx, item.OwnerID.String(), itemID, req.RequestID, &buyerID, domain.NotifTypePurchaseRequested)
	}
	return req, err
}

func (s *PurchaseService) AcceptRequest(ctx context.Context, requestID, itemID, callerID string) (string, error) {
	// Verify Owner
	parsedItemID, err := uuid.Parse(itemID)
	if err != nil {
		return "", err
	}
	item, err := s.itemRepo.GetItemDetail(ctx, parsedItemID)
	if err != nil || item.OwnerID.String() != callerID {
		return "", errors.New("only the item owner can accept purchase requests")
	}

	req, err := s.repo.GetPurchaseRequestByID(ctx, requestID)
	if err != nil {
		return "", err
	}
	if req.Status != domain.PurchaseRequestPending {
		return "", errors.New("only pending requests can be accepted")
	}

	err = s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestPendingTx)
	if err != nil {
		return "", err
	}
	err = s.repo.FreezeOtherRequests(ctx, itemID, requestID)
	
	// Create Transaction
	tx, err := s.txRepo.CreateTransaction(ctx, itemID, req.BuyerID, callerID)
	if err != nil {
		return "", err
	}

	if err == nil {
		sendPurchaseNotif(s, ctx, req.BuyerID, itemID, requestID, &callerID, domain.NotifTypePurchaseAccepted)
	}
	return tx.TransactionID, err
}

func (s *PurchaseService) RejectRequest(ctx context.Context, requestID, itemID, callerID string) error {
	// Verify Owner
	parsedItemID, err := uuid.Parse(itemID)
	if err != nil {
		return err
	}
	item, err := s.itemRepo.GetItemDetail(ctx, parsedItemID)
	if err != nil || item.OwnerID.String() != callerID {
		return errors.New("only the item owner can reject purchase requests")
	}
	
	req, err := s.repo.GetPurchaseRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.PurchaseRequestPending {
		return errors.New("only pending requests can be rejected")
	}

	err = s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestRejected)
	if err == nil {
		sendPurchaseNotif(s, ctx, req.BuyerID, itemID, requestID, &callerID, domain.NotifTypePurchaseRejected)
	}
	return err
}

func (s *PurchaseService) CancelRequest(ctx context.Context, requestID, itemID, callerID string) error {
	req, err := s.repo.GetPurchaseRequestByID(ctx, requestID)
	if err != nil || req.BuyerID != callerID {
		return errors.New("only the buyer can cancel their purchase request")
	}
	if req.Status == domain.PurchaseRequestAccepted {
		return errors.New("request already accepted; you must cancel the transaction instead")
	}
	
	err = s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestCancelled)
	if err != nil {
		return err
	}
	// If the accepted request is cancelled, unfreeze others
	err = s.repo.UnfreezeRequests(ctx, itemID)
	if err == nil {
		// Fetch item owner
		parsedItemID, _ := uuid.Parse(itemID)
		item, _ := s.itemRepo.GetItemDetail(ctx, parsedItemID)
		if item != nil {
			sendPurchaseNotif(s, ctx, item.OwnerID.String(), itemID, requestID, &callerID, domain.NotifTypeGiveawayCancelled)
		}
	}
	return err
}

func (s *PurchaseService) ListRequests(ctx context.Context, itemID string) ([]domain.PurchaseRequest, error) {
	return s.repo.GetPurchaseRequestsByItem(ctx, itemID)
}

func (s *PurchaseService) GetMyRequests(ctx context.Context, buyerID string) ([]domain.PurchaseRequest, error) {
	return s.repo.GetPurchaseRequestsByBuyer(ctx, buyerID)
}

func sendPurchaseNotif(s *PurchaseService, ctx context.Context, userID, itemID, requestID string, actingUserID *string, typ domain.NotificationType) {
	if s.notifs == nil {
		return
	}
	notifCtx := NotificationContext{
		ItemID:       &itemID,
		EntryID:      &requestID,
		ActingUserID: actingUserID,
	}
	_ = s.notifs.CreateNotification(ctx, userID, typ, notifCtx)
}
