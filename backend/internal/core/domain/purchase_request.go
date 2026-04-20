package domain

import (
	"time"
)

type PurchaseRequestStatus string

const (
	PurchaseRequestPending   PurchaseRequestStatus = "PENDING"
	PurchaseRequestAccepted  PurchaseRequestStatus = "ACCEPTED"
	PurchaseRequestRejected  PurchaseRequestStatus = "REJECTED"
	PurchaseRequestFrozen    PurchaseRequestStatus = "FROZEN"
	PurchaseRequestCancelled PurchaseRequestStatus = "CANCELLED"
)

type PurchaseRequest struct {
	RequestID string
	ItemID    string
	BuyerID   string
	Status    PurchaseRequestStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
