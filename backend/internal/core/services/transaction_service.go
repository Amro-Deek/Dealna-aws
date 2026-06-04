package services

import (
	"context"
	"errors"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/google/uuid"
)

type TransactionService struct {
	repo     ports.ITransactionRepository
	prRepo   ports.IPurchaseRequestRepository
	itemRepo ports.ItemRepository
	notifs   *NotificationService
}

func NewTransactionService(
	repo ports.ITransactionRepository,
	prRepo ports.IPurchaseRequestRepository,
	itemRepo ports.ItemRepository,
	notifs *NotificationService,
) *TransactionService {
	return &TransactionService{
		repo:     repo,
		prRepo:   prRepo,
		itemRepo: itemRepo,
		notifs:   notifs,
	}
}

func (s *TransactionService) StartTransaction(ctx context.Context, itemID, buyerID, sellerID string) (*domain.Transaction, error) {
	return s.repo.CreateTransaction(ctx, itemID, buyerID, sellerID)
}

func (s *TransactionService) ConfirmSeller(ctx context.Context, transactionID, callerID string) error {
	t, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if t.SellerID != callerID {
		return errors.New("only the seller can confirm seller handoff")
	}
	if t.Status != domain.TransactionPending {
		return errors.New("transaction is not pending")
	}

	err = s.repo.ConfirmSeller(ctx, transactionID)
	if err != nil {
		return err
	}
	// Check if both confirmed
	t, _ = s.repo.GetTransactionByID(ctx, transactionID)
	if t != nil && t.SellerConfirmed && t.BuyerConfirmed {
		s.repo.CompleteTransaction(ctx, transactionID)
		if s.prRepo != nil {
			s.prRepo.UpdatePurchaseRequestStatusByItemAndBuyer(ctx, t.ItemID, t.BuyerID, domain.PurchaseRequestCompleted)
		}
		if s.itemRepo != nil {
			itemUUID, err := uuid.Parse(t.ItemID)
			if err == nil {
				s.itemRepo.UpdateItemStatus(ctx, itemUUID, domain.ItemStatusSold)
			}
		}
		sendTxNotif(s, ctx, t.BuyerID, t.ItemID, transactionID, &callerID, domain.NotifTypeTransactionDone)
		sendTxNotif(s, ctx, t.SellerID, t.ItemID, transactionID, &callerID, domain.NotifTypeTransactionDone)
	}
	return nil
}

func (s *TransactionService) ConfirmBuyer(ctx context.Context, transactionID, callerID string) error {
	t, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if t.BuyerID != callerID {
		return errors.New("only the buyer can confirm buyer receipt")
	}
	if t.Status != domain.TransactionPending {
		return errors.New("transaction is not pending")
	}

	err = s.repo.ConfirmBuyer(ctx, transactionID)
	if err != nil {
		return err
	}
	// Verify Completion
	t, _ = s.repo.GetTransactionByID(ctx, transactionID)
	if t != nil && t.SellerConfirmed && t.BuyerConfirmed {
		s.repo.CompleteTransaction(ctx, transactionID)
		if s.prRepo != nil {
			s.prRepo.UpdatePurchaseRequestStatusByItemAndBuyer(ctx, t.ItemID, t.BuyerID, domain.PurchaseRequestCompleted)
		}
		if s.itemRepo != nil {
			itemUUID, err := uuid.Parse(t.ItemID)
			if err == nil {
				s.itemRepo.UpdateItemStatus(ctx, itemUUID, domain.ItemStatusSold)
			}
		}
		sendTxNotif(s, ctx, t.BuyerID, t.ItemID, transactionID, &callerID, domain.NotifTypeTransactionDone)
		sendTxNotif(s, ctx, t.SellerID, t.ItemID, transactionID, &callerID, domain.NotifTypeTransactionDone)
	}
	return nil
}

func sendTxNotif(s *TransactionService, ctx context.Context, userID, itemID, txID string, actingUserID *string, typ domain.NotificationType) {
	if s.notifs == nil {
		return
	}
	notifCtx := NotificationContext{
		ItemID:       &itemID,
		TxID:         &txID,
		ActingUserID: actingUserID,
	}
	_ = s.notifs.CreateNotification(ctx, userID, typ, notifCtx)
}

func (s *TransactionService) CancelTransaction(ctx context.Context, transactionID, callerID string) error {
	t, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if t.BuyerID != callerID && t.SellerID != callerID {
		return errors.New("only the buyer or seller can cancel the transaction")
	}
	if t.Status != domain.TransactionPending {
		return errors.New("cannot cancel a transaction that is not pending")
	}
	err = s.repo.CancelTransaction(ctx, transactionID)
	if err != nil {
		return err
	}

	if s.prRepo != nil {
		s.prRepo.UpdatePurchaseRequestStatusByItemAndBuyer(ctx, t.ItemID, t.BuyerID, domain.PurchaseRequestCancelled)
		s.prRepo.UnfreezeRequests(ctx, t.ItemID)
	}

	return nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, itemID string) (*domain.Transaction, error) {
	return s.repo.GetTransactionByItem(ctx, itemID)
}
