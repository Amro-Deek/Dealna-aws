package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	coreDomain "github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
)

type UniversityRepository struct {
	q *generated.Queries
}

var _ ports.IUniversityRepository = (*UniversityRepository)(nil)

func NewUniversityRepository(pool *pgxpool.Pool) *UniversityRepository {
	return &UniversityRepository{
		q: generated.New(pool),
	}
}

func (r *UniversityRepository) GetByDomain(
	ctx context.Context,
	domain string,
) (*coreDomain.University, error) {

	row, err := r.q.GetUniversityByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}

	return &coreDomain.University{
		ID:     row.UniversityID.String(),
		Name:   row.Name,
		Domain: row.Domain,
		Status: row.Status,
	}, nil
}
