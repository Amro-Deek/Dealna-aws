package middleware

import (
	"net/http"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)
// middleware are just pipe , it calls , passes ,does not decide what to do with the error 
// , it just passes it to the next layer (handler) to decide how to handle it
func AuthMiddleware(
	auth ports.IAuthContextProvider,
	logger StructuredLoggerInterface,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			authCtx, err := auth.Authenticate(r.Context(), authHeader)
			if err != nil {
				// ✅ مرّر الخطأ كما هو (NO duplication)
				WriteErrorResponse(
					w,
					r.Context(),
					err,
					logger,
				)
				return
			}

			ctx := r.Context()
			ctx = WithUserID(ctx, authCtx.UserID)
			ctx = WithRole(ctx, authCtx.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
