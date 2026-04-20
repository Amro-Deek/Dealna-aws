package domain

import (
	"time"
)

type QueueEntryStatus string

const (
	QueueStatusWaiting   QueueEntryStatus = "WAITING"
	QueueStatusReserved  QueueEntryStatus = "RESERVED"
	QueueStatusExpired   QueueEntryStatus = "EXPIRED"
	QueueStatusCompleted QueueEntryStatus = "COMPLETED"
	QueueStatusCancelled QueueEntryStatus = "CANCELLED"
)

type QueueEntry struct {
	EntryID       string
	ItemID        string
	UserID        string
	JoinedAt      time.Time
	EntryStatus   QueueEntryStatus
	TurnStartedAt *time.Time
}

type QueuePosition struct {
	Entry    *QueueEntry
	Position int
	Total    int
}
