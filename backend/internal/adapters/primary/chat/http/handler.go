package http

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type ChatHandler struct {
	chatSvc *services.ChatService
	logger  middleware.StructuredLoggerInterface
}

func NewChatHandler(chatSvc *services.ChatService, logger middleware.StructuredLoggerInterface) *ChatHandler {
	return &ChatHandler{
		chatSvc: chatSvc,
		logger:  logger,
	}
}

// @Summary Get Firebase Chat Token
// @Description Returns a custom Firebase Auth token for the logged-in user to connect to Firebase directly.
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/chat/token [get]
func (h *ChatHandler) GetChatToken(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("unauthorized"), h.logger)
		return
	}

	token, err := h.chatSvc.GetChatToken(r.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"firebaseToken": token,
	})
}

type SendChatNotificationRequest struct {
	RoomID     string `json:"room_id"`
	ReceiverID string `json:"receiver_id"`
	ItemID     string `json:"item_id"`
}

// @Summary Send Chat Notification
// @Description Sends a push notification to the receiver of a chat message. Called by mobile app after writing to Firestore.
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SendChatNotificationRequest true "Notification details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/chat/notify [post]
func (h *ChatHandler) SendChatNotification(w http.ResponseWriter, r *http.Request) {
	senderID := middleware.UserIDFromContext(r.Context())
	if senderID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewUnauthorizedError("unauthorized"), h.logger)
		return
	}

	var req SendChatNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("invalid JSON payload"), h.logger)
		return
	}

	if req.RoomID == "" || req.ReceiverID == "" || req.ItemID == "" {
		middleware.WriteErrorResponse(w, r.Context(), middleware.NewValidationError("room_id, receiver_id, and item_id are required"), h.logger)
		return
	}

	err := h.chatSvc.SendChatNotification(r.Context(), senderID, req.ReceiverID, req.RoomID, req.ItemID)
	if err != nil {
		middleware.WriteErrorResponse(w, r.Context(), err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification sent successfully",
	})
}
