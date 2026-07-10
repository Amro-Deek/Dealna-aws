package services

import (
	"context"
	"errors"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/jackc/pgx/v5"
)

type ProfileDTO struct {
	ProfileID                string    `json:"profile_id"`
	UserID                   string    `json:"user_id"`
	Email                    string    `json:"email"`
	Role                     string    `json:"role"`
	DisplayName              string    `json:"display_name"`
	Bio                      string    `json:"bio"`
	ProfilePictureURL        string    `json:"profile_picture_url"`
	DisplayNameLastChangedAt time.Time `json:"display_name_last_changed_at"`
	RatingCount              int       `json:"rating_count"`        // Deprecated
	TotalReviewsCount        int       `json:"total_reviews_count"` // Deprecated
	BayesianRating           float64   `json:"bayesian_rating"`
	TotalRatings             int       `json:"total_ratings"`
	SoldItemsCount           int       `json:"sold_items_count"`
	FollowerCount            int       `json:"follower_count"`
	FollowingCount           int       `json:"following_count"`
	JoinedAt                 time.Time `json:"joined_at"`

	// Student specific
	Major        *string `json:"major,omitempty"`
	AcademicYear *int    `json:"academic_year,omitempty"`
	StudentID    *string `json:"student_id,omitempty"` // Only returned for own profile
}

type ProfileService struct {
	users ports.IUserRepository
}

func NewProfileService(users ports.IUserRepository) *ProfileService {
	return &ProfileService{users: users}
}

func (s *ProfileService) GetMyProfile(ctx context.Context, userID string) (*ProfileDTO, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, middleware.NewDatabaseError("get user by id", err)
	}

	profile, student, err := s.users.GetProfile(ctx, userID)
	if err != nil {
		return nil, middleware.NewDatabaseError("get profile", err)
	}

	dto := &ProfileDTO{
		ProfileID: profile.ProfileID,
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
	}

	if profile != nil {
		dto.DisplayName = profile.DisplayName
		dto.Bio = profile.Bio
		dto.ProfilePictureURL = profile.ProfilePictureURL
		dto.DisplayNameLastChangedAt = profile.DisplayNameLastChangedAt
		dto.RatingCount = profile.RatingCount
		dto.TotalReviewsCount = profile.TotalReviewsCount
		dto.BayesianRating = user.BayesianRating
		dto.TotalRatings = user.TotalRatings
		dto.SoldItemsCount = profile.SoldItemsCount
		dto.FollowerCount = profile.FollowerCount
		dto.FollowingCount = profile.FollowingCount
		dto.JoinedAt = profile.CreatedAt
	}

	if student != nil {
		dto.Major = &student.Major
		dto.AcademicYear = &student.AcademicYear
		dto.StudentID = &student.StudentID
	}

	return dto, nil
}

func (s *ProfileService) GetPublicProfile(ctx context.Context, profileID string) (*ProfileDTO, error) {
	profile, err := s.users.GetProfileByProfileID(ctx, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, middleware.NewUserNotFoundError(profileID)
		}
		return nil, middleware.NewDatabaseError("get public profile", err)
	}

	user, err := s.users.GetByID(ctx, profile.UserID)
	if err != nil {
		return nil, middleware.NewDatabaseError("get user for public profile", err)
	}

	// We only return public info — student_id is intentionally excluded for privacy
	dto := &ProfileDTO{
		ProfileID:         profile.ProfileID,
		UserID:            user.ID,
		Email:             user.Email,
		DisplayName:       profile.DisplayName,
		Bio:               profile.Bio,
		ProfilePictureURL: profile.ProfilePictureURL,
		RatingCount:       profile.RatingCount,
		TotalReviewsCount: profile.TotalReviewsCount,
		BayesianRating:    user.BayesianRating,
		TotalRatings:      user.TotalRatings,
		SoldItemsCount:    profile.SoldItemsCount,
		FollowerCount:     profile.FollowerCount,
		FollowingCount:    profile.FollowingCount,
		JoinedAt:          profile.CreatedAt,
		Role:              user.Role,
	}

	// Fetch student-specific public fields (major and academic year)
	_, student, err := s.users.GetProfile(ctx, user.ID)
	if err == nil && student != nil {
		dto.Major = &student.Major
		dto.AcademicYear = &student.AcademicYear
		// NOTE: StudentID is NOT populated here for privacy
	}

	return dto, nil
}

func (s *ProfileService) GetPublicProfileByUserID(ctx context.Context, userID string) (*ProfileDTO, error) {
	profile, err := s.users.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If profile doesn't exist, try to get basic user info anyway for admin/display purposes
			user, uErr := s.users.GetByID(ctx, userID)
			if uErr != nil {
				return nil, middleware.NewUserNotFoundError(userID)
			}
			return &ProfileDTO{
				UserID:      user.ID,
				Email:       user.Email,
				Role:        user.Role,
				DisplayName: "No Profile Yet",
				Bio:         "This user has not set up their profile.",
				JoinedAt:    time.Now(),
			}, nil
		}
		return nil, middleware.NewDatabaseError("get public profile by user id", err)
	}

	user, err := s.users.GetByID(ctx, profile.UserID)
	if err != nil {
		return nil, middleware.NewDatabaseError("get user for public profile", err)
	}

	// We only return public info — student_id is intentionally excluded for privacy
	dto := &ProfileDTO{
		ProfileID:         profile.ProfileID,
		UserID:            user.ID,
		Email:             user.Email,
		DisplayName:       profile.DisplayName,
		Bio:               profile.Bio,
		ProfilePictureURL: profile.ProfilePictureURL,
		RatingCount:       profile.RatingCount,
		TotalReviewsCount: profile.TotalReviewsCount,
		BayesianRating:    user.BayesianRating,
		TotalRatings:      user.TotalRatings,
		SoldItemsCount:    profile.SoldItemsCount,
		FollowerCount:     profile.FollowerCount,
		FollowingCount:    profile.FollowingCount,
		JoinedAt:          profile.CreatedAt,
		Role:              user.Role,
	}

	// Fetch student-specific public fields (major and academic year)
	_, student, err := s.users.GetProfile(ctx, user.ID)
	if err == nil && student != nil {
		dto.Major = &student.Major
		dto.AcademicYear = &student.AcademicYear
		// NOTE: StudentID is NOT populated here for privacy
	}

	return dto, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, displayName, bio, profilePictureURL *string) error {
	var newDisplayNameLastChangedAt *time.Time

	// If displayName is being updated, enforce business rules
	if displayName != nil {
		profile, _, err := s.users.GetProfile(ctx, userID)
		if err != nil {
			return middleware.NewDatabaseError("get profile for verification", err)
		}

		if profile != nil && profile.DisplayName != *displayName {
			// Check rate limit: 30 days
			daysSinceLastChange := int(time.Since(profile.DisplayNameLastChangedAt).Hours() / 24)
			if daysSinceLastChange < 30 {
				return middleware.NewValidationError("display_name", "Display name can only be changed once every 30 days")
			}

			now := time.Now()
			newDisplayNameLastChangedAt = &now
		}
	}

	return s.users.UpdateProfile(ctx, userID, displayName, bio, profilePictureURL, newDisplayNameLastChangedAt)
}

func (s *ProfileService) UpdateStudentDetails(ctx context.Context, userID string, major *string, year *int) error {
	return s.users.UpdateStudent(ctx, userID, major, year)
}

func (s *ProfileService) UpdateDeviceToken(ctx context.Context, userID string, token string) error {
	return s.users.UpdateDeviceToken(ctx, userID, token)
}
