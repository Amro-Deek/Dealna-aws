package postgres

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	coreDomain "github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type UserRepository struct {
    q    *generated.Queries
    pool *pgxpool.Pool
}

// Compile-time interface check ✅
var _ ports.IUserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{
        q:    generated.New(pool),
        pool: pool,
    }
}
func (r *UserRepository) GetByID(
	ctx context.Context,
	userID string,
) (*coreDomain.User, error) {

	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, err
	}

	row, err := r.q.GetUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &coreDomain.User{
		ID:          uuidToString(row.UserID),
		Email:       row.Email,
		Role:        row.Role,
		KeycloakSub: uuidToString(row.KeycloakSub),
	}, nil
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*coreDomain.User, error) {

	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &coreDomain.User{
		ID:          uuidToString(row.UserID),
		Email:       row.Email,
		Role:        row.Role,
		KeycloakSub: uuidToString(row.KeycloakSub),
	}, nil
}

func (r *UserRepository) GetByKeycloakSub(
	ctx context.Context,
	sub string,
) (*coreDomain.User, error) {

	var subUUID pgtype.UUID
	if err := subUUID.Scan(sub); err != nil {
		return nil, err
	}

	row, err := r.q.GetUserByKeycloakSub(ctx, subUUID)
	if err != nil {
		return nil, err
	}

	return &coreDomain.User{
		ID:          uuidToString(row.UserID),
		Email:       row.Email,
		Role:        row.Role,
		KeycloakSub: uuidToString(row.KeycloakSub),
	}, nil
}
// -----------------------------
// Shared mapper (DRY)
// -----------------------------
func mapUser(
	id pgtype.UUID,
	email string,
	passwordHash string,
	role string,
	keycloakSub string,
) *coreDomain.User {
	return &coreDomain.User{
		ID:           id.String(),
		Email:        email,
		Role:         role,
		KeycloakSub:  keycloakSub,
	}
}

/*
func (r *UserRepository) CreateStudent(
	ctx context.Context,
	displayName string,
	email string,
	passwordHash string,
	major *string,
	year *int,
	universityID string,
	studentID string,
) (*coreDomain.User, error) {

	userRow, err := r.q.CreateStudentUser(ctx, generated.CreateStudentUserParams{
		Email:        email,
		PasswordHash: toText(passwordHash),
		UniversityID: toUUID(universityID),
	})
	if err != nil {
		return nil, err
	}

	err = r.q.CreateStudent(ctx, generated.CreateStudentParams{
		UserID:       userRow.UserID,
		StudentID:    studentID,
		Major:        toNullableText(major),
		AcademicYear: toNullableInt32(year),
	})
	if err != nil {
		return nil, err
	}

	err = r.q.CreateProfile(ctx, generated.CreateProfileParams{
		UserID:      userRow.UserID,
		DisplayName: toNullableText(&displayName),
	})
	if err != nil {
		return nil, err
	}

	return &coreDomain.User{
	ID:           userRow.UserID.String(),
	Email:        userRow.Email,
	Role:         userRow.Role,
	PasswordHash: "",
	KeycloakSub:  "",
}, nil
}
*/


func (r *UserRepository) CreateStudent(
	ctx context.Context,
	displayName string,
	email string,
	keycloakSub string,
	major *string,
	year *int,
	universityID string,
	studentID string,
) (*coreDomain.User, error) {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	userRow, err := qtx.CreateStudentUser(ctx, generated.CreateStudentUserParams{
		Email:        email,
		UniversityID: toUUID(universityID),
		KeycloakSub:  toUUID(keycloakSub),
	})
	if err != nil {
		return nil, err
	}

	err = qtx.CreateStudent(ctx, generated.CreateStudentParams{
		UserID:       userRow.UserID,
		StudentID:    studentID,
		Major:        toNullableText(major),
		AcademicYear: toNullableInt32(year),
	})
	if err != nil {
		return nil, err
	}

	err = qtx.CreateProfile(ctx, generated.CreateProfileParams{
		UserID:      userRow.UserID,
		DisplayName: toNullableText(&displayName),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &coreDomain.User{
		ID:          uuidToString(userRow.UserID),
		Email:       userRow.Email,
		Role:        userRow.Role,
		KeycloakSub: uuidToString(userRow.KeycloakSub),
	}, nil
}
func extractStudentID(email string) string {
	parts := strings.Split(email, "@")
	local := parts[0]

	re := regexp.MustCompile(`\d+`)
	return re.FindString(local)
}

func (r *UserRepository) GetProfile(ctx context.Context, userID string) (*coreDomain.Profile, *coreDomain.Student, error) {
	uid := toUUID(userID)

	profileRow, err := r.q.GetProfileByUserID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	profile := &coreDomain.Profile{
		ProfileID:                uuidToString(profileRow.ProfileID),
		UserID:                   uuidToString(profileRow.UserID),
		DisplayName:              fromNullableText(profileRow.DisplayName),
		Bio:                      fromNullableText(profileRow.Bio),
		ProfilePictureURL:        fromNullableText(profileRow.ProfilePictureUrl),
		DisplayNameLastChangedAt: fromNullableTime(profileRow.DisplayNameLastChangedAt),
		RatingCount:              int(profileRow.RatingCount),
		TotalReviewsCount:        int(profileRow.TotalReviewsCount),
		SoldItemsCount:           int(profileRow.SoldItemsCount),
		FollowerCount:            int(profileRow.FollowerCount),
		FollowingCount:           int(profileRow.FollowingCount),
	}

	var student *coreDomain.Student
	studentRow, err := r.q.GetStudentByUserID(ctx, uid)
	if err == nil {
		student = &coreDomain.Student{
			UserID:             uuidToString(studentRow.UserID),
			StudentID:          studentRow.StudentID,
			Major:              fromNullableText(studentRow.Major),
			AcademicYear:       int(fromNullableInt32(studentRow.AcademicYear)),
		}
	}

	return profile, student, nil
}

func (r *UserRepository) GetProfileByProfileID(ctx context.Context, profileID string) (*coreDomain.Profile, error) {
	pid := toUUID(profileID)

	profileRow, err := r.q.GetProfileByProfileID(ctx, pid)
	if err != nil {
		return nil, err
	}

	return &coreDomain.Profile{
		ProfileID:                uuidToString(profileRow.ProfileID),
		UserID:                   uuidToString(profileRow.UserID),
		DisplayName:              fromNullableText(profileRow.DisplayName),
		Bio:                      fromNullableText(profileRow.Bio),
		ProfilePictureURL:        fromNullableText(profileRow.ProfilePictureUrl),
		DisplayNameLastChangedAt: fromNullableTime(profileRow.DisplayNameLastChangedAt),
		RatingCount:              int(profileRow.RatingCount),
		TotalReviewsCount:        int(profileRow.TotalReviewsCount),
		SoldItemsCount:           int(profileRow.SoldItemsCount),
		FollowerCount:            int(profileRow.FollowerCount),
		FollowingCount:           int(profileRow.FollowingCount),
	}, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID string, displayName, bio, profilePictureURL *string, displayNameLastChangedAt *time.Time) error {
	return r.q.UpdateProfile(ctx, generated.UpdateProfileParams{
		UserID:                   toUUID(userID),
		DisplayName:              toNullableText(displayName),
		Bio:                      toNullableText(bio),
		ProfilePictureUrl:        toNullableText(profilePictureURL),
		DisplayNameLastChangedAt: toNullableTimePtr(displayNameLastChangedAt),
	})
}

func (r *UserRepository) UpdateStudent(ctx context.Context, userID string, major *string, year *int) error {
	return r.q.UpdateStudent(ctx, generated.UpdateStudentParams{
		UserID:       toUUID(userID),
		Major:        toNullableText(major),
		AcademicYear: toNullableInt32Ptr(year),
	})
}



