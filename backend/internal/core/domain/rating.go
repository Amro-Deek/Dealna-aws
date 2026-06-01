package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRatingWindowExpired = errors.New("the 14-day window to rate this transaction has expired")
	ErrRatingNotAllowed    = errors.New("rating is not allowed for this transaction type or status")
	ErrAlreadyRated        = errors.New("you have already rated this transaction")
)

type Rating struct {
	RatingID      uuid.UUID `json:"rating_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	RaterID       uuid.UUID `json:"rater_id"`
	RatedUserID   uuid.UUID `json:"rated_user_id"`
	Stars         int       `json:"stars"`
	Comment       string    `json:"comment"`
	IsFrozen      bool      `json:"is_frozen"`
	CreatedAt     time.Time `json:"created_at"`
}

type PendingRating struct {
	TransactionID       uuid.UUID `json:"transaction_id"`
	ItemID              uuid.UUID `json:"item_id"`
	ItemTitle           string    `json:"item_title"`
	SellerID            uuid.UUID `json:"seller_id"`
	SellerName          string    `json:"seller_name"`
	DaysSinceCompletion int       `json:"days_since_completion"`
}

type CreateRatingCommand struct {
	RaterID       uuid.UUID `json:"rater_id"`
	RatedUserID   uuid.UUID `json:"rated_user_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	Stars         int       `json:"stars"`
	Comment       string    `json:"comment"`
}
