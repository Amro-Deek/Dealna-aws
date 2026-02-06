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

	"github.com/Amro-Deek/Dealna-aws/internal/config"
	"github.com/Amro-Deek/Dealna-aws/internal/database"

	authHandler "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/auth"
	authHTTP "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/auth/http"

	userHandler "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/users"
	userHTTP "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/users/http"

	"github.com/Amro-Deek/Dealna-aws/internal/adapters/secondary/persistence"
	"github.com/Amro-Deek/Dealna-aws/internal/core/services"
)

func main() {
	cfg := config.Load()

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

	// Repositories
	repoFactory := persistence.NewRepositoryFactory(db)
	userRepo := repoFactory.User()

	// Services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	// Handlers
	userH := userHandler.NewHandler(userService)
	authH := authHandler.NewHandler(authService)

	// Routes
	userRoutes := userHTTP.NewRoutes(userH)
	authRoutes := authHTTP.NewRoutes(authH)

	// Router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		authRoutes.Register(r)
		userRoutes.Register(r, cfg.JWTSecret)
	})

	log.Printf("🚀 Dealna server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
