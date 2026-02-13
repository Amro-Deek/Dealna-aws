package ports

import "context"

// AuthContext represents authenticated identity
type AuthContext struct {
	UserID string
	Role   string
	JTI    string
}

// IAuthContextProvider abstracts auth mechanism (JWT, Cognito, Auth0...)
type IAuthContextProvider interface {
	Authenticate(ctx context.Context, authHeader string) (*AuthContext, error)
}
