package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

// ItemRepository defines data-access operations for the Marketplace and Feed.
type ItemRepository interface {
	// ExecuteInTx wraps operations in a transaction. (Useful if needed later).
	ExecuteInTx(ctx context.Context, fn func(repo ItemRepository) error) error

	// CreateItem persists a new item to the database.
	CreateItem(ctx context.Context, item *domain.Item) error

	// AttachImages links S3 object keys to an existing item_id.
	AttachImages(ctx context.Context, itemID uuid.UUID, objectKeys []string) error

	// UpdateItemStatus modifies the physical status of the item.
	UpdateItemStatus(ctx context.Context, itemID uuid.UUID, status domain.ItemStatus) error

	// GetFeedItems handles the primary Cold Start Feed, strictly constrained by university.
	GetFeedItems(ctx context.Context, filter domain.ItemFilter) ([]domain.FeedItem, error)

	// GetUserStorefront fetches all items owned by a specific user.
	GetUserStorefront(ctx context.Context, ownerID uuid.UUID, limit, offset int32) ([]domain.FeedItem, error)

	// GetItemDetail fetches the expanded item details (including attachments).
	GetItemDetail(ctx context.Context, itemID uuid.UUID) (*domain.ItemDetail, error)

	// GetDailyItemCount checks how many active items the user has posted in the last 24H.
	GetDailyItemCount(ctx context.Context, ownerID uuid.UUID) (int64, error)

	// ListCategories retrieves the static dropdown payload.
	ListCategories(ctx context.Context) ([]domain.Category, error)

	// SoftDeleteItem sets the deleted_at flag, retaining relations.
	SoftDeleteItem(ctx context.Context, itemID uuid.UUID) error

	// GetUniversityIDByUserID fetches the university a user belongs to.
	// Used for automatic university scoping on the feed without needing a header.
	GetUniversityIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}
