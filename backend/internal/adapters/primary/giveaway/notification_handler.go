package giveaway

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// ListNotifications godoc
// @Summary      Get user notifications
// @Description  Returns a paginated list of notifications for the current user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit    query     int  false  "limit" default(50)
// @Param        offset   query     int  false  "offset" default(0)
// @Success      200 {array}  domain.Notification
// @Router       /giveaway/notifications [get]
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	notifs, err := h.nService.GetNotificationsForUser(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(notifs)
}

// MarkRead godoc
// @Summary      Mark notification as read
// @Description  Marks a specific notification as read for the current user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        notificationId   path      string  true  "Notification ID"
// @Success      200
// @Router       /giveaway/notifications/{notificationId}/read [post]
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

// GetUnreadCount godoc
// @Summary      Get unread notification count
// @Description  Returns the number of unread notifications for the user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]int
// @Router       /giveaway/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	count, err := h.nService.CountUnreadNotifications(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"unread_count": count})
}
