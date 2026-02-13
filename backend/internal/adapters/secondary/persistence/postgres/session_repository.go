package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type SessionRepository struct {
	q *generated.Queries
}

var _ ports.ISessionRepository = (*SessionRepository)(nil)

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		q: generated.New(pool),
	}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	userID string,
	jti string,
	expiresAt time.Time,
) error {

	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}

	var j pgtype.UUID
	if err := j.Scan(jti); err != nil {
		return err
	}

	return r.q.CreateSession(ctx, generated.CreateSessionParams{
		UserID:    uid,
		Jti:       j,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
}

func (r *SessionRepository) GetByJTI(
	ctx context.Context,
	jti string,
) (*domain.Session, error) {

	var j pgtype.UUID
	if err := j.Scan(jti); err != nil {
		return nil, err
	}

	row, err := r.q.GetSessionByJTI(ctx, j)
	if err != nil {
		return nil, err
	}

	return &domain.Session{
		SessionID: row.SessionID.String(),
		UserID:    row.UserID.String(),
		JTI:       row.Jti.String(),
		Revoked:   row.Revoked,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}

func (r *SessionRepository) RevokeByJTI(
	ctx context.Context,
	jti string,
) error {

	var j pgtype.UUID
	if err := j.Scan(jti); err != nil {
		return err
	}

	return r.q.RevokeSessionByJTI(ctx, j)
}

func (r *SessionRepository) RevokeAllForUser(
	ctx context.Context,
	userID string,
) error {

	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}

	return r.q.RevokeAllSessionsForUser(ctx, uid)
}
