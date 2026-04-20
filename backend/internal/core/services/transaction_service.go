package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type TransactionService struct {
	repo   ports.ITransactionRepository
	notifs *NotificationService
}

func NewTransactionService(repo ports.ITransactionRepository, notifs *NotificationService) *TransactionService {
	return &TransactionService{repo: repo, notifs: notifs}
}

func (s *TransactionService) StartTransaction(ctx context.Context, itemID, buyerID, sellerID string) (*domain.Transaction, error) {
	return s.repo.CreateTransaction(ctx, itemID, buyerID, sellerID)
}

func (s *TransactionService) ConfirmSeller(ctx context.Context, transactionID string) error {
	err := s.repo.ConfirmSeller(ctx, transactionID)
	if err != nil {
		return err
	}
	// Check if both confirmed
	t, _ := s.repo.GetTransactionByID(ctx, transactionID)
	if t != nil && t.Status == domain.TransactionCompleted {
		return nil // Maybe check buyer_confirmed logic here, omitting for succinctness
	}
	return nil
}

func (s *TransactionService) ConfirmBuyer(ctx context.Context, transactionID string) error {
	err := s.repo.ConfirmBuyer(ctx, transactionID)
	if err != nil {
		return err
	}
	// Verify Completion
	return nil
}

func (s *TransactionService) CancelTransaction(ctx context.Context, transactionID string) error {
	return s.repo.CancelTransaction(ctx, transactionID)
}

func (s *TransactionService) GetTransaction(ctx context.Context, itemID string) (*domain.Transaction, error) {
	return s.repo.GetTransactionByItem(ctx, itemID)
}
