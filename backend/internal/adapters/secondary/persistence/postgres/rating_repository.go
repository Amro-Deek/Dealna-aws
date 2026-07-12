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
	rows, err := r.q.GetTransactionsToRemind(ctx, int32(days))
	if err != nil {
		return nil, err
	}

	var res []domain.PendingRating
	for _, row := range rows {
		txUUID, _ := uuid.Parse(uuidToString(row.TransactionID))
		buyerUUID, _ := uuid.Parse(uuidToString(row.BuyerID))
		sellerUUID, _ := uuid.Parse(uuidToString(row.SellerID))

		res = append(res, domain.PendingRating{
			TransactionID: txUUID,
			BuyerID:       buyerUUID,
			SellerID:      sellerUUID,
			ItemTitle:     row.ItemTitle,
		})
	}
	return res, nil
}
func (r *RatingRepository) CountRatingsBetweenUsers(ctx context.Context, user1, user2 uuid.UUID) (int, error) {
	count, err := r.q.CountRatingsBetweenUsers(ctx, generated.CountRatingsBetweenUsersParams{
		BuyerID:  toUUID(user1.String()),
		SellerID: toUUID(user2.String()),
	})
	return int(count), err
}
func (r *RatingRepository) UpdateUserRating(ctx context.Context, userID uuid.UUID, total int, sum int, bayesian float64) error {
	return r.q.UpdateUserRating(ctx, generated.UpdateUserRatingParams{
		UserID:         toUUID(userID.String()),
		TotalRatings:   int32(total),
		SumRatings:     int32(sum),
		BayesianRating: bayesian,
	})
}
func (r *RatingRepository) GetGlobalRatingAverage(ctx context.Context) (float64, int, error) {
	row, err := r.q.GetGlobalRatingAverage(ctx)
	if err != nil {
		return 4.0, 0, err
	}
	return row.GlobalAvg, int(row.TotalCount), nil
}
func (r *RatingRepository) UpdateSysConfig(ctx context.Context, key, value string) error {
	return nil
}
func (r *RatingRepository) GetSysConfig(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (r *RatingRepository) GetUserReviews(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Review, error) {
	rows, err := r.q.GetUserReviews(ctx, generated.GetUserReviewsParams{
		RatedUserID: toUUID(userID.String()),
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []domain.Review
	for _, row := range rows {
		ratingUUID, _ := uuid.Parse(uuidToString(row.RatingID))
		res = append(res, domain.Review{
			RatingID:  ratingUUID,
			Stars:     int(row.Stars),
			Comment:   row.Comment.String,
			CreatedAt: row.CreatedAt.Time,
			RaterName: row.RaterName.String,
		})
	}
	return res, nil
}
