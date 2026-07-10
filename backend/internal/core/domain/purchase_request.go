package domain

import (
	"time"
)

type PurchaseRequestStatus string

const (
	PurchaseRequestPending   PurchaseRequestStatus = "PENDING"
	PurchaseRequestAccepted  PurchaseRequestStatus = "ACCEPTED"
	PurchaseRequestPendingTx PurchaseRequestStatus = "PENDING_TX"
	PurchaseRequestRejected  PurchaseRequestStatus = "REJECTED"
	PurchaseRequestFrozen    PurchaseRequestStatus = "FROZEN"
	PurchaseRequestCancelled PurchaseRequestStatus = "CANCELLED"
	PurchaseRequestCompleted PurchaseRequestStatus = "COMPLETED"
)

type PurchaseRequest struct {
	RequestID string
	ItemID    string
	BuyerID   string
	Status    PurchaseRequestStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	BuyerName string
	BuyerPic  string

	// Hydrated fields
	ItemTitle     string
	ItemPrice     float64
	ItemImage     string
	TransactionID *string `json:"transactionId,omitempty"`
}
