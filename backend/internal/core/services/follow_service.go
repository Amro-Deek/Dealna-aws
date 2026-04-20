package services

import (
	"context"
	"errors"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type FollowService struct {
	repo ports.IFollowRepository
}

func NewFollowService(repo ports.IFollowRepository) *FollowService {
	return &FollowService{repo: repo}
}

// GetProfileID looks up the profile_id for a given user_id.
func (s *FollowService) GetProfileID(ctx context.Context, userID string) (string, error) {
	return s.repo.GetProfileIDByUserID(ctx, userID)
}

// FollowUser makes followerUserID follow the given targetProfileID.
// It resolves the follower's profile_id internally.
func (s *FollowService) FollowUser(ctx context.Context, followerUserID, targetProfileID string) error {
	followerProfileID, err := s.repo.GetProfileIDByUserID(ctx, followerUserID)
	if err != nil {
		return err
	}
	if followerProfileID == targetProfileID {
		return errors.New("cannot follow yourself")
	}
	return s.repo.Follow(ctx, followerProfileID, targetProfileID)
}

// UnfollowUser makes followerUserID unfollow the given targetProfileID.
func (s *FollowService) UnfollowUser(ctx context.Context, followerUserID, targetProfileID string) error {
	followerProfileID, err := s.repo.GetProfileIDByUserID(ctx, followerUserID)
	if err != nil {
		return err
	}
	return s.repo.Unfollow(ctx, followerProfileID, targetProfileID)
}

// IsFollowing checks if followerUserID follows targetProfileID.
func (s *FollowService) IsFollowing(ctx context.Context, followerUserID, targetProfileID string) (bool, error) {
	followerProfileID, err := s.repo.GetProfileIDByUserID(ctx, followerUserID)
	if err != nil {
		return false, err
	}
	return s.repo.IsFollowing(ctx, followerProfileID, targetProfileID)
}

// GetFollowers returns all profiles following the given profileID.
func (s *FollowService) GetFollowers(ctx context.Context, profileID string) ([]domain.Follow, error) {
	return s.repo.GetFollowers(ctx, profileID)
}

// GetFollowing returns all profiles that the given profileID follows.
func (s *FollowService) GetFollowing(ctx context.Context, profileID string) ([]domain.Follow, error) {
	return s.repo.GetFollowing(ctx, profileID)
}
