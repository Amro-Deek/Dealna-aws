package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type ProviderPreRegistrationRepository struct {
	q *generated.Queries
}

var _ ports.IProviderPreRegistrationRepository = (*ProviderPreRegistrationRepository)(nil)

func NewProviderPreRegistrationRepository(pool *pgxpool.Pool) *ProviderPreRegistrationRepository {
	return &ProviderPreRegistrationRepository{
		q: generated.New(pool),
	}
}

func (r *ProviderPreRegistrationRepository) Create(
	ctx context.Context,
	pre *domain.ProviderPreRegistration,
) error {

	_, err := r.q.CreateProviderPreRegistration(ctx, generated.CreateProviderPreRegistrationParams{
		Email:     pre.Email,
		Token:     toUUID(pre.Token),
		ExpiresAt: toTimestamp(pre.ExpiresAt),
	})
	return err
}

func (r *ProviderPreRegistrationRepository) GetByToken(
	ctx context.Context,
	token string,
) (*domain.ProviderPreRegistration, error) {

	row, err := r.q.GetProviderPreRegistrationByToken(ctx, toUUID(token))
	if err != nil {
		return nil, err
	}

	return &domain.ProviderPreRegistration{
		ID:                fromUUID(row.ID),
		Email:             row.Email,
		Token:             fromUUID(row.Token),
		ExpiresAt:         fromTimestamp(row.ExpiresAt),
		UsedAt:            fromNullableTimestamp(row.UsedAt),
		ResendCount:       int(row.ResendCount),
		ResendWindowStart: fromNullableTimestamp(row.ResendWindowStart),
		VerifiedAt:        fromNullableTimestamp(row.VerifiedAt),
	}, nil
}

func (r *ProviderPreRegistrationRepository) Update(
	ctx context.Context,
	pre *domain.ProviderPreRegistration,
) error {

	return r.q.UpdateProviderPreRegistration(ctx, generated.UpdateProviderPreRegistrationParams{
		ID:                toUUID(pre.ID),
		Token:             toUUID(pre.Token),
		ExpiresAt:         toTimestamp(pre.ExpiresAt),
		UsedAt:            toNullableTimestamp(pre.UsedAt),
		ResendCount:       int32(pre.ResendCount),
		ResendWindowStart: toNullableTimestamp(pre.ResendWindowStart),
		VerifiedAt:        toNullableTimestamp(pre.VerifiedAt),
	})
}

func (r *ProviderPreRegistrationRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.ProviderPreRegistration, error) {

	row, err := r.q.GetProviderPreRegistrationByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &domain.ProviderPreRegistration{
		ID:                fromUUID(row.ID),
		Email:             row.Email,
		Token:             fromUUID(row.Token),
		ExpiresAt:         fromTimestamp(row.ExpiresAt),
		UsedAt:            fromNullableTimestamp(row.UsedAt),
		ResendCount:       int(row.ResendCount),
		ResendWindowStart: fromNullableTimestamp(row.ResendWindowStart),
		VerifiedAt:        fromNullableTimestamp(row.VerifiedAt),
	}, nil
}
