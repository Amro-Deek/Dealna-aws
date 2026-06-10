package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IRatingRepository interface {
	CreateRating(ctx context.Context, cmd domain.CreateRatingCommand, isFrozen bool) (domain.Rating, error)
	GetPendingRatings(ctx context.Context, buyerID uuid.UUID) ([]domain.PendingRating, error)
	GetTransactionsToRemind(ctx context.Context, days int) ([]domain.PendingRating, error)
	CountRatingsBetweenUsers(ctx context.Context, user1, user2 uuid.UUID) (int, error)
	UpdateUserRating(ctx context.Context, userID uuid.UUID, total int, sum int, bayesian float64) error
	GetGlobalRatingAverage(ctx context.Context) (globalAvg float64, totalCount int, err error)
	UpdateSysConfig(ctx context.Context, key, value string) error
	GetSysConfig(ctx context.Context, key string) (string, error)
	GetUserReviews(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Review, error)
}

type ITransactionRepository interface {
	GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error)
}

type IUserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

type RatingService struct {
	ratingRepo IRatingRepository
	txRepo     ITransactionRepository
	userRepo   IUserRepository
	notifs     *NotificationService
}

func NewRatingService(ratingRepo IRatingRepository, txRepo ITransactionRepository, userRepo IUserRepository, notifs *NotificationService) *RatingService {
	return &RatingService{
		ratingRepo: ratingRepo,
		txRepo:     txRepo,
		userRepo:   userRepo,
		notifs:     notifs,
	}
}

func (s *RatingService) CreateRating(ctx context.Context, cmd domain.CreateRatingCommand) (domain.Rating, error) {
	tx, err := s.txRepo.GetTransactionByID(ctx, cmd.TransactionID.String())
	if err != nil {
		return domain.Rating{}, err
	}

	if tx.Status != domain.TransactionCompleted {
		return domain.Rating{}, domain.ErrRatingNotAllowed
	}
	
	var targetUserID uuid.UUID
	if cmd.RaterID.String() == tx.BuyerID {
		target, _ := uuid.Parse(tx.SellerID)
		targetUserID = target
	} else {
		return domain.Rating{}, domain.ErrRatingNotAllowed
	}

	count, err := s.ratingRepo.CountRatingsBetweenUsers(ctx, cmd.RaterID, targetUserID)
	if err != nil {
		return domain.Rating{}, err
	}
	isFrozen := count > 5

	cmd.RatedUserID = targetUserID

	rating, err := s.ratingRepo.CreateRating(ctx, cmd, isFrozen)
	if err != nil {
		return domain.Rating{}, err
	}

	if isFrozen {
		return rating, nil
	}

	// 5. Update Target User Profile (Bayesian Calculation)
	user, err := s.userRepo.GetByID(ctx, targetUserID.String())
	if err != nil {
		return rating, nil // Non-fatal to rating creation if user fetch fails, but should be logged.
	}

	newTotal := user.TotalRatings + 1
	newSum := user.SumRatings + cmd.Stars

	// Fetch Global Average
	globalAvg, totalRatingsSystem, _ := s.ratingRepo.GetGlobalRatingAverage(ctx)
	
	// Cold Start protection
	if totalRatingsSystem < 100 {
		globalAvg = 4.0
	}

	// Bayesian Math: R = (n * r_avg + m * C) / (n + m)
	m := 10.0
	rAvg := float64(newSum) / float64(newTotal)
	n := float64(newTotal)
	
	bayesian := (n*rAvg + m*globalAvg) / (n + m)

	_ = s.ratingRepo.UpdateUserRating(ctx, targetUserID, newTotal, newSum, bayesian)

	return rating, nil
}

func (s *RatingService) GetPendingRatings(ctx context.Context, buyerID uuid.UUID) ([]domain.PendingRating, error) {
	return s.ratingRepo.GetPendingRatings(ctx, buyerID)
}

func (s *RatingService) GetUserReviews(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.ratingRepo.GetUserReviews(ctx, userID, limit, offset)
}
