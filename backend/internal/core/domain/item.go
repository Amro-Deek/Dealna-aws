package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrUniversityNotFound is returned when a user has no university_id linked.
var ErrUniversityNotFound = errors.New("user has no linked university")

// ItemStatus defines the current state of an item listing.
type ItemStatus string

const (
	ItemStatusAvailable ItemStatus = "AVAILABLE"
	ItemStatusReserved  ItemStatus = "RESERVED"
	ItemStatusSold      ItemStatus = "SOLD"
)

// Category represents a product/service classification.
type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Item represents the core listing entity.
type Item struct {
	ID             uuid.UUID  `json:"id"`
	OwnerID        uuid.UUID  `json:"owner_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Price          float64    `json:"price"`
	PickupLocation string     `json:"pickup_location"`
	Status         ItemStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Attachment represents a linked S3 object (e.g., image).
type Attachment struct {
	ID         uuid.UUID `json:"id"`
	ItemID     uuid.UUID `json:"item_id"`
	FilePath   string    `json:"file_path"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// FeedItem is the specialized aggregate struct returned by the main chronological Feed.
// It bundles the generic Item data alongside the owner's minimal profile and a thumbnail.
type FeedItem struct {
	ID             uuid.UUID  `json:"id"`
	OwnerID        uuid.UUID  `json:"owner_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	CategoryName   string     `json:"category_name"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Price          float64    `json:"price"`
	PickupLocation string     `json:"pickup_location"`
	Status         ItemStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`

	OwnerDisplayName   string `json:"owner_display_name"`
	OwnerProfilePicURL string `json:"owner_profile_pic_url"`
	ThumbnailURL       string `json:"thumbnail_url"`
}

// ItemDetail extends FeedItem to include an array of all full-resolution attachments.
type ItemDetail struct {
	FeedItem
	Attachments []Attachment `json:"attachments"`
}

// CreateItemCommand wraps the data required from a user payload to create a new ad.
type CreateItemCommand struct {
	OwnerID        uuid.UUID  `json:"owner_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Price          float64    `json:"price"`
	PickupLocation string     `json:"pickup_location"`
	ObjectKeys     []string   `json:"object_keys"` // S3 keys pointing to uploaded images
}

// ItemFilter defines search parameters for the Marketplace Feed.
type ItemFilter struct {
	RequesterUniversityID uuid.UUID // Crucial: Ensures users only browse within their own university (Birzeit MVP).
	ExcludedOwnerID       uuid.UUID // New: To hide logged-in user's items from their own feed.
	CategoryID            *uuid.UUID
	MinPrice              *float64
	MaxPrice              *float64
	SearchQuery           *string
	Limit                 int32
	Offset                int32
}
