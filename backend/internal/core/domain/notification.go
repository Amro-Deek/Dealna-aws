package domain

import (
	"encoding/json"
	"time"
)

type NotificationType string

const (
	NotifTypeTurnStarted       NotificationType = "TURN_STARTED"
	NotifTypeTurnExpired       NotificationType = "TURN_EXPIRED"
	NotifTypeGiveawayCancelled NotificationType = "GIVEAWAY_CANCELLED"
	NotifTypePurchaseRequested NotificationType = "PURCHASE_REQUESTED"
	NotifTypePurchaseAccepted  NotificationType = "PURCHASE_ACCEPTED"
	NotifTypePurchaseRejected  NotificationType = "PURCHASE_REJECTED"
	NotifTypeTransactionDone   NotificationType = "TRANSACTION_COMPLETED"
)

type Notification struct {
	NotificationID string
	UserID         string
	Type           NotificationType
	Payload        json.RawMessage
	IsRead         bool
	CreatedAt      time.Time
}
