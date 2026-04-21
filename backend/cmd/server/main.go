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
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/giveaway"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/items"
	profileHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile"
	profileHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile/http"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/social"
	userHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	// Secondary adapters
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	emailAdapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/email"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence"
	postgres "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence/postgres"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/storage"

	// AWS
	context "context"

	// Core services
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/services"

	// Swagger docs (USED via routes.go)
	_ "github.com/Amro-Deek/Dealna-aws/backend/docs"
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
	studentPreRegRepo := repoFactory.StudentPreRegistration()
	universityRepo := repoFactory.University()
	itemRepo := repoFactory.Item()

	// =========================
	// Secondary adapters
	// =========================
	//hasher := auth.NewBcryptHasher()
	jwksURL := cfg.KeycloakBaseURL + "/realms/" + cfg.KeycloakRealm + "/protocol/openid-connect/certs"
	jwtProvider, err := auth.NewKeycloakJWTProvider(jwksURL, userRepo)
	if err != nil {
		log.Fatalf("failed to initialize Keycloak JWT provider: %v", err)
	}
	// =========================
	// Secondary adapters
	// =========================
	keycloakIdentity := auth.NewKeycloakIdentityProvider(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
		&http.Client{},
	)

	s3Provider, err := storage.NewS3Provider(context.TODO(), "us-east-1", "dealna-bzu-storage")
	if err != nil {
		log.Fatalf("failed to initialize S3 provider: %v", err)
	}

	// =========================
	// Logger
	// =========================
	appLogger := middleware.NewStdLogger()

	// =========================
	// Core services
	// =========================
	authService := services.NewAuthService(
		userRepo,
		keycloakIdentity,
	)
	userService := services.NewUserService(userRepo)
	StudentRegistrationService := services.NewStudentRegistrationService(
		userRepo,
		studentPreRegRepo,
		emailAdapter.NewSMTPEmailService(cfg.SMTP), // TODO: replace with real email service
		keycloakIdentity,
		universityRepo,
	)
	profileService := services.NewProfileService(userRepo)
	storageService := services.NewStorageService(s3Provider)
	itemService := services.NewItemService(itemRepo, s3Provider)

	// =========================
	// Handlers
	// =========================
	authH := authHandler.NewHandler(authService, StudentRegistrationService, userService)
	userH := userHandler.NewHandler(userService)
	profileH := profileHandler.NewProfileHandler(profileService, storageService, appLogger)
	itemH := items.NewItemHandler(itemService, appLogger)

	// =========================
	// Routes
	// =========================
	authRoutes := authHTTP.NewRoutes(authH, appLogger)
	userRoutes := userHTTP.NewRoutes(userH)
	profileRoutes := profileHTTP.NewRoutes(profileH)
	itemRoutes := items.NewRoutes(itemH)

	// =========================
	// Giveaway Setup
	// =========================
	queueRepo := postgres.NewQueueRepository(db)
	purchaseRepo := postgres.NewPurchaseRequestRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)

	notificationSvc := services.NewNotificationService(notificationRepo)
	queueSvc := services.NewQueueService(queueRepo, notificationSvc)
	queueSvc.StartWorkers(context.Background())
	purchaseSvc := services.NewPurchaseService(purchaseRepo, notificationSvc)
	transactionSvc := services.NewTransactionService(transactionRepo, notificationSvc)

	queueH := giveaway.NewQueueHandler(queueSvc)
	purchaseH := giveaway.NewPurchaseHandler(purchaseSvc)
	transactionH := giveaway.NewTransactionHandler(transactionSvc)
	notificationH := giveaway.NewNotificationHandler(notificationSvc)

	giveawayRoutes := giveaway.NewRoutes(queueH, purchaseH, transactionH, notificationH)

	// =========================
	// Follow / Social Setup
	// =========================
	followRepo := postgres.NewFollowRepository(db)
	followSvc := services.NewFollowService(followRepo)
	followH := social.NewFollowHandler(followSvc)
	socialRoutes := social.NewRoutes(followH, profileH)

	// =========================
	// HTTP Router Adapter
	// =========================
	router := httpadapter.NewRouter(
		cfg,
		authRoutes,
		userRoutes,
		profileRoutes,
		itemRoutes,
		giveawayRoutes,
		socialRoutes,
		jwtProvider,
		appLogger,
	)

	// =========================
	// Server
	// =========================
	log.Printf("🚀 Dealna server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
