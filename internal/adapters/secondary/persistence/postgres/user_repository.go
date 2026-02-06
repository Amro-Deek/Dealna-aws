package postgres

import (
    "context"

    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/Amro-Deek/Dealna-aws/internal/core/domain"
    "github.com/Amro-Deek/Dealna-aws/internal/core/ports"
    "github.com/Amro-Deek/Dealna-aws/internal/database/generated"
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
) (*domain.User, error) {

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
    return &domain.User{
        ID:   row.UserID.String(),
        Role: row.Role,
    }, nil
}

// -----------------------------
// Get user by Email (NEW)
// -----------------------------
func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

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
) *domain.User {
	return &domain.User{
		ID:           id.String(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}
}