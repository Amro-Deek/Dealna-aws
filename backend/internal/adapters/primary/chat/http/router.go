package http

import (
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *ChatHandler
}

func NewRoutes(h *ChatHandler) *Routes {
	return &Routes{handler: h}
}

func (r *Routes) RegisterProtected(router chi.Router) {
	router.Route("/chat", func(router chi.Router) {
		router.Get("/token", r.handler.GetChatToken)
	})
}
