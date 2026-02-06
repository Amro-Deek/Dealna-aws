package http

import (
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/users"
	"github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/users/dto"
	"github.com/Amro-Deek/Dealna-aws/internal/middleware"
	"github.com/Amro-Deek/Dealna-aws/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *users.Handler
}

func NewRoutes(handler *users.Handler) *Routes {
	return &Routes{handler: handler}
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
func (rt *Routes) Register(router chi.Router, jwtSecret string) {
	router.Route("/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtSecret))

		r.Get("/me", func(w http.ResponseWriter, req *http.Request) {

			userID := middleware.UserIDFromContext(req.Context())
			if userID == "" {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, nil)
				return
			}

			user, err := rt.handler.GetMe(req.Context(), userID)
			if err != nil {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, nil)
				return
			}

			resp := dto.MeResponse{
				ID:    user.ID,
				Email: user.Email,
				Role:  user.Role,
			}

			utils.WriteJSON(w, http.StatusOK, true, "OK", resp, nil)
		})
	})
}
