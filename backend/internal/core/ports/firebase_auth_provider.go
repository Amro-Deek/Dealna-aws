package ports

import "context"

type IFirebaseAuthProvider interface {
	GenerateCustomToken(ctx context.Context, userID string) (string, error)
}
