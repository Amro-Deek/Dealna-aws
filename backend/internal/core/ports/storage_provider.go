package ports

import (
	"context"
	"io"
	"time"
)

type IStorageProvider interface {
	// GeneratePresignedUploadURL generates a temporary URL that allows a client to upload a file directly to the storage provider.
	GeneratePresignedUploadURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error)
	// GeneratePresignedDownloadURL generates a temporary URL that allows a client to download a file directly from the storage provider.
	GeneratePresignedDownloadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	UploadFile(ctx context.Context, objectKey string, contentType string, content io.Reader) (string, error)
}
