package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/dto"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type Routes struct {
	handler *auth.Handler
	logger  middleware.StructuredLoggerInterface
}

func NewRoutes(
	handler *auth.Handler,
	logger middleware.StructuredLoggerInterface,
) *Routes {
	return &Routes{
		handler: handler,
		logger:  logger,
	}
}
// Register public auth routes (no auth middleware)
func (rt *Routes) RegisterPublic(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", rt.LoginHandler)
		r.Post("/refresh", rt.RefreshHandler)
	})
}

// Register protected auth routes (must be inside AuthMiddleware group)
func (rt *Routes) RegisterProtected(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/logout", rt.LogoutHandler)
	})
}


// LoginHandler handles user authentication
// @Summary Login
// @Description Authenticate user and return access & refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} middleware.ErrorFrame
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/login [post]
func (rt *Routes) LoginHandler(w http.ResponseWriter, req *http.Request) {
    var body dto.LoginRequest
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
        return
    }

    result, err := rt.handler.Login(req.Context(), body.Email, body.Password)
    if err != nil {
        middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
        return
    }

    middleware.WriteJSONResponse(w, http.StatusOK, dto.LoginResponse{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
    })
}

// RefreshHandler handles token refresh
// @Summary Refresh token
// @Description Refresh access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RefreshRequest true "Refresh payload"
// @Success 200 {object} dto.RefreshResponse
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/refresh [post]
func (rt *Routes) RefreshHandler(w http.ResponseWriter, req *http.Request) {
    var body dto.RefreshRequest
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
        return
    }

    result, err := rt.handler.Refresh(req.Context(), body.RefreshToken)
    if err != nil {
        middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
        return
    }

    middleware.WriteJSONResponse(w, http.StatusOK, dto.RefreshResponse{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
    })
}

// LogoutHandler handles session revocation
// @Summary Logout
// @Description Revoke current session
// @Tags Auth
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/logout [post]
func (rt *Routes) LogoutHandler(w http.ResponseWriter, req *http.Request) {
    jti := middleware.JTIFromContext(req.Context())
    if jti == "" {
        middleware.WriteErrorResponse(w, req.Context(), middleware.NewUnauthorizedError("unauthorized"), rt.logger)
        return
    }

    if err := rt.handler.Logout(req.Context(), jti); err != nil {
        middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}


// @Summary Request activation link
// @Description Send activation email to university address
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RequestActivationRequest true "Email"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/request-activation [post]
func (rt *Routes) RequestActivationHandler(w http.ResponseWriter, req *http.Request) {

	var body dto.RequestActivationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.RequestStudentActivation(req.Context(), body.Email); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Verify activation token
// @Description Check activation link validity
// @Tags Auth
// @Produce json
// @Param token query string true "Activation token"
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/activate [get]
func (rt *Routes) VerifyActivationHandler(w http.ResponseWriter, req *http.Request) {

	token := req.URL.Query().Get("token")
	if token == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("token", "required"), rt.logger)
		return
	}

	if err := rt.handler.VerifyStudentActivation(req.Context(), token); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Complete registration
// @Description Finalize student account
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.CompleteStudentRegistrationRequest true "Registration data"
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/complete [post]
func (rt *Routes) CompleteStudentRegistrationHandler(w http.ResponseWriter, req *http.Request) {

	var body dto.CompleteStudentRegistrationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.CompleteStudentRegistration(
		req.Context(),
		body.Email,
		body.DisplayName,
		body.Password,
		body.Major,
		body.AcademicYear,
	); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Resend activation link
// @Description Resend activation email
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.ResendActivationRequest true "Email"
// @Success 204
// @Failure 429 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/resend [post]
func (rt *Routes) ResendActivationHandler(w http.ResponseWriter, req *http.Request) {

	var body dto.ResendActivationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.ResendStudentActivation(req.Context(), body.Email); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
