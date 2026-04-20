package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type ITransactionRepository interface {
	CreateTransaction(ctx context.Context, itemID, buyerID, sellerID string) (*domain.Transaction, error)
	GetTransactionByID(ctx context.Context, transactionID string) (*domain.Transaction, error)
	GetTransactionByItem(ctx context.Context, itemID string) (*domain.Transaction, error)
	ConfirmSeller(ctx context.Context, transactionID string) error
	ConfirmBuyer(ctx context.Context, transactionID string) error
	CompleteTransaction(ctx context.Context, transactionID string) error
	CancelTransaction(ctx context.Context, transactionID string) error
}
