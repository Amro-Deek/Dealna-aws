package http

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/auth"
	"github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/auth/dto"
	"github.com/Amro-Deek/Dealna-aws/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *auth.Handler
}

func NewRoutes(handler *auth.Handler) *Routes {
	return &Routes{handler: handler}
}
// Login godoc
// @Summary Login
// @Description Authenticate user and return JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.LoginResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/login [post]
func (rt *Routes) Register(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", func(w http.ResponseWriter, req *http.Request) {

			var body dto.LoginRequest
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				utils.WriteJSON(w, http.StatusBadRequest, false, "Invalid body", nil, nil)
				return
			}

			result, err := rt.handler.Login(req.Context(), body.Email, body.Password)
			if err != nil {
				utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
				return
			}

			utils.WriteJSON(w, http.StatusOK, true, "OK", dto.LoginResponse{
				AccessToken: result.AccessToken,
			}, nil)
		})
	})
}
