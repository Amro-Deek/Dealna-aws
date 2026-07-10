package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/google/uuid"
)

// ISearchRepository defines operations for the vector search engine (Qdrant).
type ISearchRepository interface {
	SearchItems(ctx context.Context, vector []float32, filter domain.ItemFilter) ([]uuid.UUID, error)
	FindSimilar(ctx context.Context, itemID string, limit uint64) ([]uuid.UUID, error)
}

