package giveaway

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type QueueHandler struct {
	qService *services.QueueService
}

func NewQueueHandler(qService *services.QueueService) *QueueHandler {
	return &QueueHandler{qService: qService}
}

// JoinQueue godoc
// @Summary      Join an item's giveaway queue
// @Description  Adds the authenticated user to the giveaway queue for the specified item
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Success      200     {object}  domain.QueueEntry
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /queue/{itemId}/join [post]
func (h *QueueHandler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	entry, err := h.qService.JoinQueue(r.Context(), itemID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entry)
}

// LeaveQueue godoc
// @Summary      Leave an item's giveaway queue
// @Description  Removes the authenticated user from the giveaway queue for the specified item
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Success      200     "OK"
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /queue/{itemId}/leave [delete]
func (h *QueueHandler) LeaveQueue(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.qService.LeaveQueue(r.Context(), itemID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetQueuePosition godoc
// @Summary      Get user position in queue
// @Description  Retrieves the current numerical position of a specific queue entry for an item
// @Tags         Giveaway Queue
// @Param        itemId   path  string  true  "Item ID"
// @Param        entryId  path  string  true  "Queue Entry ID"
// @Success      200      {object}  map[string]int "Returns { \"position\": 1 }"
// @Failure      500      {string}  string  "internal error"
// @Router       /queue/{itemId}/position/{entryId} [get]
func (h *QueueHandler) GetQueuePosition(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	entryID := chi.URLParam(r, "entryId")
	pos, err := h.qService.GetPosition(r.Context(), itemID, entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"position": pos})
}
