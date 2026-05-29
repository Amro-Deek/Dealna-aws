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
// @Router       /api/v1/purchases/items/{itemId}/request [post]
func (h *PurchaseHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	buyerID := middleware.UserIDFromContext(r.Context())
	if buyerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if middleware.RoleFromContext(r.Context()) == "PROVIDER" {
		http.Error(w, "Providers cannot make purchases", http.StatusForbidden)
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
// @Router       /api/v1/purchases/items/{itemId}/requests [get]
func (h *PurchaseHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqs, err := h.pService.ListRequests(r.Context(), itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(reqs)
}

// AcceptRequest godoc
// @Summary      Accept a purchase request
// @Description  Allows the seller to accept a specific purchase request
// @Tags         Purchases
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Param        requestId  path  string  true  "Request ID"
// @Success      200     "OK"
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /api/v1/purchases/items/{itemId}/requests/{requestId}/accept [post]
func (h *PurchaseHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqID := chi.URLParam(r, "requestId")
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.pService.AcceptRequest(r.Context(), reqID, itemID, callerID)
	if err != nil {
		if err.Error() == "only the item owner can accept purchase requests" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// RejectRequest godoc
// @Summary      Reject a purchase request
// @Description  Allows the seller to reject a specific purchase request
// @Tags         Purchases
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Param        requestId  path  string  true  "Request ID"
// @Success      200     "OK"
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /api/v1/purchases/items/{itemId}/requests/{requestId}/reject [post]
func (h *PurchaseHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqID := chi.URLParam(r, "requestId")
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.pService.RejectRequest(r.Context(), reqID, itemID, callerID)
	if err != nil {
		if err.Error() == "only the item owner can reject purchase requests" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// CancelRequest godoc
// @Summary      Cancel a purchase request
// @Description  Allows the buyer to cancel their own purchase request
// @Tags         Purchases
// @Security     BearerAuth
// @Param        itemId  path  string  true  "Item ID"
// @Param        requestId  path  string  true  "Request ID"
// @Success      200     "OK"
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /api/v1/purchases/items/{itemId}/requests/{requestId}/cancel [post]
func (h *PurchaseHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	reqID := chi.URLParam(r, "requestId")
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.pService.CancelRequest(r.Context(), reqID, itemID, callerID)
	if err != nil {
		if err.Error() == "only the buyer can cancel their purchase request" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetMyRequests godoc
// @Summary      Get my purchase requests
// @Description  Returns a list of all purchase requests made by the authenticated user
// @Tags         Purchases
// @Security     BearerAuth
// @Success      200     {array}   domain.PurchaseRequest
// @Failure      401     {string}  string  "unauthorized"
// @Failure      500     {string}  string  "internal error"
// @Router       /api/v1/purchases/me [get]
func (h *PurchaseHandler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	reqs, err := h.pService.GetMyRequests(r.Context(), callerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(reqs)
}
