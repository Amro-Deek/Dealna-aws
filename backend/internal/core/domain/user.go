package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             string
	Email          string
	Role           string
	KeycloakSub    string
	UniversityID   uuid.UUID
	TotalRatings   int
	SumRatings     int
	BayesianRating float64
}

type PasswordResetToken struct {
	Email     string
	Token     string
	ExpiresAt time.Time
}
