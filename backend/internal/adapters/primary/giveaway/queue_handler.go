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
// @Router       /api/v1/giveaway/queue/{itemId}/join [post]
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
// @Router       /api/v1/giveaway/queue/{itemId}/leave [post]
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
// @Router       /api/v1/giveaway/queue/{itemId}/position/{entryId} [get]
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

// AcceptTurn godoc
// @Summary      Owner accepts turn
// @Description  Owner accepts the current reserved user's turn
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId   path  string  true  "Item ID"
// @Param        entryId  path  string  true  "Queue Entry ID"
// @Success      200      "OK"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/{itemId}/entries/{entryId}/accept [post]
func (h *QueueHandler) AcceptTurn(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	entryID := chi.URLParam(r, "entryId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.qService.AcceptTurn(r.Context(), itemID, entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// RejectTurn godoc
// @Summary      Owner rejects turn
// @Description  Owner rejects the current reserved user's turn and skips to next
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId   path  string  true  "Item ID"
// @Param        entryId  path  string  true  "Queue Entry ID"
// @Success      200      "OK"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/{itemId}/entries/{entryId}/reject [post]
func (h *QueueHandler) RejectTurn(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	entryID := chi.URLParam(r, "entryId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.qService.RejectTurn(r.Context(), itemID, entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// InitiateHandoff godoc
// @Summary      Owner initiates handoff
// @Description  Owner initiates the item handoff to the confirmed user
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId   path  string  true  "Item ID"
// @Param        entryId  path  string  true  "Queue Entry ID"
// @Success      200      "OK"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/{itemId}/entries/{entryId}/handoff [post]
func (h *QueueHandler) InitiateHandoff(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	entryID := chi.URLParam(r, "entryId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.qService.InitiateHandoff(r.Context(), itemID, entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ConfirmHandoff godoc
// @Summary      Receiver confirms handoff
// @Description  Receiver confirms they received the item
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId   path  string  true  "Item ID"
// @Param        entryId  path  string  true  "Queue Entry ID"
// @Success      200      "OK"
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/{itemId}/entries/{entryId}/complete [post]
func (h *QueueHandler) ConfirmHandoff(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	entryID := chi.URLParam(r, "entryId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.qService.ConfirmHandoff(r.Context(), itemID, entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetMyQueues godoc
// @Summary      Get user's queue entries
// @Description  Get all active queue entries for the authenticated user
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Success      200      {array}   domain.QueuePosition
// @Failure      401      {string}  string  "unauthorized"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/me [get]
func (h *QueueHandler) GetMyQueues(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	entries, err := h.qService.GetQueueEntriesByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

// GetQueueEntries godoc
// @Summary      Get queue entries for an item
// @Description  Get all active queue entries for a specific item (Owner only)
// @Tags         Giveaway Queue
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Success      200      {array}   domain.QueueEntry
// @Failure      401      {string}  string  "unauthorized"
// @Failure      403      {string}  string  "forbidden"
// @Failure      500      {string}  string  "internal error"
// @Router       /api/v1/giveaway/queue/{itemId}/entries [get]
func (h *QueueHandler) GetQueueEntries(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	entries, err := h.qService.GetQueueEntriesByItem(r.Context(), itemID, userID)
	if err != nil {
		if err.Error() == "only the item owner can view its queue entries" {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(entries)
}
