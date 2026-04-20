package postgres

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowRepository struct {
	q *generated.Queries
}

func NewFollowRepository(conn *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{q: generated.New(conn)}
}

func (r *FollowRepository) GetProfileIDByUserID(ctx context.Context, userID string) (string, error) {
	id, err := r.q.GetProfileIDByUserID(ctx, toUUID(userID))
	if err != nil {
		return "", err
	}
	return uuidToString(id), nil
}

func (r *FollowRepository) Follow(ctx context.Context, followerProfileID, followingProfileID string) error {
	followerUUID := toUUID(followerProfileID)
	followingUUID := toUUID(followingProfileID)

	if err := r.q.FollowUser(ctx, generated.FollowUserParams{
		FollowerProfileID:  followerUUID,
		FollowingProfileID: followingUUID,
	}); err != nil {
		return err
	}
	// Bump counters atomically
	_ = r.q.UpdateFollowingCount(ctx, generated.UpdateFollowingCountParams{ProfileID: followerUUID, FollowingCount: 1})
	_ = r.q.UpdateFollowerCount(ctx, generated.UpdateFollowerCountParams{ProfileID: followingUUID, FollowerCount: 1})
	return nil
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerProfileID, followingProfileID string) error {
	followerUUID := toUUID(followerProfileID)
	followingUUID := toUUID(followingProfileID)

	if err := r.q.UnfollowUser(ctx, generated.UnfollowUserParams{
		FollowerProfileID:  followerUUID,
		FollowingProfileID: followingUUID,
	}); err != nil {
		return err
	}
	// Decrement counters
	_ = r.q.UpdateFollowingCount(ctx, generated.UpdateFollowingCountParams{ProfileID: followerUUID, FollowingCount: -1})
	_ = r.q.UpdateFollowerCount(ctx, generated.UpdateFollowerCountParams{ProfileID: followingUUID, FollowerCount: -1})
	return nil
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerProfileID, followingProfileID string) (bool, error) {
	return r.q.IsFollowing(ctx, generated.IsFollowingParams{
		FollowerProfileID:  toUUID(followerProfileID),
		FollowingProfileID: toUUID(followingProfileID),
	})
}

func (r *FollowRepository) GetFollowers(ctx context.Context, profileID string) ([]domain.Follow, error) {
	rows, err := r.q.GetFollowers(ctx, toUUID(profileID))
	if err != nil {
		return nil, err
	}
	result := make([]domain.Follow, len(rows))
	for i, row := range rows {
		result[i] = domain.Follow{
			FollowerProfileID:  uuidToString(row.FollowerProfileID),
			FollowingProfileID: uuidToString(row.FollowingProfileID),
			FollowedAt:         row.FollowedAt.Time,
			DisplayName:        row.DisplayName.String,
			ProfilePictureURL:  row.ProfilePictureUrl.String,
		}
	}
	return result, nil
}

func (r *FollowRepository) GetFollowing(ctx context.Context, profileID string) ([]domain.Follow, error) {
	rows, err := r.q.GetFollowing(ctx, toUUID(profileID))
	if err != nil {
		return nil, err
	}
	result := make([]domain.Follow, len(rows))
	for i, row := range rows {
		result[i] = domain.Follow{
			FollowerProfileID:  uuidToString(row.FollowerProfileID),
			FollowingProfileID: uuidToString(row.FollowingProfileID),
			FollowedAt:         row.FollowedAt.Time,
			DisplayName:        row.DisplayName.String,
			ProfilePictureURL:  row.ProfilePictureUrl.String,
		}
	}
	return result, nil
}
