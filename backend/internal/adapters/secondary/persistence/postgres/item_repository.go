package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type ItemRepository struct {
	db      *pgxpool.Pool
	queries *generated.Queries
}

func NewItemRepository(db *pgxpool.Pool) ports.ItemRepository {
	return &ItemRepository{
		db:      db,
		queries: generated.New(db),
	}
}

// ExecuteInTx implements a transaction wrapper.
func (r *ItemRepository) ExecuteInTx(ctx context.Context, fn func(repo ports.ItemRepository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &ItemRepository{
		db:      r.db,
		queries: r.queries.WithTx(tx),
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ItemRepository) CreateItem(ctx context.Context, item *domain.Item) error {
	var categoryID pgtype.UUID
	if item.CategoryID != nil {
		categoryID = pgtype.UUID{Bytes: *item.CategoryID, Valid: true}
	} else {
		categoryID = pgtype.UUID{Valid: false}
	}

	var description pgtype.Text
	if item.Description != "" {
		description = pgtype.Text{String: item.Description, Valid: true}
	}

	var pickupLoc pgtype.Text
	if item.PickupLocation != "" {
		pickupLoc = pgtype.Text{String: item.PickupLocation, Valid: true}
	}

	// pgtype.Numeric.Scan only accepts string/[]byte in pgx v5 — float64 must be converted first.
	priceNumeric := pgtype.Numeric{}
	if err := priceNumeric.Scan(strconv.FormatFloat(item.Price, 'f', 10, 64)); err != nil {
		return fmt.Errorf("failed to encode price: %w", err)
	}

	params := generated.InsertItemParams{
		OwnerID:        pgtype.UUID{Bytes: item.OwnerID, Valid: true},
		CategoryID:     categoryID,
		Title:          item.Title,
		Description:    description,
		Price:          priceNumeric,
		PickupLocation: pickupLoc,
	}

	res, err := r.queries.InsertItem(ctx, params)
	if err != nil {
		return err
	}

	item.ID = res.ItemID.Bytes
	item.CreatedAt = res.CreatedAt.Time
	item.UpdatedAt = res.UpdatedAt.Time
	item.Status = domain.ItemStatus(res.ItemStatus)

	return nil
}

func (r *ItemRepository) AttachImages(ctx context.Context, itemID uuid.UUID, objectKeys []string) error {
	// We run these in a loop for simplicity, inside the Tx context if ExecuteInTx was called.
	for _, key := range objectKeys {
		params := generated.InsertAttachmentParams{
			ItemID:   pgtype.UUID{Bytes: itemID, Valid: true},
			FilePath: key,
		}
		if _, err := r.queries.InsertAttachment(ctx, params); err != nil {
			return err
		}
	}
	return nil
}

func (r *ItemRepository) UpdateItemStatus(ctx context.Context, itemID uuid.UUID, status domain.ItemStatus) error {
	params := generated.UpdateItemStatusParams{
		ItemID:     pgtype.UUID{Bytes: itemID, Valid: true},
		ItemStatus: string(status),
	}
	return r.queries.UpdateItemStatus(ctx, params)
}

func (r *ItemRepository) GetFeedItems(ctx context.Context, filter domain.ItemFilter) ([]domain.FeedItem, error) {
	var categoryID pgtype.UUID
	if filter.CategoryID != nil {
		categoryID = pgtype.UUID{Bytes: *filter.CategoryID, Valid: true}
	}

	var minPrice, maxPrice pgtype.Numeric
	if filter.MinPrice != nil {
		minPrice.Scan(strconv.FormatFloat(*filter.MinPrice, 'f', 10, 64)) //nolint:errcheck — NULL is safe fallback for optional filter
	}
	if filter.MaxPrice != nil {
		maxPrice.Scan(strconv.FormatFloat(*filter.MaxPrice, 'f', 10, 64)) //nolint:errcheck
	}

	var searchQuery string
	if filter.SearchQuery != nil {
		searchQuery = *filter.SearchQuery
	}

	params := generated.GetFeedItemsParams{
		UniversityID: pgtype.UUID{Bytes: filter.RequesterUniversityID, Valid: true},
		Column2:      categoryID,
		Column3:      minPrice,
		Column4:      maxPrice,
		Column5:      searchQuery,
		Offset:       filter.Offset,
		Limit:        filter.Limit,
	}

	rows, err := r.queries.GetFeedItems(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FeedItem, 0, len(rows))
	for _, row := range rows {
		price, _ := row.Price.Float64Value()
		
		var catID *uuid.UUID
		if row.CategoryID.Valid {
			id := uuid.UUID(row.CategoryID.Bytes)
			catID = &id
		}

		items = append(items, domain.FeedItem{
			ID:                 row.ItemID.Bytes,
			OwnerID:            row.OwnerID.Bytes,
			CategoryID:         catID,
			CategoryName:       row.CategoryName.String,
			Title:              row.Title,
			Description:        row.Description.String,
			Price:              price.Float64,
			PickupLocation:     row.PickupLocation.String,
			Status:             domain.ItemStatus(row.ItemStatus),
			CreatedAt:          row.CreatedAt.Time,
			OwnerDisplayName:   row.OwnerDisplayName.String,
			OwnerProfilePicURL: row.OwnerProfilePictureUrl.String,
			ThumbnailURL:       row.ThumbnailUrl.(string),
		})
	}
	return items, nil
}

func (r *ItemRepository) GetItemDetail(ctx context.Context, itemID uuid.UUID) (*domain.ItemDetail, error) {
	idPg := pgtype.UUID{Bytes: itemID, Valid: true}
	
	row, err := r.queries.GetItemDetails(ctx, idPg)
	if err != nil {
		return nil, err
	}

	price, _ := row.Price.Float64Value()
	var catID *uuid.UUID
	if row.CategoryID.Valid {
		id := uuid.UUID(row.CategoryID.Bytes)
		catID = &id
	}

	detail := &domain.ItemDetail{
		FeedItem: domain.FeedItem{
			ID:                 row.ItemID.Bytes,
			OwnerID:            row.OwnerID.Bytes,
			CategoryID:         catID,
			CategoryName:       row.CategoryName.String,
			Title:              row.Title,
			Description:        row.Description.String,
			Price:              price.Float64,
			PickupLocation:     row.PickupLocation.String,
			Status:             domain.ItemStatus(row.ItemStatus),
			CreatedAt:          row.CreatedAt.Time,
			OwnerDisplayName:   row.OwnerDisplayName.String,
			OwnerProfilePicURL: row.OwnerProfilePictureUrl.String,
		},
	}

	attachRows, err := r.queries.GetAttachmentsByItem(ctx, idPg)
	if err != nil {
		return detail, nil // return partial if attachments fail gracefully
	}

	detail.Attachments = make([]domain.Attachment, 0, len(attachRows))
	for _, a := range attachRows {
		detail.Attachments = append(detail.Attachments, domain.Attachment{
			ID:         a.AttachmentID.Bytes,
			ItemID:     a.ItemID.Bytes,
			FilePath:   a.FilePath,
			UploadedAt: a.UploadedAt.Time,
		})
	}

	if len(detail.Attachments) > 0 {
		detail.ThumbnailURL = detail.Attachments[0].FilePath // Assign primary attachment as thumbnail.
	}

	return detail, nil
}

func (r *ItemRepository) GetDailyItemCount(ctx context.Context, ownerID uuid.UUID) (int64, error) {
	return r.queries.GetDailyItemCount(ctx, pgtype.UUID{Bytes: ownerID, Valid: true})
}

func (r *ItemRepository) GetUserStorefront(ctx context.Context, ownerID uuid.UUID, limit, offset int32) ([]domain.FeedItem, error) {
	params := generated.GetUserStorefrontParams{
		OwnerID: pgtype.UUID{Bytes: ownerID, Valid: true},
		Limit:   limit,
		Offset:  offset,
	}

	rows, err := r.queries.GetUserStorefront(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FeedItem, 0, len(rows))
	for _, row := range rows {
		price, _ := row.Price.Float64Value()
		items = append(items, domain.FeedItem{
			ID:             row.ItemID.Bytes,
			Title:          row.Title,
			Price:          price.Float64,
			Status:         domain.ItemStatus(row.ItemStatus),
			CreatedAt:      row.CreatedAt.Time,
			ThumbnailURL:   row.ThumbnailUrl.(string),
		})
	}
	return items, nil
}

func (r *ItemRepository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.queries.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, domain.Category{
			ID:          row.CategoryID.Bytes,
			Name:        row.Name,
			Description: row.Description.String,
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return categories, nil
}

func (r *ItemRepository) SoftDeleteItem(ctx context.Context, itemID uuid.UUID) error {
	return r.queries.DeleteItem(ctx, pgtype.UUID{Bytes: itemID, Valid: true})
}

func (r *ItemRepository) GetUniversityIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	result, err := r.queries.GetUserUniversityID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return uuid.UUID{}, err
	}
	if !result.Valid {
		return uuid.UUID{}, domain.ErrUniversityNotFound
	}
	return uuid.UUID(result.Bytes), nil
}

