package main

import (
	"log"
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"

	// Primary adapters
	httpadapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary"
	authHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	authHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/http"
	userHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	// Secondary adapters
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence"

	// Core services
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"

	// Swagger docs (USED via routes.go)
	_ "github.com/Amro-Deek/Dealna-aws/backend/docs"

	emailAdapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/email"
)

// @title           Dealna API
// @version         1.0
// @description     Dealna backend API
// @contact.name    Amro Deek
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @BasePath        /api/v1

func main() {

	// =========================
	// Config
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
	sessionRepo := repoFactory.Session()
	studentPreRegRepo := repoFactory.StudentPreRegistration()
	universityRepo := repoFactory.University()

	// =========================
	// Secondary adapters
	// =========================
	hasher := auth.NewBcryptHasher()
	jwtProvider := auth.NewJWTProvider(cfg.JWTSecret, sessionRepo)

	// =========================
	// Logger
	// =========================
	appLogger := middleware.NewStdLogger()

	// =========================
	// Core services
	// =========================
	authService := services.NewAuthService(
		userRepo,
		hasher,
		jwtProvider,
		sessionRepo,
	)
	userService := services.NewUserService(userRepo)
	StudentRegistrationService := services.NewStudentRegistrationService(
		userRepo,
		studentPreRegRepo,
		emailAdapter.NewSMTPEmailService(cfg.SMTP), // TODO: replace with real email service
		hasher,
		universityRepo,
	)

	// =========================
	// Handlers
	// =========================
	authH := authHandler.NewHandler(authService, StudentRegistrationService, userService)
	userH := userHandler.NewHandler(userService)

	// =========================
	// Routes
	// =========================
	authRoutes := authHTTP.NewRoutes(authH, appLogger)
	userRoutes := userHTTP.NewRoutes(userH)

	// =========================
	// HTTP Router Adapter
	// =========================
	router := httpadapter.NewRouter(
		cfg,
		authRoutes,
		userRoutes,
		jwtProvider,
		appLogger,
	)

	// =========================
	// Server
	// =========================
	log.Printf("🚀 Dealna server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
