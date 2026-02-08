package http

import (
	"encoding/json"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/dto"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	handler *auth.Handler
	logger  middleware.StructuredLoggerInterface
}

func NewRoutes(handler *auth.Handler, logger middleware.StructuredLoggerInterface) *Routes {
	return &Routes{
		handler: handler,
		logger:  logger,
	}
}

// Login godoc
// @Summary Login
// @Description Authenticate user and return JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/login [post]
func (rt *Routes) Register(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", func(w http.ResponseWriter, req *http.Request) {

			var body dto.LoginRequest
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				middleware.WriteErrorResponse(
					w,
					req.Context(),
					middleware.NewValidationError("body", "invalid json"),
					rt.logger,
				)
				return
			}

			result, err := rt.handler.Login(req.Context(), body.Email, body.Password)
			if err != nil {
				middleware.WriteErrorResponse(
					w,
					req.Context(),
					err,
					rt.logger,
				)
				return
			}

			middleware.WriteJSONResponse(
				w,
				http.StatusOK,
				dto.LoginResponse{
					AccessToken: result.AccessToken,
				},
			)
		})
	})
}
