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
}