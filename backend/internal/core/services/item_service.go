package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

var (
	ErrDailyItemLimitReached = errors.New("you have reached the maximum allowed items (5) for today")
	ErrInvalidTitle          = errors.New("item title must be between 5 and 100 characters")
	ErrInvalidPrice          = errors.New("item price cannot be negative")
	ErrImagesRequired        = errors.New("at least one image must be attached to the item")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrContentTypeNotSupported = errors.New("content type not supported. Please use image/jpeg or image/png")
)

type ItemService struct {
	repo            ports.ItemRepository
	storageProvider ports.IStorageProvider
}

func NewItemService(repo ports.ItemRepository, storage ports.IStorageProvider) *ItemService {
	return &ItemService{
		repo:            repo,
		storageProvider: storage,
	}
}

// CreateItem performs business validation (e.g. daily limit, text constraints)
// and persists the item and its attachments.
func (s *ItemService) CreateItem(ctx context.Context, cmd domain.CreateItemCommand) (*domain.Item, error) {
	// 1. Validate Input constraints
	cmd.Title = strings.TrimSpace(cmd.Title)
	if len(cmd.Title) < 5 || len(cmd.Title) > 100 {
		return nil, ErrInvalidTitle
	}

	if cmd.Price < 0 {
		return nil, ErrInvalidPrice
	}

	if len(cmd.ObjectKeys) == 0 {
		return nil, ErrImagesRequired
	}
	if len(cmd.ObjectKeys) > 5 {
		return nil, errors.New("maximum of 5 images allowed per item")
	}

	// 2. Enforce Daily Listing Limit (NFR-1)
	count, err := s.repo.GetDailyItemCount(ctx, cmd.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily item limit: %w", err)
	}
	if count >= 5 {
		return nil, ErrDailyItemLimitReached
	}

	// 3. Create the Item Entity
	item := &domain.Item{
		OwnerID:        cmd.OwnerID,
		CategoryID:     cmd.CategoryID,
		Title:          cmd.Title,
		Description:    cmd.Description,
		Price:          cmd.Price,
		PickupLocation: cmd.PickupLocation,
		Status:         domain.ItemStatusAvailable,
	}

	// 4. Persist Item
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to persist item: %w", err)
	}

	// 5. Attach Images
	if err := s.repo.AttachImages(ctx, item.ID, cmd.ObjectKeys); err != nil {
		// Ideally wrap this in Tx, but doing it sequentially here is fine for MVP.
		return item, fmt.Errorf("failed to attach images: %w", err)
	}

	return item, nil
}

// GetGlobalFeed retrieves the chronological list of available items restricted to the university.
func (s *ItemService) GetGlobalFeed(ctx context.Context, filter domain.ItemFilter) ([]domain.FeedItem, error) {
	// Hard enforce bounds to prevent massive payloads.
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	items, err := s.repo.GetFeedItems(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed items: %w", err)
	}
	return items, nil
}

// GetItemDetail fetches the expanded details including all images.
func (s *ItemService) GetItemDetail(ctx context.Context, itemID uuid.UUID) (*domain.ItemDetail, error) {
	detail, err := s.repo.GetItemDetail(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item details: %w", err)
	}
	return detail, nil
}

// GetUserStorefront retrieves items owned by a specific user.
func (s *ItemService) GetUserStorefront(ctx context.Context, ownerID uuid.UUID, limit, offset int32) ([]domain.FeedItem, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetUserStorefront(ctx, ownerID, limit, offset)
}

// ListCategories statically fetches all dropdown categories.
func (s *ItemService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx)
}

// GetUserUniversityID resolves the university a user belongs to — used to enforce feed isolation.
func (s *ItemService) GetUserUniversityID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetUniversityIDByUserID(ctx, userID)
}

// UpdateStatus allows an owner to transition an item's status.
func (s *ItemService) UpdateStatus(ctx context.Context, itemID uuid.UUID, newStatus domain.ItemStatus) error {
	switch newStatus {
	case domain.ItemStatusAvailable, domain.ItemStatusReserved, domain.ItemStatusSold:
		return s.repo.UpdateItemStatus(ctx, itemID, newStatus)
	default:
		return ErrInvalidStatusTransition
	}
}

// SoftDeleteItem removes the item from the feed but keeps its data for audits/transactions.
func (s *ItemService) SoftDeleteItem(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.SoftDeleteItem(ctx, itemID)
}

// GenerateItemImageUploadURL returns a presigned S3 url bound to the items prefix.
func (s *ItemService) GenerateItemImageUploadURL(ctx context.Context, userID uuid.UUID, contentType string) (string, string, error) {
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		return "", "", ErrContentTypeNotSupported
	}

	objectKey := fmt.Sprintf("items/%s/%s.png", userID.String(), uuid.New().String())
	url, err := s.storageProvider.GeneratePresignedUploadURL(ctx, objectKey, contentType, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate S3 url: %w", err)
	}

	return url, objectKey, nil
}
