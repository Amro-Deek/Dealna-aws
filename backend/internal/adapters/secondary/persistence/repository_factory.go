package persistence

import (
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence/postgres"
    "github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type RepositoryFactory struct {
    pool *pgxpool.Pool
}

func NewRepositoryFactory(pool *pgxpool.Pool) *RepositoryFactory {
    return &RepositoryFactory{pool: pool}
}

func (f *RepositoryFactory) User() ports.IUserRepository {
    return postgres.NewUserRepository(f.pool)
}

