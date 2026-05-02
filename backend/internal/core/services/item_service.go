package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
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
	searchSync      ports.ISearchSyncPublisher
	searchRepo      ports.ISearchRepository
}

func NewItemService(repo ports.ItemRepository, storage ports.IStorageProvider, searchSync ports.ISearchSyncPublisher, searchRepo ports.ISearchRepository) *ItemService {
	return &ItemService{
		repo:            repo,
		storageProvider: storage,
		searchSync:      searchSync,
		searchRepo:      searchRepo,
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

	// 6. Push to SQS (Async, after DB commit so we don't have ghost data in Qdrant)
	univID, _ := s.repo.GetUniversityIDByUserID(ctx, cmd.OwnerID)
	
	var categoryStr string
	if item.CategoryID != nil {
		categoryStr = item.CategoryID.String()
	}

	sqsData := domain.SQSItemEventData{
		ItemID:      item.ID.String(),
		Title:       item.Title,
		Description: item.Description,
		Payload: domain.QdrantItemPayload{
			UniversityID: univID.String(),
			Category:     categoryStr,
			Price:        item.Price,
			Status:       string(item.Status),
			Condition:    "used", // Assume used for now until added to DB schema
			IsGiveaway:   item.Price == 0,
		},
	}

	event := domain.SearchSyncEvent{
		EventID:   uuid.New().String(),
		Action:    "create",
		Data:      sqsData,
		Timestamp: time.Now(),
	}
	_ = s.searchSync.PublishSyncEvent(context.Background(), event) // Fire and forget in background context

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
		err := s.repo.UpdateItemStatus(ctx, itemID, newStatus)
		if err == nil {
			// Emit status update to SQS
			event := domain.SearchSyncEvent{
				EventID: uuid.New().String(),
				Action:  "update_status",
				Data: domain.SQSItemEventData{
					ItemID: itemID.String(),
					Payload: domain.QdrantItemPayload{
						Status: string(newStatus),
					},
				},
				Timestamp: time.Now(),
			}
			_ = s.searchSync.PublishSyncEvent(context.Background(), event)
		}
		return err
	default:
		return ErrInvalidStatusTransition
	}
}

// SoftDeleteItem removes the item from the feed but keeps its data for audits/transactions.
func (s *ItemService) SoftDeleteItem(ctx context.Context, itemID uuid.UUID) error {
	err := s.repo.SoftDeleteItem(ctx, itemID)
	if err == nil {
		// Emit delete event to SQS so Qdrant drops it from vector search
		event := domain.SearchSyncEvent{
			EventID: uuid.New().String(),
			Action:  "delete",
			Data: domain.SQSItemEventData{
				ItemID: itemID.String(),
			},
			Timestamp: time.Now(),
		}
		_ = s.searchSync.PublishSyncEvent(context.Background(), event)
	}
	return err
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

// SearchItems coordinates semantic search by:
// 1. Vectorizing the query string via Lambda → Qdrant dense search
// 2. Postgres pg_trgm fuzzy keyword search (concurrent)
// 3. Merging results with Reciprocal Rank Fusion (RRF)
func (s *ItemService) SearchItems(ctx context.Context, query string, filter domain.ItemFilter) ([]domain.FeedItem, error) {
	if filter.ExcludedOwnerID != uuid.Nil {
		univID, err := s.repo.GetUniversityIDByUserID(ctx, filter.ExcludedOwnerID)
		if err == nil {
			filter.RequesterUniversityID = univID
		}
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return s.GetGlobalFeed(ctx, filter)
	}

	// --- Fire both searches CONCURRENTLY ---
	type denseResult struct {
		ids []uuid.UUID
		err error
	}
	type kwResult struct {
		ids []uuid.UUID
		err error
	}

	denseCh := make(chan denseResult, 1)
	kwCh := make(chan kwResult, 1)

	// Goroutine 1: Lambda → Qdrant dense vector search
	go func() {
		vector, err := s.searchSync.GenerateEmbedding(ctx, query)
		if err != nil {
			denseCh <- denseResult{err: fmt.Errorf("embedding failed: %w", err)}
			return
		}
		ids, err := s.searchRepo.SearchItems(ctx, vector, filter)
		denseCh <- denseResult{ids: ids, err: err}
	}()

	// Goroutine 2: Postgres pg_trgm keyword search
	go func() {
		ids, err := s.repo.KeywordSearchItems(ctx, query, filter)
		kwCh <- kwResult{ids: ids, err: err}
	}()

	denseRes := <-denseCh
	kwRes := <-kwCh

	// Log non-fatal errors but don't abort — partial results are better than none
	if denseRes.err != nil {
		log.Printf("[WARN] dense search error (non-fatal): %v", denseRes.err)
	}
	if kwRes.err != nil {
		log.Printf("[WARN] keyword search error (non-fatal): %v", kwRes.err)
	}

	// If both failed, surface the dense error
	if denseRes.err != nil && kwRes.err != nil {
		return nil, denseRes.err
	}

	// --- Reciprocal Rank Fusion (RRF) merge ---
	// score(id) = Σ 1/(k + rank_i) where k=60 (standard constant)
	const rrfK = 60.0
	scores := make(map[uuid.UUID]float64)

	for rank, id := range denseRes.ids {
		scores[id] += 1.0 / (rrfK + float64(rank+1))
	}
	for rank, id := range kwRes.ids {
		scores[id] += 1.0 / (rrfK + float64(rank+1))
	}

	// Collect unique IDs sorted by descending RRF score
	type scored struct {
		id    uuid.UUID
		score float64
	}
	ranked := make([]scored, 0, len(scores))
	for id, sc := range scores {
		ranked = append(ranked, scored{id, sc})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	limit := int(filter.Limit)
	if limit <= 0 {
		limit = 20
	}
	mergedIDs := make([]uuid.UUID, 0, min(len(ranked), limit))
	for i, r := range ranked {
		if i >= limit {
			break
		}
		mergedIDs = append(mergedIDs, r.id)
	}

	if len(mergedIDs) == 0 {
		return []domain.FeedItem{}, nil
	}

	// Hydrate from Postgres
	hydratedItems, err := s.repo.GetFeedItemsByIDs(ctx, mergedIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate hybrid search results: %w", err)
	}

	// Filter out requester's own items (safety net — Qdrant excludes, Postgres already excludes via SQL)
	items := make([]domain.FeedItem, 0, len(hydratedItems))
	for _, item := range hydratedItems {
		if filter.ExcludedOwnerID != uuid.Nil && item.OwnerID == filter.ExcludedOwnerID {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

