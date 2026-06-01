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
