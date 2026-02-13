package postgres

import (
    "context"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"
    "strings"
	"regexp"

	coreDomain "github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
    "github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
    "github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
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

    // استدعاء sqlc
    row, err := r.q.GetUserByID(ctx, uid)
    if err != nil {
        return nil, err
    }

    // mapping من DB model → Domain model
    return &coreDomain.User{
        ID:   row.UserID.String(),
        Role: row.Role,
        Email: row.Email,
    }, nil
}

// -----------------------------
// Get user by Email (NEW)
// -----------------------------
func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*coreDomain.User, error) {

	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return mapUser(row.UserID, row.Email, row.PasswordHash.String, row.Role), nil
}

// -----------------------------
// Shared mapper (DRY)
// -----------------------------
func mapUser(
	id pgtype.UUID,
	email string,
	passwordHash string,
	role string,
) *coreDomain.User {
	return &coreDomain.User{
		ID:           id.String(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}
}
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
		ID:    userRow.UserID.String(),
		Email: userRow.Email,
		Role:  userRow.Role,
	}, nil
}




func extractStudentID(email string) string {
	parts := strings.Split(email, "@")
	local := parts[0]

	re := regexp.MustCompile(`\d+`)
	return re.FindString(local)
}



