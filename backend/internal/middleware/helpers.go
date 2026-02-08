package middleware

import "context"

// =========================
// Getters
// =========================

func CorrelationIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ContextCorrelationID).(string)
	return id
}

func IsVerbose(ctx context.Context) bool {
	v, _ := ctx.Value(ContextVerbose).(bool)
	return v
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(ContextUserID).(string)
	return userID
}

func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ContextRole).(string)
	return role
}
