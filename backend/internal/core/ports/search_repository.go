package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

// ISearchRepository defines operations for the vector search engine (Qdrant).
type ISearchRepository interface {
	SearchItems(ctx context.Context, vector []float32, filter domain.ItemFilter) ([]uuid.UUID, error)
}
