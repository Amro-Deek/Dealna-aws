package postgres

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	coreDomain "github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type UserRepository struct {
    q *generated.Queries
}

// Compile-time interface check ✅
var _ ports.IUserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{
        q: generated.New(pool),
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

	return &coreDomain.User{
	ID:           userRow.UserID.String(),
	Email:        userRow.Email,
	Role:         userRow.Role,
	PasswordHash: "",
	KeycloakSub:  "",
}, nil
}
*/


// Dummy for now until signup flow is migrated to Keycloak
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
	return nil, middleware.NewUnauthorizedError("student creation is not migrated yet")
}
func extractStudentID(email string) string {
	parts := strings.Split(email, "@")
	local := parts[0]

	re := regexp.MustCompile(`\d+`)
	return re.FindString(local)
}



