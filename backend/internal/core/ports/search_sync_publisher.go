package ports

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

// ISearchSyncPublisher defines how we send events to the Search Indexing Worker (Lambda/Qdrant)
type ISearchSyncPublisher interface {
	PublishSyncEvent(ctx context.Context, event domain.SearchSyncEvent) error
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}
