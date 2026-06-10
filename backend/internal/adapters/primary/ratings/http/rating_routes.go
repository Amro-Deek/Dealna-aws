package http

import (
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *RatingHandler
}

func NewRoutes(h *RatingHandler) *Routes {
	return &Routes{handler: h}
}

func (r *Routes) Register(router chi.Router) {
	router.Post("/transactions/{transactionId}/rate", r.handler.CreateRating)
	router.Get("/users/me/pending-ratings", r.handler.GetPendingRatings)
	router.Get("/users/{userId}/ratings", r.handler.GetUserReviews)
}
