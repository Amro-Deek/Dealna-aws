package httpadapter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"

	adminHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/admin/http"
	authHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/http"
	chatHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/chat/http"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/giveaway"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/items"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/marketplace"
	profileHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile/http"
	ratingsHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/ratings/http"
	reportsHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/reports/http"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/social"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	cfg *config.Config,
	authRoutes *authHTTP.Routes,
	userRoutes *userHTTP.Routes,
	profileRoutes *profileHTTP.Routes,
	itemRoutes *items.Routes,
	giveawayRoutes *giveaway.Routes,
	marketplaceRoutes *marketplace.Routes,
	socialRoutes *social.Routes,
	chatRoutes *chatHTTP.Routes,
	ratingRoutes *ratingsHTTP.Routes,
	adminHandler *adminHTTP.AdminHandler,
	reportRoutes *reportsHTTP.Routes,
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
	// CORS Config
	// =========================
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// =========================
	// Web Fallback Activation
	// =========================
	r.Get("/provider/activate", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "internal/adapters/primary/auth/http/templates/activate.html")
	})

	// =========================
	// App Links / Universal Links
	// =========================
	fs := http.FileServer(http.Dir("internal/adapters/primary/static"))
	r.Handle("/.well-known/*", http.StripPrefix("/", fs))

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
			r.Get("/student/status", authRoutes.GetRegistrationStatusHandler)
			r.Get("/student/check-name", authRoutes.CheckDisplayNameHandler)

			// =============================
			// Provider Registration Flow (Public)
			// =============================
			r.Post("/providers/request-activation", authRoutes.RequestProviderActivationHandler)
			r.Get("/providers/activate", authRoutes.VerifyProviderActivationHandler)
			r.Post("/providers/complete", authRoutes.CompleteProviderRegistrationHandler)
			r.Post("/providers/resend", authRoutes.ResendProviderActivationHandler)
			r.Get("/providers/status", authRoutes.GetProviderRegistrationStatusHandler)

			// ---------
			// Protected
			// ---------
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(authProvider, logger))
				r.Post("/logout", authRoutes.LogoutHandler)

				// =============================
				// Provider Registration Flow (Protected)
				// =============================
				r.Route("/providers/application", func(r chi.Router) {
					r.Post("/start", authRoutes.StartProviderApplicationHandler)
					r.Post("/document-url", authRoutes.GetDocumentUploadURLHandler)
					r.Post("/document/confirm", authRoutes.ConfirmDocumentUploadHandler)
					r.Post("/submit", authRoutes.SubmitProviderApplicationHandler)
					r.Get("/status", authRoutes.GetProviderApplicationStatusHandler)
				})

			})
		})

		// =============================
		// Admin Routes (Protected)
		// =============================
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authProvider, logger))
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole("ADMIN", logger))

				// Dashboard APIs
				adminHandler.Register(r)

				// Provider actions
				r.Post("/providers/{id}/approve", authRoutes.ApproveProviderApplicationHandler)
				r.Post("/providers/{id}/reject", authRoutes.RejectProviderApplicationHandler)
			})
		})

		// =====================
		// Public item routes (no auth)
		// =====================
		itemRoutes.RegisterPublic(r)

		// =====================
		// Protected user routes
		// =====================
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authProvider, logger))
			userRoutes.Register(r)
			profileRoutes.Register(r)
			itemRoutes.RegisterProtected(r)
			giveawayRoutes.Register(r)
			marketplaceRoutes.Register(r)
			socialRoutes.Register(r)
			chatRoutes.RegisterProtected(r)
			ratingRoutes.Register(r)
			reportRoutes.RegisterProtected(r)
		})
	})

	return r
}
