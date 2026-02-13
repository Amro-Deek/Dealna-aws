package middleware

import (
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports")

// AuthMiddleware is a pure pipe:
// - extracts auth header
// - delegates auth decision to provider
// - enriches context
// - never decides business meaning of the error
func AuthMiddleware(
	auth ports.IAuthContextProvider,
	logger StructuredLoggerInterface,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			authCtx, err := auth.Authenticate(r.Context(), authHeader)
			if err != nil {
				WriteErrorResponse(w, r.Context(), err, logger)
				return
			}

			ctx := WithUserID(r.Context(), authCtx.UserID)
			ctx = WithRole(ctx, authCtx.Role)
			ctx = WithJTI(ctx, authCtx.JTI)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
