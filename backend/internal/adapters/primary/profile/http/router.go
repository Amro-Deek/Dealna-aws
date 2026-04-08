package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile"
)

type Routes struct {
	handler *profile.ProfileHandler
}

func NewRoutes(h *profile.ProfileHandler) *Routes {
	return &Routes{handler: h}
}

func (r *Routes) Register(router chi.Router) {
	router.Route("/profile", func(router chi.Router) {
		router.Get("/", r.handler.GetMyProfile)
		router.Put("/", r.handler.UpdateProfile)
		router.Put("/student", r.handler.UpdateStudent)
		router.Post("/picture/upload-url", r.handler.GenerateUploadURL)
	})
}
