package ports

import (
	"context"
	"time"
)

type IStorageProvider interface {
	// GeneratePresignedUploadURL generates a temporary URL that allows a client to upload a file directly to the storage provider.
	GeneratePresignedUploadURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error)
}
