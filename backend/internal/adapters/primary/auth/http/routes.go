package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/dto"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/utils"
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
		r.Post("/password/reset/request", rt.RequestPasswordResetHandler)
		r.Post("/password/reset/confirm", rt.ConfirmPasswordResetHandler)
	})
}

// Register protected auth routes (must be inside AuthMiddleware group)
func (rt *Routes) RegisterProtected(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/logout", rt.LogoutHandler)

		// Provider Registration
		r.Route("/providers/application", func(r chi.Router) {
			r.Post("/start", rt.StartProviderApplicationHandler)
			r.Post("/document-url", rt.GetDocumentUploadURLHandler)
			r.Post("/document/confirm", rt.ConfirmDocumentUploadHandler)
			r.Post("/submit", rt.SubmitProviderApplicationHandler)
			r.Get("/status", rt.GetProviderApplicationStatusHandler)
		})

		// Admin Routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireRole("ADMIN", rt.logger))
			r.Post("/providers/{id}/approve", rt.ApproveProviderApplicationHandler)
			r.Post("/providers/{id}/reject", rt.RejectProviderApplicationHandler)
		})
	})
}

// Register public student registration routes
func (rt *Routes) RegisterRegistration(router chi.Router) {
	router.Route("/auth/student", func(r chi.Router) {
		r.Post("/request-activation", rt.RequestActivationHandler)
		r.Get("/activate", rt.VerifyActivationHandler)
		r.Post("/complete", rt.CompleteStudentRegistrationHandler)
		r.Post("/resend", rt.ResendActivationHandler)
		r.Get("/status", rt.GetRegistrationStatusHandler)
		r.Get("/check-name", rt.CheckDisplayNameHandler)
	})

	router.Route("/auth/providers", func(r chi.Router) {
		r.Post("/request-activation", rt.RequestProviderActivationHandler)
		r.Get("/activate", rt.VerifyProviderActivationHandler)
		r.Post("/complete", rt.CompleteProviderRegistrationHandler)
		r.Post("/resend", rt.ResendProviderActivationHandler)
		r.Get("/status", rt.GetProviderRegistrationStatusHandler)
	})
}

// LoginHandler handles user authentication
// @Summary Login
// @Description Authenticate user via Keycloak and return access & refresh tokens
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
		ExpiresIn:    result.ExpiresIn,
		User: dto.LoginUser{
			ID:    result.User.ID,
			Email: result.User.Email,
			Role:  result.User.Role,
		},
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
// @Accept json
// @Produce json
// @Param payload body dto.RefreshRequest true "Refresh payload"
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/logout [post]
func (rt *Routes) LogoutHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.RefreshRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.Logout(req.Context(), body.RefreshToken); err != nil {
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

	if !utils.IsValidEmail(body.Email) {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("email", "invalid email address format"), rt.logger)
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

	wantsHTML := strings.Contains(req.Header.Get("Accept"), "text/html")

	if err := rt.handler.VerifyStudentActivation(req.Context(), token); err != nil {
		if wantsHTML {
			redirectURL := fmt.Sprintf("http://dealna-web-hosting.s3-website-us-east-1.amazonaws.com/student/activate/index.html?error=true")
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	if wantsHTML {
		redirectURL := fmt.Sprintf("http://dealna-web-hosting.s3-website-us-east-1.amazonaws.com/student/activate/index.html?token=%s", token)
		http.Redirect(w, req, redirectURL, http.StatusFound)
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

// CheckDisplayNameHandler handles checking if a display name is available
// @Summary Check Display Name Availability
// @Description Check if a display name is available for registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param name query string true "Display Name"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/check-name [get]
func (rt *Routes) CheckDisplayNameHandler(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if name == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("name", "name is required"), rt.logger)
		return
	}

	if err := rt.handler.CheckDisplayName(req.Context(), name); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]bool{"available": true})
}

// @Summary Get registration status
// @Description Check if student email is verified and registration can be completed
// @Tags Auth
// @Produce json
// @Param email query string true "Student email"
// @Success 200 {object} dto.RegistrationStatusResponse
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/student/status [get]
func (rt *Routes) GetRegistrationStatusHandler(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if email == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("email", "required"), rt.logger)
		return
	}

	pre, err := rt.handler.GetStudentRegistrationStatus(req.Context(), email)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, dto.RegistrationStatusResponse{
		Email:                   pre.Email,
		IsVerified:              pre.VerifiedAt != nil,
		IsUsed:                  pre.UsedAt != nil,
		ExpiresAt:               pre.ExpiresAt,
		VerifiedAt:              pre.VerifiedAt,
		CanCompleteRegistration: pre.VerifiedAt != nil && pre.UsedAt == nil && time.Now().Before(pre.ExpiresAt),
	})
}

// @Summary Get provider registration status
// @Description Get provider registration status by email
// @Tags Auth
// @Accept json
// @Produce json
// @Param email query string true "Provider email"
// @Success 200 {object} dto.RegistrationStatusResponse
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/status [get]
func (rt *Routes) GetProviderRegistrationStatusHandler(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if email == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("email", "required"), rt.logger)
		return
	}

	pre, err := rt.handler.GetProviderRegistrationStatus(req.Context(), email)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, dto.RegistrationStatusResponse{
		Email:                   pre.Email,
		IsVerified:              pre.VerifiedAt != nil,
		IsUsed:                  pre.UsedAt != nil,
		ExpiresAt:               pre.ExpiresAt,
		VerifiedAt:              pre.VerifiedAt,
		CanCompleteRegistration: pre.VerifiedAt != nil && pre.UsedAt == nil && time.Now().Before(pre.ExpiresAt),
	})
}

// @Summary Request provider activation
// @Description Start provider registration process and send verification email
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RequestActivationRequest true "Registration data"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/request-activation [post]
func (rt *Routes) RequestProviderActivationHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.RequestActivationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if !utils.IsValidEmail(body.Email) {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("email", "invalid email address format"), rt.logger)
		return
	}

	if err := rt.handler.RequestProviderActivation(req.Context(), body.Email); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Verify provider activation token
// @Description Check activation link validity
// @Tags Auth
// @Produce json
// @Param token query string true "Activation token"
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/activate [get]
func (rt *Routes) VerifyProviderActivationHandler(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")
	if token == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("token", "required"), rt.logger)
		return
	}

	wantsHTML := strings.Contains(req.Header.Get("Accept"), "text/html")

	if err := rt.handler.VerifyProviderActivation(req.Context(), token); err != nil {
		if wantsHTML {
			redirectURL := fmt.Sprintf("http://dealna-web-hosting.s3-website-us-east-1.amazonaws.com/provider/activate/index.html?error=true")
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	if wantsHTML {
		redirectURL := fmt.Sprintf("http://dealna-web-hosting.s3-website-us-east-1.amazonaws.com/provider/activate/index.html?token=%s", token)
		http.Redirect(w, req, redirectURL, http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Complete provider registration
// @Description Finalize provider account
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.CompleteProviderRegistrationRequest true "Registration data"
// @Success 204
// @Failure 401 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/complete [post]
func (rt *Routes) CompleteProviderRegistrationHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.CompleteProviderRegistrationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	result, err := rt.handler.CompleteProviderRegistration(req.Context(), body.Email, body.Password)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User: dto.LoginUser{
			ID:    result.User.ID,
			Email: result.User.Email,
			Role:  result.User.Role,
		},
	})
}

// @Summary Resend activation link
// @Description Resend activation email for provider
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RequestActivationRequest true "Email"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/resend [post]
func (rt *Routes) ResendProviderActivationHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.RequestActivationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.ResendProviderActivation(req.Context(), body.Email); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Start provider application
// @Description Creates a DRAFT application for the provider
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body dto.StartProviderApplicationRequest true "Application details"
// @Success 200
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/application/start [post]
func (rt *Routes) StartProviderApplicationHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.StartProviderApplicationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	userID := middleware.UserIDFromContext(req.Context())

	app, err := rt.handler.StartProviderApplication(
		req.Context(),
		userID,
		body.UniversityID,
		body.BusinessName,
		body.PhoneNumber,
		body.BusinessType,
		body.Address,
	)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, app)
}

// @Summary Get document upload URL
// @Description Get an S3 presigned URL to upload a document
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body dto.GetDocumentUploadURLRequest true "Document details"
// @Success 200
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/application/document-url [post]
func (rt *Routes) GetDocumentUploadURLHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.GetDocumentUploadURLRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	userID := middleware.UserIDFromContext(req.Context())

	url, objectKey, err := rt.handler.GetDocumentUploadURL(
		req.Context(),
		userID,
		body.DocumentType,
		body.OriginalFilename,
		body.ContentType,
	)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{
		"url":        url,
		"object_key": objectKey,
	})
}

// @Summary Confirm document upload
// @Description Confirms that a document was uploaded successfully
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body dto.ConfirmDocumentUploadRequest true "Document details"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/application/document/confirm [post]
func (rt *Routes) ConfirmDocumentUploadHandler(w http.ResponseWriter, req *http.Request) {
	var body dto.ConfirmDocumentUploadRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	userID := middleware.UserIDFromContext(req.Context())

	err := rt.handler.ConfirmDocumentUpload(
		req.Context(),
		userID,
		body.ObjectKey,
		body.DocumentType,
		body.OriginalFilename,
		body.ContentType,
		body.SizeBytes,
	)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Submit provider application
// @Description Submits the application for review
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/application/submit [post]
func (rt *Routes) SubmitProviderApplicationHandler(w http.ResponseWriter, req *http.Request) {
	userID := middleware.UserIDFromContext(req.Context())

	err := rt.handler.SubmitProviderApplication(req.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Approve provider application
// @Description Approves a provider application and upgrades the applicant to a PROVIDER role
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Applicant ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/admin/providers/{id}/approve [post]
// @Summary Get provider application status
// @Description Gets the current status of the logged in user's provider application
// @Tags Provider Registration
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.ProviderApplication
// @Failure 401 {object} middleware.ErrorFrame
// @Failure 404 {object} middleware.ErrorFrame
// @Router /api/v1/auth/providers/application/status [get]
func (rt *Routes) GetProviderApplicationStatusHandler(w http.ResponseWriter, req *http.Request) {
	userID := middleware.UserIDFromContext(req.Context())

	app, err := rt.handler.GetProviderApplicationStatus(req.Context(), userID)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(app); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewInternalError(err), rt.logger)
		return
	}
}

func (rt *Routes) ApproveProviderApplicationHandler(w http.ResponseWriter, req *http.Request) {
	applicantID := chi.URLParam(req, "id")
	if applicantID == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("id", "applicant id is required"), rt.logger)
		return
	}
	adminID := middleware.UserIDFromContext(req.Context())

	err := rt.handler.ApproveProviderApplication(req.Context(), adminID, applicantID)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Reject provider application
// @Description Rejects a provider application and provides a comment
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Applicant ID"
// @Param payload body dto.RejectProviderApplicationRequest true "Rejection Details"
// @Success 204
// @Failure 400 {object} middleware.ErrorFrame
// @Router /api/v1/auth/admin/providers/{id}/reject [post]
func (rt *Routes) RejectProviderApplicationHandler(w http.ResponseWriter, req *http.Request) {
	applicantID := chi.URLParam(req, "id")
	if applicantID == "" {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("id", "applicant id is required"), rt.logger)
		return
	}

	var body dto.RejectProviderApplicationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	adminID := middleware.UserIDFromContext(req.Context())

	err := rt.handler.RejectProviderApplication(req.Context(), adminID, applicantID, body.Comment)
	if err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *Routes) RequestPasswordResetHandler(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}
	body.Email = strings.TrimSpace(body.Email)

	if err := rt.handler.RequestPasswordReset(req.Context(), body.Email); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func (rt *Routes) ConfirmPasswordResetHandler(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), middleware.NewValidationError("body", "invalid json"), rt.logger)
		return
	}

	if err := rt.handler.ConfirmPasswordReset(req.Context(), body.Email, body.Token, body.Password); err != nil {
		middleware.WriteErrorResponse(w, req.Context(), err, rt.logger)
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}
