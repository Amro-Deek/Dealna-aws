package http

import (
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *ReportHandler
}

func NewRoutes(h *ReportHandler) *Routes {
	return &Routes{handler: h}
}

func (r *Routes) RegisterProtected(router chi.Router) {
	router.Post("/reports", r.handler.CreateReport)
	router.Post("/reports/", r.handler.CreateReport)
}
