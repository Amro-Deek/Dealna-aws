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
