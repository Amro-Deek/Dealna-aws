package httpadapter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"

	authHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/http"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	cfg *config.Config,
	authRoutes *authHTTP.Routes,
	userRoutes *userHTTP.Routes,
	authProvider ports.IAuthContextProvider,
	logger middleware.StructuredLoggerInterface,
) http.Handler {

	r := chi.NewRouter()

	// =========================
	// Global middlewares
	// =========================
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// =========================
	// Swagger
	// =========================
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// =========================
	// API v1
	// =========================
	r.Route("/api/v1", func(r chi.Router) {

		// =====================
		// Auth routes (ONE /auth)
		// =====================
		r.Route("/auth", func(r chi.Router) {

			// ---------
			// Public
			// ---------
			r.Post("/login", authRoutes.LoginHandler)
			r.Post("/refresh", authRoutes.RefreshHandler)

			// =============================
			// Student Registration Flow
			// =============================
			r.Post("/student/request-activation", authRoutes.RequestActivationHandler)
			r.Get("/student/activate", authRoutes.VerifyActivationHandler)
			r.Post("/student/complete", authRoutes.CompleteStudentRegistrationHandler)
			r.Post("/student/resend", authRoutes.ResendActivationHandler)

			// ---------
			// Protected
			// ---------
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(authProvider, logger))
				r.Post("/logout", authRoutes.LogoutHandler)
			})
		})

		// =====================
		// Protected user routes
		// =====================
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authProvider, logger))
			userRoutes.Register(r)
		})
	})

	return r
}
