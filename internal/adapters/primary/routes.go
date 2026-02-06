package httpadapter

import (
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/internal/config"

	authHTTP "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/auth/http"
	userHTTP "github.com/Amro-Deek/Dealna-aws/internal/adapters/primary/users/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	authRoutes *authHTTP.Routes
	userRoutes *userHTTP.Routes
	cfg        *config.Config
}

func NewRouter(
	cfg *config.Config,
	authRoutes *authHTTP.Routes,
	userRoutes *userHTTP.Routes,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		authRoutes.Register(r)
		userRoutes.Register(r, cfg.JWTSecret)
	})

	return r
}
