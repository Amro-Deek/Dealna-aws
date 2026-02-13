package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type StudentPreRegistrationRepository struct {
	q *generated.Queries
}

var _ ports.IStudentPreRegistrationRepository = (*StudentPreRegistrationRepository)(nil)

func NewStudentPreRegistrationRepository(pool *pgxpool.Pool) *StudentPreRegistrationRepository {
	return &StudentPreRegistrationRepository{
		q: generated.New(pool),
	}
}

func (r *StudentPreRegistrationRepository) Create(
	ctx context.Context,
	pre *domain.StudentPreRegistration,
) error {

	return r.q.CreateStudentPreRegistration(ctx, generated.CreateStudentPreRegistrationParams{
		Email:     pre.Email,
		Token:     toUUID(pre.Token),
		ExpiresAt: toTimestamp(pre.ExpiresAt),
	})
}

func (r *StudentPreRegistrationRepository) GetByToken(
	ctx context.Context,
	token string,
) (*domain.StudentPreRegistration, error) {

	row, err := r.q.GetStudentPreRegistrationByToken(ctx, toUUID(token))
	if err != nil {
		return nil, err
	}

	return &domain.StudentPreRegistration{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Token:     fromUUID(row.Token),
		ExpiresAt: fromTimestamp(row.ExpiresAt),
		UsedAt:    fromNullableTimestamp(row.UsedAt),
	}, nil
}
func (r *StudentPreRegistrationRepository) Update(
	ctx context.Context,
	pre *domain.StudentPreRegistration,
) error {

	return r.q.UpdateStudentPreRegistration(ctx, generated.UpdateStudentPreRegistrationParams{
		ID:                toUUID(pre.ID),
		Token:             toUUID(pre.Token),
		ExpiresAt:         toTimestamp(pre.ExpiresAt),
		UsedAt:            toNullableTimestamp(pre.UsedAt),
		ResendCount:       int32(pre.ResendCount),
		ResendWindowStart: toNullableTimestamp(pre.ResendWindowStart),
		VerifiedAt:        toNullableTimestamp(pre.VerifiedAt),
	})
}


func (r *StudentPreRegistrationRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.StudentPreRegistration, error) {

	row, err := r.q.GetStudentPreRegistrationByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &domain.StudentPreRegistration{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Token:     fromUUID(row.Token),
		ExpiresAt: fromTimestamp(row.ExpiresAt),
		UsedAt:    fromNullableTimestamp(row.UsedAt),
		VerifiedAt: fromNullableTimestamp(row.VerifiedAt),
	}, nil
}


