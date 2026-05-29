package domain

import (
	"encoding/json"
	"time"
)

type NotificationType string

const (
	NotifTypeTurnStarted       NotificationType = "TURN_STARTED"
	NotifTypeTurnAccepted      NotificationType = "TURN_ACCEPTED"
	NotifTypeHandoffInitiated  NotificationType = "HANDOFF_INITIATED"
	NotifTypeGiveawayCompleted NotificationType = "GIVEAWAY_COMPLETED"
	NotifTypeTurnExpired       NotificationType = "TURN_EXPIRED"
	NotifTypeGiveawayCancelled NotificationType = "GIVEAWAY_CANCELLED"
	NotifTypePurchaseRequested NotificationType = "PURCHASE_REQUESTED"
	NotifTypePurchaseAccepted  NotificationType = "PURCHASE_ACCEPTED"
	NotifTypePurchaseRejected  NotificationType = "PURCHASE_REJECTED"
	NotifTypeTransactionDone   NotificationType = "TRANSACTION_COMPLETED"
	NotifTypeApplicationApproved NotificationType = "APPLICATION_APPROVED"
	NotifTypeApplicationRejected NotificationType = "APPLICATION_REJECTED"
)

type Notification struct {
	NotificationID string
	UserID         string
	Type           NotificationType
	Payload        json.RawMessage  `json:"payload" swaggertype:"object"`
	IsRead         bool             `json:"is_read"`
	CreatedAt      time.Time
}
