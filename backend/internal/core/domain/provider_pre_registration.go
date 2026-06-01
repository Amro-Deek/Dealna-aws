package domain

import "time"

type ProviderPreRegistration struct {
	ID                string
	Email             string
	Token             string
	ExpiresAt         time.Time
	UsedAt            *time.Time
	ResendCount       int
	ResendWindowStart *time.Time
	VerifiedAt        *time.Time
}
