package items

import (
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *ItemHandler
}

func NewRoutes(handler *ItemHandler) *Routes {
	return &Routes{handler: handler}
}

// RegisterPublic registers routes that do NOT require authentication.
func (rt *Routes) RegisterPublic(r chi.Router) {
	r.Get("/categories", rt.handler.GetCategories)
}

// RegisterProtected registers routes that require a valid JWT.
func (rt *Routes) RegisterProtected(r chi.Router) {
	r.Route("/items", func(r chi.Router) {
		r.Post("/", rt.handler.CreateItem)
		r.Get("/feed", rt.handler.GetFeed)
		r.Get("/search", rt.handler.SearchItems)
		r.Get("/my", rt.handler.GetMyItems)
		r.Post("/picture/upload-url", rt.handler.GenerateUploadURL)

		r.Get("/{id}", rt.handler.GetItemDetail)
		r.Patch("/{id}/status", rt.handler.UpdateStatus)
		r.Delete("/{id}", rt.handler.DeleteItem)
	})
}
