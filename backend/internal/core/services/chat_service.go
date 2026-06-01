package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type ChatService struct {
	authProvider ports.IFirebaseAuthProvider
}

func NewChatService(authProvider ports.IFirebaseAuthProvider) *ChatService {
	return &ChatService{authProvider: authProvider}
}

func (s *ChatService) GetChatToken(ctx context.Context, userID string) (string, error) {
	return s.authProvider.GenerateCustomToken(ctx, userID)
}
