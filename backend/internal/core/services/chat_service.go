package services

import (
	"context"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type ChatService struct {
	authProvider ports.IFirebaseAuthProvider
	notifs       *NotificationService
}

func NewChatService(authProvider ports.IFirebaseAuthProvider, notifs *NotificationService) *ChatService {
	return &ChatService{authProvider: authProvider, notifs: notifs}
}

func (s *ChatService) GetChatToken(ctx context.Context, userID string) (string, error) {
	return s.authProvider.GenerateCustomToken(ctx, userID)
}

func (s *ChatService) SendChatNotification(ctx context.Context, senderID, receiverID, roomID, itemID string) error {
	if s.notifs == nil {
		return nil
	}
	
	notifCtx := NotificationContext{
		ActingUserID: &senderID,
		RoomID:       &roomID,
		ItemID:       &itemID,
	}
	
	return s.notifs.CreateNotification(ctx, receiverID, domain.NotifTypeChatMessage, notifCtx)
}
