package giveaway

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"
	"github.com/go-chi/chi/v5"
)

type TransactionHandler struct {
	tService *services.TransactionService
}

func NewTransactionHandler(tService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{tService: tService}
}

func (h *TransactionHandler) ConfirmSeller(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "transactionId")
	err := h.tService.ConfirmSeller(r.Context(), txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Confirmed"})
}
