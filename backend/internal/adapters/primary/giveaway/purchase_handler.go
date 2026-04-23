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

// SendPurchaseRequest godoc
// @Summary      Send a purchase request
// @Description  Creates a new purchase request from the authenticated user to buy a specific item
// @Tags         Purchases
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Success      200     {object}  domain.PurchaseRequest
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /purchases/items/{itemId}/request [post]
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

// ListPurchaseRequests godoc
// @Summary      List purchase requests for an item
// @Description  Returns a list of all purchase requests made for a specific item
// @Tags         Purchases
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Success      200     {array}   domain.PurchaseRequest
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /purchases/items/{itemId}/requests [get]
func (h *PurchaseHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqs, err := h.pService.ListRequests(r.Context(), itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(reqs)
}
