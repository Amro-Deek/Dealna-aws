package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Amro-Deek/Dealna-aws/internal/database"
	"github.com/Amro-Deek/Dealna-aws/internal/handlers"
	"github.com/Amro-Deek/Dealna-aws/internal/middleware"
	"github.com/Amro-Deek/Dealna-aws/internal/utils"

	_ "github.com/Amro-Deek/Dealna-aws/docs"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// 1. Load .env
	_ = godotenv.Load()

	// 2. Database Connection
	db, err := database.Connect(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	if err != nil {
		log.Fatalf("❌ Database Error: %v", err)
	}
	database.SetDB(db)
	defer db.Close()

	// 3. Router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// 4. Routes
	r.Route("/api/v1", func(r chi.Router) {

		// Health
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, true, "Dealna API is online", map[string]string{
				"status": "ok",
			}, nil)
		})

		// Auth
		r.Post("/auth/login", handlers.Login)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(os.Getenv("JWT_SECRET")))

			r.Get("/me", handlers.GetMe)
		})
	})

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Dealna Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
