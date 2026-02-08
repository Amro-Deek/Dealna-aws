// @title           Dealna API
// @version         1.0
// @description     Dealna backend API
// @contact.name    Amro Deek
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @BasePath        /api/v1

package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	// Config & Infra
	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"

	// Primary adapters (Handlers + Routes)
	authHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	authHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/http"
	userHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	// Secondary adapters
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence"

	// Core services
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"

	// Swagger
	_ "github.com/Amro-Deek/Dealna-aws/backend/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	// =========================
	// Load config
	// =========================
	cfg := config.Load()

	// =========================
	// Database
	// =========================
	db, err := database.Connect(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer db.Close()

	// =========================
	// Repositories
	// =========================
	repoFactory := persistence.NewRepositoryFactory(db)
	userRepo := repoFactory.User()

	// =========================
	// Secondary adapters (Auth)
	// =========================
	hasher := auth.NewBcryptHasher()
	jwtProvider := auth.NewJWTProvider(cfg.JWTSecret)

	// =========================
	// Logger (Middleware)
	// =========================
	appLogger := middleware.NewStdLogger()

	// =========================
	// Core services
	// =========================
	authService := services.NewAuthService(
		userRepo,
		hasher,
		jwtProvider,
	)

	userService := services.NewUserService(userRepo)

	// =========================
	// Handlers (Primary)
	// =========================
	authH := authHandler.NewHandler(authService)
	userH := userHandler.NewHandler(userService)

	// =========================
	// Routes (Primary HTTP)
	// =========================
	authRoutes := authHTTP.NewRoutes(authH, appLogger)
	userRoutes := userHTTP.NewRoutes(userH)

	// =========================
	// Router
	// =========================
	r := chi.NewRouter()

	// Global middlewares
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// =========================
	// API v1
	// =========================
	r.Route("/api/v1", func(r chi.Router) {

		// Public routes
		authRoutes.Register(r)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtProvider, appLogger))
			userRoutes.Register(r)
		})
	})

	log.Printf("🚀 Dealna server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
