package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IPurchaseRequestRepository interface {
	CreatePurchaseRequest(ctx context.Context, itemID, buyerID string) (*domain.PurchaseRequest, error)
	GetPurchaseRequestsByItem(ctx context.Context, itemID string) ([]domain.PurchaseRequest, error)
	GetPurchaseRequestByID(ctx context.Context, requestID string) (*domain.PurchaseRequest, error)
	UpdatePurchaseRequestStatus(ctx context.Context, requestID string, status domain.PurchaseRequestStatus) error
	FreezeOtherRequests(ctx context.Context, itemID, excludeRequestID string) error
	UnfreezeRequests(ctx context.Context, itemID string) error
}
