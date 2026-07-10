package http

import (
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *ChatHandler
	logger  middleware.StructuredLoggerInterface
}

func NewRoutes(h *ChatHandler, logger middleware.StructuredLoggerInterface) *Routes {
	return &Routes{handler: h, logger: logger}
}

func (r *Routes) RegisterProtected(router chi.Router) {
	router.Route("/chat", func(rg chi.Router) {
		rg.Group(func(rg chi.Router) {
			rg.Use(middleware.ForbidRole("LIMITED_STUDENT", r.logger))
			rg.Get("/token", r.handler.GetChatToken)
			rg.Post("/notify", r.handler.SendChatNotification)
		})
	})
}

