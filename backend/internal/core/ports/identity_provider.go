package ports

import "context"

type IdentityLoginResult struct {
	AccessToken  string
	RefreshToken string
	Subject      string
	ExpiresIn    int
}

type IIdentityProvider interface {
	Login(ctx context.Context, username, password string) (*IdentityLoginResult, error)
	// RegisterUser creates a new user in Keycloak and returns their unique Subject (ID).
	RegisterUser(ctx context.Context, email, password, firstName, lastName string) (keycloakSub string, err error)
	// DeleteUser removes a user from Keycloak by their Subject (ID).
	DeleteUser(ctx context.Context, keycloakSub string) error

	// Refresh exchanges a refresh token for a new access token
	Refresh(ctx context.Context, refreshToken string) (*IdentityLoginResult, error)
	// Logout invalidates a session using the refresh token
	Logout(ctx context.Context, refreshToken string) error
}