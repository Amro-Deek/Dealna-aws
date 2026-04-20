package postgres

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseRequestRepository struct {
	q *generated.Queries
}

func NewPurchaseRequestRepository(conn *pgxpool.Pool) *PurchaseRequestRepository {
	return &PurchaseRequestRepository{
		q: generated.New(conn),
	}
}

func mapPurchaseRequest(r generated.PurchaseRequest) *domain.PurchaseRequest {
	return &domain.PurchaseRequest{
		RequestID: uuidToString(r.RequestID),
		ItemID:    uuidToString(r.ItemID),
		BuyerID:   uuidToString(r.BuyerID),
		Status:    domain.PurchaseRequestStatus(r.Status),
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func (r *PurchaseRequestRepository) CreatePurchaseRequest(ctx context.Context, itemID, buyerID string) (*domain.PurchaseRequest, error) {
	req, err := r.q.CreatePurchaseRequest(ctx, generated.CreatePurchaseRequestParams{
		ItemID:  toUUID(itemID),
		BuyerID: toUUID(buyerID),
	})
	if err != nil {
		return nil, err
	}
	return mapPurchaseRequest(req), nil
}

func (r *PurchaseRequestRepository) GetPurchaseRequestsByItem(ctx context.Context, itemID string) ([]domain.PurchaseRequest, error) {
	reqs, err := r.q.GetPurchaseRequestsByItem(ctx, toUUID(itemID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseRequest, len(reqs))
	for i, req := range reqs {
		res[i] = *mapPurchaseRequest(req)
	}
	return res, nil
}

func (r *PurchaseRequestRepository) GetPurchaseRequestByID(ctx context.Context, requestID string) (*domain.PurchaseRequest, error) {
	req, err := r.q.GetPurchaseRequestByID(ctx, toUUID(requestID))
	if err != nil {
		return nil, err
	}
	return mapPurchaseRequest(req), nil
}

func (r *PurchaseRequestRepository) UpdatePurchaseRequestStatus(ctx context.Context, requestID string, status domain.PurchaseRequestStatus) error {
	return r.q.UpdatePurchaseRequestStatus(ctx, generated.UpdatePurchaseRequestStatusParams{
		RequestID: toUUID(requestID),
		Status:    string(status),
	})
}

func (r *PurchaseRequestRepository) FreezeOtherRequests(ctx context.Context, itemID, excludeRequestID string) error {
	return r.q.FreezeOtherRequests(ctx, generated.FreezeOtherRequestsParams{
		ItemID:    parseUUID(itemID),
		RequestID: parseUUID(excludeRequestID),
	})
}

func (r *PurchaseRequestRepository) UnfreezeRequests(ctx context.Context, itemID string) error {
	return r.q.UnfreezeRequests(ctx, parseUUID(itemID))
}
