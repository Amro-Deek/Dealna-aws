package middleware

import "context"

// unexported type (safe context keys)
type contextKey string

const (
	ContextCorrelationID contextKey = "correlation_id"
	ContextVerbose       contextKey = "verbose"

	ContextUserID contextKey = "user_id"
	ContextRole   contextKey = "role"
)

// =========================
// Setters
// =========================

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, ContextCorrelationID, correlationID)
}

func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, ContextVerbose, verbose)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextUserID, userID)
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ContextRole, role)
}
