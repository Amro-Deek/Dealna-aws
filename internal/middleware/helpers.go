package middleware

import (
	"context"
)

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(ContextUserID).(string)
	return userID
}

func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ContextRole).(string)
	return role
}
