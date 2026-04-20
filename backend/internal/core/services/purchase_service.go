package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type PurchaseService struct {
	repo   ports.IPurchaseRequestRepository
	notifs *NotificationService
}

func NewPurchaseService(repo ports.IPurchaseRequestRepository, notifs *NotificationService) *PurchaseService {
	return &PurchaseService{repo: repo, notifs: notifs}
}

func (s *PurchaseService) SendRequest(ctx context.Context, itemID, buyerID string) (*domain.PurchaseRequest, error) {
	return s.repo.CreatePurchaseRequest(ctx, itemID, buyerID)
}

func (s *PurchaseService) AcceptRequest(ctx context.Context, requestID, itemID string) error {
	err := s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestAccepted)
	if err != nil {
		return err
	}
	return s.repo.FreezeOtherRequests(ctx, itemID, requestID)
}

func (s *PurchaseService) RejectRequest(ctx context.Context, requestID string) error {
	return s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestRejected)
}

func (s *PurchaseService) CancelRequest(ctx context.Context, requestID, itemID string) error {
	err := s.repo.UpdatePurchaseRequestStatus(ctx, requestID, domain.PurchaseRequestCancelled)
	if err != nil {
		return err
	}
	// If the accepted request is cancelled, unfreeze others
	return s.repo.UnfreezeRequests(ctx, itemID)
}

func (s *PurchaseService) ListRequests(ctx context.Context, itemID string) ([]domain.PurchaseRequest, error) {
	return s.repo.GetPurchaseRequestsByItem(ctx, itemID)
}
