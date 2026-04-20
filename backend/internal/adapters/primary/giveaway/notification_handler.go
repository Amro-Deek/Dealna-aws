package giveaway

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	nService *services.NotificationService
}

func NewNotificationHandler(nService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{nService: nService}
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	notifs, err := h.nService.GetNotificationsForUser(r.Context(), userID, 50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "notificationId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.nService.MarkNotificationRead(r.Context(), notificationID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
