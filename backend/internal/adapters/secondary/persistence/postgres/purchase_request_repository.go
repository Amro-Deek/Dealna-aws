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
		var txID *string
		if req.TransactionID.Valid {
			str := uuidToString(req.TransactionID)
			txID = &str
		}

		res[i] = domain.PurchaseRequest{
			RequestID:     uuidToString(req.RequestID),
			ItemID:        uuidToString(req.ItemID),
			BuyerID:       uuidToString(req.BuyerID),
			Status:        domain.PurchaseRequestStatus(req.Status),
			CreatedAt:     req.CreatedAt.Time,
			UpdatedAt:     req.UpdatedAt.Time,
			BuyerName:     req.BuyerName.String,
			BuyerPic:      req.BuyerPic.String,
			TransactionID: txID,
		}
	}
	return res, nil
}

func (r *PurchaseRequestRepository) UpdatePurchaseRequestStatusByItemAndBuyer(ctx context.Context, itemID string, buyerID string, status domain.PurchaseRequestStatus) error {
	return r.q.UpdatePurchaseRequestStatusByItemAndBuyer(ctx, generated.UpdatePurchaseRequestStatusByItemAndBuyerParams{
		ItemID:  toUUID(itemID),
		BuyerID: toUUID(buyerID),
		Status:  string(status),
	})
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

func (r *PurchaseRequestRepository) GetPurchaseRequestsByBuyer(ctx context.Context, buyerID string) ([]domain.PurchaseRequest, error) {
	reqs, err := r.q.GetPurchaseRequestsByBuyer(ctx, toUUID(buyerID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseRequest, len(reqs))
	for i, req := range reqs {
		price, _ := req.ItemPrice.Float64Value()
		var img string
		if s, ok := req.ItemImage.(string); ok {
			img = s
		}
		
		var txID *string
		if req.TransactionID.Valid {
			str := uuidToString(req.TransactionID)
			txID = &str
		}
		
		res[i] = domain.PurchaseRequest{
			RequestID:     uuidToString(req.RequestID),
			ItemID:        uuidToString(req.ItemID),
			BuyerID:       uuidToString(req.BuyerID),
			Status:        domain.PurchaseRequestStatus(req.Status),
			CreatedAt:     req.CreatedAt.Time,
			UpdatedAt:     req.UpdatedAt.Time,
			ItemTitle:     req.ItemTitle,
			ItemPrice:     price.Float64,
			ItemImage:     img,
			TransactionID: txID,
		}
	}
	return res, nil
}
