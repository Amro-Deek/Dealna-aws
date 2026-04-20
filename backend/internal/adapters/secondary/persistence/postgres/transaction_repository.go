package postgres

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	q *generated.Queries
}

func NewTransactionRepository(conn *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		q: generated.New(conn),
	}
}

func mapTransaction(t generated.Transaction) *domain.Transaction {
	return &domain.Transaction{
		TransactionID: uuidToString(t.TransactionID),
		ItemID:        uuidToString(t.ItemID),
		BuyerID:       uuidToString(t.BuyerID),
		SellerID:      uuidToString(t.SellerID),
		Status:        domain.TransactionStatus(t.TransactionStatus),
		CreatedAt:     t.CreatedAt.Time,
	}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, itemID, buyerID, sellerID string) (*domain.Transaction, error) {
	t, err := r.q.CreateTransaction(ctx, generated.CreateTransactionParams{
		ItemID:   toUUID(itemID),
		BuyerID:  toUUID(buyerID),
		SellerID: toUUID(sellerID),
	})
	if err != nil {
		return nil, err
	}
	return mapTransaction(t), nil
}

func (r *TransactionRepository) GetTransactionByID(ctx context.Context, transactionID string) (*domain.Transaction, error) {
	t, err := r.q.GetTransactionByID(ctx, toUUID(transactionID))
	if err != nil {
		return nil, err
	}
	return mapTransaction(t), nil
}

func (r *TransactionRepository) GetTransactionByItem(ctx context.Context, itemID string) (*domain.Transaction, error) {
	t, err := r.q.GetTransactionByItem(ctx, toUUID(itemID))
	if err != nil {
		return nil, err
	}
	return mapTransaction(t), nil
}

func (r *TransactionRepository) ConfirmSeller(ctx context.Context, transactionID string) error {
	return r.q.ConfirmSeller(ctx, toUUID(transactionID))
}

func (r *TransactionRepository) ConfirmBuyer(ctx context.Context, transactionID string) error {
	return r.q.ConfirmBuyer(ctx, toUUID(transactionID))
}

func (r *TransactionRepository) CompleteTransaction(ctx context.Context, transactionID string) error {
	return r.q.CompleteTransaction(ctx, toUUID(transactionID))
}

func (r *TransactionRepository) CancelTransaction(ctx context.Context, transactionID string) error {
	return r.q.CancelTransaction(ctx, toUUID(transactionID))
}
