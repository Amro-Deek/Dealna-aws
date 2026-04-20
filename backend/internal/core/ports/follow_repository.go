package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

type IFollowRepository interface {
	Follow(ctx context.Context, followerProfileID, followingProfileID string) error
	Unfollow(ctx context.Context, followerProfileID, followingProfileID string) error
	IsFollowing(ctx context.Context, followerProfileID, followingProfileID string) (bool, error)
	GetFollowers(ctx context.Context, profileID string) ([]domain.Follow, error)
	GetFollowing(ctx context.Context, profileID string) ([]domain.Follow, error)
	GetProfileIDByUserID(ctx context.Context, userID string) (string, error)
}
