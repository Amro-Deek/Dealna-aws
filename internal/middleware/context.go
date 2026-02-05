package middleware

// contextKey is an unexported custom type to avoid collisions
type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextRole   contextKey = "role"
)
