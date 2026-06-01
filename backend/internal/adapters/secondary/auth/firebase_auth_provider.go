package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type FirebaseAuthProvider struct {
	client *auth.Client
}

func NewFirebaseAuthProvider(ctx context.Context) (*FirebaseAuthProvider, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting auth client: %v", err)
	}

	return &FirebaseAuthProvider{client: client}, nil
}

func (p *FirebaseAuthProvider) GenerateCustomToken(ctx context.Context, userID string) (string, error) {
	token, err := p.client.CustomToken(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("error generating custom token: %v", err)
	}
	return token, nil
}
