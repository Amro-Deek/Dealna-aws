package postgres

import (
	"context"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RatingRepository struct {
	q *generated.Queries
}

func NewRatingRepository(q *generated.Queries) *RatingRepository {
	return &RatingRepository{q: q}
}

func (r *RatingRepository) CreateRating(ctx context.Context, cmd domain.CreateRatingCommand, isFrozen bool) (domain.Rating, error) {
	row, err := r.q.CreateRating(ctx, generated.CreateRatingParams{
		RaterID:       toUUID(cmd.RaterID.String()),
		RatedUserID:   toUUID(cmd.RatedUserID.String()),
		TransactionID: toUUID(cmd.TransactionID.String()),
		Stars:         int32(cmd.Stars),
		Comment:       pgtype.Text{String: cmd.Comment, Valid: cmd.Comment != ""},
		IsFrozen:      isFrozen,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			return domain.Rating{}, domain.ErrAlreadyRated
		}
		return domain.Rating{}, err
	}
	
	raterUUID, _ := uuid.Parse(uuidToString(row.RaterID))
	targetUUID, _ := uuid.Parse(uuidToString(row.RatedUserID))
	txUUID, _ := uuid.Parse(uuidToString(row.TransactionID))
	ratingUUID, _ := uuid.Parse(uuidToString(row.RatingID))

	return domain.Rating{
		RatingID:      ratingUUID,
		TransactionID: txUUID,
		RaterID:       raterUUID,
		RatedUserID:   targetUUID,
		Stars:         int(row.Stars),
		Comment:       row.Comment.String,
		IsFrozen:      row.IsFrozen,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (r *RatingRepository) GetPendingRatings(ctx context.Context, buyerID uuid.UUID) ([]domain.PendingRating, error) {
	rows, err := r.q.GetPendingRatings(ctx, toUUID(buyerID.String()))
	if err != nil {
		return nil, err
	}
	
	var res []domain.PendingRating
	for _, row := range rows {
		txUUID, _ := uuid.Parse(uuidToString(row.TransactionID))
		itemUUID, _ := uuid.Parse(uuidToString(row.ItemID))
		sellerUUID, _ := uuid.Parse(uuidToString(row.SellerID))

		res = append(res, domain.PendingRating{
			TransactionID:       txUUID,
			ItemID:              itemUUID,
			ItemTitle:           row.ItemTitle,
			SellerID:            sellerUUID,
			SellerName:          row.SellerName.String,
			DaysSinceCompletion: int(row.DaysSinceCompletion),
		})
	}
	return res, nil
}

func (r *RatingRepository) GetTransactionsToRemind(ctx context.Context, days int) ([]domain.PendingRating, error) {
	return nil, nil // Stub for now
}
func (r *RatingRepository) CountRatingsBetweenUsers(ctx context.Context, user1, user2 uuid.UUID) (int, error) {
	count, err := r.q.CountRatingsBetweenUsers(ctx, generated.CountRatingsBetweenUsersParams{
		BuyerID:  toUUID(user1.String()),
		SellerID: toUUID(user2.String()),
	})
	return int(count), err
}
func (r *RatingRepository) UpdateUserRating(ctx context.Context, userID uuid.UUID, total int, sum int, bayesian float64) error {
	// Stub for now, update queries to use correctly named parameters.
	return nil
}
func (r *RatingRepository) GetGlobalRatingAverage(ctx context.Context) (float64, int, error) {
	return 4.0, 0, nil // Stub for now
}
func (r *RatingRepository) UpdateSysConfig(ctx context.Context, key, value string) error {
	return nil
}
func (r *RatingRepository) GetSysConfig(ctx context.Context, key string) (string, error) {
	return "", nil
}
