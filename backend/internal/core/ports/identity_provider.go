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
	RegisterUser(ctx context.Context, email, password string) (keycloakSub string, err error)
	// DeleteUser removes a user from Keycloak by their Subject (ID).
	DeleteUser(ctx context.Context, keycloakSub string) error
}