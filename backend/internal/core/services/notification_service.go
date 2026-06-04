package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/messaging"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type NotificationService struct {
	repo      ports.INotificationRepository
	userRepo  ports.IUserRepository
	itemRepo  ports.ItemRepository
	fcmClient *messaging.FCMClient
}

func NewNotificationService(repo ports.INotificationRepository, userRepo ports.IUserRepository, itemRepo ports.ItemRepository, fcmClient *messaging.FCMClient) *NotificationService {
	return &NotificationService{repo: repo, userRepo: userRepo, itemRepo: itemRepo, fcmClient: fcmClient}
}

type NotificationContext struct {
	ActingUserID *string
	ItemID       *string
	EntryID      *string
	TxID         *string
}

func (s *NotificationService) CreateNotification(ctx context.Context, userID string, typ domain.NotificationType, notifCtx NotificationContext) error {
	// Build rich payload
	payloadMap := make(map[string]interface{})
	if notifCtx.ItemID != nil {
		payloadMap["item_id"] = *notifCtx.ItemID
		// Enrich with item data
		parsedItemID, err := uuid.Parse(*notifCtx.ItemID)
		if err == nil {
			item, err := s.itemRepo.GetItemDetail(ctx, parsedItemID)
			if err == nil {
				payloadMap["item_title"] = item.Title
				if item.ThumbnailURL != "" {
					payloadMap["item_image_url"] = item.ThumbnailURL
				}
			}
		}
	}
	if notifCtx.EntryID != nil {
		payloadMap["entry_id"] = *notifCtx.EntryID
	}
	if notifCtx.TxID != nil {
		payloadMap["tx_id"] = *notifCtx.TxID
	}
	
	actingUserName := "System"
	if notifCtx.ActingUserID != nil {
		payloadMap["acting_user_id"] = *notifCtx.ActingUserID
		// Enrich with acting user data
		profile, err := s.userRepo.GetProfileByUserID(ctx, *notifCtx.ActingUserID)
		if err == nil && profile != nil {
			actingUserName = profile.DisplayName
			payloadMap["acting_user_name"] = actingUserName
		}
	}

	payload, _ := json.Marshal(payloadMap)

	_, err := s.repo.CreateNotification(ctx, &domain.Notification{
		UserID:  userID,
		Type:    typ,
		Payload: payload,
	})
	if err != nil {
		return err
	}

	// Fetch user's device token to send FCM Push
	if s.fcmClient != nil && s.userRepo != nil {
		profile, _ := s.userRepo.GetProfileByUserID(ctx, userID)
		if profile != nil && profile.DeviceToken != nil {
			unreadCount, _ := s.CountUnreadNotifications(ctx, userID)
			
			// Map NotificationType to Title & Body
			title, body := getNotificationText(typ, actingUserName, payloadMap)
			
			// Convert payloadMap to string map for FCM Data
			data := make(map[string]string)
			data["notification_type"] = string(typ)
			for k, v := range payloadMap {
				data[k] = fmt.Sprintf("%v", v)
			}
			data["unread_count"] = fmt.Sprintf("%d", unreadCount)

			// Fire and forget
			go func() {
				err := s.fcmClient.SendVisiblePush(context.Background(), *profile.DeviceToken, title, body, data)
				if err != nil {
					fmt.Printf("❌ Failed to send FCM push notification to token %s: %v\n", *profile.DeviceToken, err)
				} else {
					fmt.Printf("✅ Successfully sent FCM push notification to token %s\n", *profile.DeviceToken)
				}
			}()
		} else {
			fmt.Printf("⚠️ Cannot send FCM push notification: user %s has no device token\n", userID)
		}
	} else {
		fmt.Printf("⚠️ Cannot send FCM push notification: fcmClient or userRepo is nil\n")
	}

	return nil
}

func getNotificationText(typ domain.NotificationType, actingUserName string, payloadMap map[string]interface{}) (string, string) {
	itemTitle, _ := payloadMap["item_title"].(string)
	if itemTitle == "" {
		itemTitle = "an item"
	}

	switch typ {
	case domain.NotifTypeTurnStarted:
		return "It's your turn! 🎉", "You are next in line for " + itemTitle + ". You have 24 hours to accept!"
	case domain.NotifTypeHandoffInitiated:
		return "Handoff Initiated 🤝", actingUserName + " has initiated the handoff for " + itemTitle + "."
	case domain.NotifTypeGiveawayCompleted:
		return "Giveaway Completed! ✅", actingUserName + " confirmed receiving " + itemTitle + "."
	case domain.NotifTypePurchaseRequested:
		return "New Purchase Request! 🛍️", actingUserName + " wants to buy your " + itemTitle + "."
	case domain.NotifTypePurchaseAccepted:
		return "Purchase Accepted! 🥳", actingUserName + " accepted your request for " + itemTitle + "."
	case domain.NotifTypePurchaseRejected:
		return "Purchase Rejected ❌", actingUserName + " rejected your request for " + itemTitle + "."
	case domain.NotifTypeGiveawayCancelled:
		return "Request Cancelled ⚠️", actingUserName + " cancelled their request for " + itemTitle + "."
	case domain.NotifTypeTransactionDone:
		return "Transaction Completed! ✅", actingUserName + " confirmed the transaction for " + itemTitle + "."
	default:
		return "New Notification", "You have a new update regarding " + itemTitle + "."
	}
}

func (s *NotificationService) GetNotificationsForUser(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkNotificationRead(ctx context.Context, notificationID, userID string) error {
	return s.repo.MarkNotificationRead(ctx, notificationID, userID)
}

func (s *NotificationService) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnreadNotifications(ctx, userID)
}
