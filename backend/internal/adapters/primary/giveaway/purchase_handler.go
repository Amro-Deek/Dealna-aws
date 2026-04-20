package giveaway

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type PurchaseHandler struct {
	pService *services.PurchaseService
}

func NewPurchaseHandler(pService *services.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{pService: pService}
}

func (h *PurchaseHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	buyerID := middleware.UserIDFromContext(r.Context())
	if buyerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	req, err := h.pService.SendRequest(r.Context(), itemID, buyerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func (h *PurchaseHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqs, err := h.pService.ListRequests(r.Context(), itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(reqs)
}
