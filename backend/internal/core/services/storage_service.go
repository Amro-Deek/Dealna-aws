package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type StorageService struct {
	storage ports.IStorageProvider
}

func NewStorageService(storage ports.IStorageProvider) *StorageService {
	return &StorageService{storage: storage}
}

func (s *StorageService) GenerateProfilePictureUploadURL(ctx context.Context, userID, filename string, contentType string) (string, string, error) {
	// 1. Enforce content types (Security Requirment)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		return "", "", middleware.NewValidationError("content_type", "Only JPG, JPEG, and PNG images are allowed")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", "", middleware.NewValidationError("filename", "Invalid file extension")
	}

	// 2. Generate unique object key inside subfolder
	objectKey := fmt.Sprintf("profiles/%s/%s%s", userID, uuid.NewString(), ext)

	// 3. Generate presigned URL valid for 15 minutes
	url, err := s.storage.GeneratePresignedUploadURL(ctx, objectKey, contentType, 15*time.Minute)
	if err != nil {
		return "", "", middleware.NewInternalError(fmt.Errorf("failed to generate upload URL: %w", err))
	}

	return url, objectKey, nil
}
