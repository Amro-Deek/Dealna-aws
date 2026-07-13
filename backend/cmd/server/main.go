package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database/generated"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"

	// Primary adapters
	httpadapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary"
	adminHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/admin/http"
	authHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth"
	authHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/auth/http"
	chatHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/chat/http"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/giveaway"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/items"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/marketplace"
	profileHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile"
	profileHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/profile/http"
	ratingsHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/ratings/http"
	reportsHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/reports/http"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/social"
	userHandler "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users"
	userHTTP "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/users/http"

	// Secondary adapters
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	authAdapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	emailAdapter "github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/email"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/messaging"
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
	// AWS & External Infrastructure (SQS & Qdrant)
	// =========================
	// 1. Qdrant Setup
	qdrantClient, err := database.NewQdrantClient(cfg.QdrantURL, cfg.QdrantAPIKey)
	if err != nil {
		log.Printf("⚠️ Qdrant connection failed: %v", err)
	} else {
		// Initialize the collection and indexes only on startup
		if err := qdrantClient.InitQdrantSchema(context.Background()); err != nil {
			log.Printf("⚠️ Qdrant Schema Init warning: %v", err)
		}
	}

	lambdaPublisher, err := messaging.NewLambdaPublisher(context.TODO(), cfg.LambdaFunctionName, cfg.AWSRegion)
	if err != nil {
		log.Fatalf("failed to initialize Lambda Publisher: %v", err)
	}

	// 3. Keep Lambda warm to prevent 3-second cold starts!
	go func() {
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			// Silently ping the lambda to keep the container permanently hot
			_, _ = lambdaPublisher.GenerateEmbedding(context.Background(), "warmup")
		}
	}()

	// =========================
	// Repositories
	// =========================
	repoFactory := persistence.NewRepositoryFactory(db)
	userRepo := repoFactory.User()
	studentPreRegRepo := repoFactory.StudentPreRegistration()
	universityRepo := repoFactory.University()
	itemRepo := repoFactory.Item()
	providerRepo := repoFactory.Provider()
	adminRepo := persistence.NewAdminRepository(db)

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
		emailAdapter.NewSMTPEmailService(cfg.SMTP),
	)
	userService := services.NewUserService(userRepo)
	StudentRegistrationService := services.NewStudentRegistrationService(
		userRepo,
		studentPreRegRepo,
		emailAdapter.NewSMTPEmailService(cfg.SMTP), // TODO: replace with real email service
		keycloakIdentity,
		universityRepo,
	)
	notificationRepo := postgres.NewNotificationRepository(db)
	fcmClient, err := messaging.NewFCMClient(context.Background())
	if err != nil {
		log.Printf("⚠️ FCM initialization failed (push notifications disabled): %v", err)
	}
	notificationSvc := services.NewNotificationService(notificationRepo, userRepo, itemRepo, fcmClient)

	providerPreRegRepo := postgres.NewProviderPreRegistrationRepository(db)
	providerRegistrationService := services.NewProviderRegistrationService(
		userRepo,
		providerRepo,
		providerPreRegRepo,
		emailAdapter.NewSMTPEmailService(cfg.SMTP), // TODO: replace with real email service
		keycloakIdentity,
		s3Provider,
		notificationSvc,
	)
	profileService := services.NewProfileService(userRepo)
	storageService := services.NewStorageService(s3Provider)
	itemService := services.NewItemService(itemRepo, s3Provider, lambdaPublisher, qdrantClient)
	adminService := services.NewAdminService(adminRepo, s3Provider, emailAdapter.NewSMTPEmailService(cfg.SMTP), keycloakIdentity, appLogger, notificationSvc, itemRepo)

	// =========================
	// Handlers
	// =========================
	authH := authHandler.NewHandler(authService, StudentRegistrationService, providerRegistrationService, userService)
	userH := userHandler.NewHandler(userService)
	profileH := profileHandler.NewProfileHandler(profileService, storageService, appLogger)
	itemH := items.NewItemHandler(itemService, appLogger)

	// =========================
	// Reporting Setup
	// =========================
	reportRepo := postgres.NewReportRepository(generated.New(db))
	reportSvc := services.NewReportService(reportRepo, appLogger)
	reportH := reportsHTTP.NewReportHandler(reportSvc, s3Provider, appLogger)
	reportRoutes := reportsHTTP.NewRoutes(reportH)

	adminH := adminHTTP.NewAdminHandler(adminService, reportSvc, appLogger)

	// =========================
	// Routes
	// =========================
	authRoutes := authHTTP.NewRoutes(authH, appLogger)
	userRoutes := userHTTP.NewRoutes(userH)
	profileRoutes := profileHTTP.NewRoutes(profileH)
	itemRoutes := items.NewRoutes(itemH, appLogger)

	// =========================
	// Giveaway Setup
	// =========================
	queueRepo := postgres.NewQueueRepository(db)
	purchaseRepo := postgres.NewPurchaseRequestRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	queueSvc := services.NewQueueService(queueRepo, notificationSvc, itemRepo)
	queueSvc.StartWorkers(context.Background())
	purchaseSvc := services.NewPurchaseService(purchaseRepo, notificationSvc, itemRepo, transactionRepo)
	transactionSvc := services.NewTransactionService(transactionRepo, purchaseRepo, itemRepo, notificationSvc)

	queueH := giveaway.NewQueueHandler(queueSvc)
	purchaseH := giveaway.NewPurchaseHandler(purchaseSvc)
	transactionH := giveaway.NewTransactionHandler(transactionSvc)
	notificationH := giveaway.NewNotificationHandler(notificationSvc)

	giveawayRoutes := giveaway.NewRoutes(queueH, notificationH, appLogger)
	marketplaceRoutes := marketplace.NewRoutes(purchaseH, transactionH, appLogger)

	ratingRepo := postgres.NewRatingRepository(generated.New(db))
	ratingService := services.NewRatingService(ratingRepo, transactionRepo, userRepo, notificationSvc)
	go ratingService.StartRatingReminderWorker(context.Background())
	ratingH := ratingsHTTP.NewRatingHandler(ratingService)
	ratingRoutes := ratingsHTTP.NewRoutes(ratingH)

	// =========================
	// Follow / Social Setup
	// =========================
	followRepo := postgres.NewFollowRepository(db)
	followSvc := services.NewFollowService(followRepo)
	followH := social.NewFollowHandler(followSvc)
	socialRoutes := social.NewRoutes(followH, profileH)

	// =========================
	// Chat Setup
	// =========================
	firebaseAuthProv, err := authAdapter.NewFirebaseAuthProvider(context.Background())
	if err != nil {
		log.Printf("⚠️ Firebase Auth initialization failed: %v", err)
	}
	chatSvc := services.NewChatService(firebaseAuthProv, notificationSvc)
	chatH := chatHTTP.NewChatHandler(chatSvc, appLogger)
	chatRoutes := chatHTTP.NewRoutes(chatH, appLogger)

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
		marketplaceRoutes,
		socialRoutes,
		chatRoutes,
		ratingRoutes,
		adminH,
		reportRoutes,
		jwtProvider,
		appLogger,
	)

	// =========================
	// Server
	// =========================
	log.Printf("🚀 Dealna server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
