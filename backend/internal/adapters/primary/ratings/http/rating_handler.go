package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type RatingHandler struct {
	ratingService *services.RatingService
}

func NewRatingHandler(rs *services.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: rs}
}

type CreateRatingRequest struct {
	Stars   int    `json:"stars"`
	Comment string `json:"comment"`
}

func (h *RatingHandler) CreateRating(w http.ResponseWriter, r *http.Request) {
	// In a real app, raterID comes from JWT middleware ctx.
	// For this snippet, we'll assume we parse it from a header or context.
	// Assuming RaterID is extracted from context:
	raterIDStr := middleware.UserIDFromContext(r.Context())
	raterID, err := uuid.Parse(raterIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	txIDStr := chi.URLParam(r, "transactionId")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var req CreateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cmd := domain.CreateRatingCommand{
		RaterID:       raterID,
		TransactionID: txID,
		Stars:         req.Stars,
		Comment:       req.Comment,
	}

	rating, err := h.ratingService.CreateRating(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrRatingNotAllowed || err == domain.ErrAlreadyRated {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rating)
}

func (h *RatingHandler) GetPendingRatings(w http.ResponseWriter, r *http.Request) {
	buyerIDStr := middleware.UserIDFromContext(r.Context())
	buyerID, err := uuid.Parse(buyerIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pending, err := h.ratingService.GetPendingRatings(r.Context(), buyerID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if pending == nil {
		pending = []domain.PendingRating{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}
