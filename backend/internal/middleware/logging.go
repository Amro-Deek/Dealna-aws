package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

//
// ==================================================
// Logger Interface (LOCAL)
// ==================================================
//

type StructuredLoggerInterface interface {
	Debug(ctx context.Context, event string, fields map[string]any)
	Info(ctx context.Context, event string, fields map[string]any)
	Warn(ctx context.Context, event string, fields map[string]any)
	Error(ctx context.Context, event string, fields map[string]any)
}

//
// ==================================================
// Context helpers
// ==================================================
//


func CorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextCorrelationID).(string); ok {
		return v
	}
	return ""
}

//
// ==================================================
// Request Logging Middleware
// ==================================================
//

func RequestLoggingMiddleware(
	logger StructuredLoggerInterface, // ✅ بدون logging.
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			ctx := r.Context()
			ctx = WithCorrelationID(ctx, correlationID)

			if r.Header.Get("X-Debug") == "true" {
				ctx = WithVerbose(ctx, true)
			}

			w.Header().Set("X-Correlation-ID", correlationID)

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			logger.Debug(ctx, "request.start", map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			latencyMs := time.Since(start).Milliseconds()

			fields := map[string]any{
				"method":     r.Method,
				"path":       r.URL.Path,
				"statusCode": wrapped.statusCode,
				"latencyMs":  latencyMs,
			}

			if wrapped.statusCode >= 400 {
				logger.Warn(ctx, "request.complete", fields)
			} else {
				logger.Info(ctx, "request.complete", fields)
			}
		})
	}
}

//
// ==================================================
// responseWriter
// ==================================================
//

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

//
// ==================================================
// Recover Middleware
// ==================================================
//

func RecoverMiddleware(
	logger StructuredLoggerInterface,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					ctx := r.Context()

					logger.Error(ctx, "panic.recovered", map[string]any{
						"panic": rec,
						"path":  r.URL.Path,
					})

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{
						"type":"error",
						"version":"v1",
						"code":"INTERNAL_ERROR",
						"message":"An unexpected error occurred",
						"statusCode":500,
						"retryable":true
					}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
