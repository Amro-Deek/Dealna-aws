package domain

import (
	"time"
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "PENDING"
	TransactionCompleted TransactionStatus = "COMPLETED"
	TransactionCancelled TransactionStatus = "CANCELLED"
)

type Transaction struct {
	TransactionID   string
	ItemID          string
	BuyerID         string
	SellerID        string
	Status          TransactionStatus
	SellerConfirmed bool
	BuyerConfirmed  bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
