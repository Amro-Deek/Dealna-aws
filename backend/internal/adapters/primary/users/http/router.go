package http

import (
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/dto"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *users.Handler
}

func NewRoutes(handler *users.Handler) *Routes {
	return &Routes{handler: handler}
}

func (rt *Routes) Register(router chi.Router) {
	router.Route("/users", func(r chi.Router) {
		r.Get("/me", rt.getMe)
	})
}
// GetMe godoc
// @Summary Get current user
// @Description Returns authenticated user profile -test
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.MeResponse
// @Failure 401 {object} utils.APIResponse
// @Router /api/v1/users/me [get]
func (rt *Routes) getMe(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	userID := middleware.UserIDFromContext(ctx)

	user, err := rt.handler.GetMe(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp := dto.MeResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.WriteJSON(w, http.StatusOK, true, "OK", resp, nil)
}
