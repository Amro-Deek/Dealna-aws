package items

import (
	"github.com/go-chi/chi/v5"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type Routes struct {
	handler *ItemHandler
	logger  middleware.StructuredLoggerInterface
}

func NewRoutes(handler *ItemHandler, logger middleware.StructuredLoggerInterface) *Routes {
	return &Routes{handler: handler, logger: logger}
}

// RegisterPublic registers routes that do NOT require authentication.
func (rt *Routes) RegisterPublic(r chi.Router) {
	r.Get("/categories", rt.handler.GetCategories)
}

// RegisterProtected registers routes that require a valid JWT.
func (rt *Routes) RegisterProtected(r chi.Router) {
	r.Route("/items", func(r chi.Router) {
		r.Get("/feed", rt.handler.GetFeed)
		r.Get("/search", rt.handler.SearchItems)
		r.Get("/my", rt.handler.GetMyItems)
		r.Get("/{id}/similar", rt.handler.GetSimilarItems)
		r.Get("/{id}", rt.handler.GetItemDetail)

		// Restricted to non-limited students
		r.Group(func(r chi.Router) {
			r.Use(middleware.ForbidRole("LIMITED_STUDENT", rt.logger))
			r.Post("/", rt.handler.CreateItem)
			r.Post("/picture/upload-url", rt.handler.GenerateUploadURL)
			r.Patch("/{id}/status", rt.handler.UpdateStatus)
			r.Delete("/{id}", rt.handler.DeleteItem)
		})
	})

	r.Route("/users/{id}/items", func(r chi.Router) {
		r.Get("/", rt.handler.GetUserItems)
	})
}
