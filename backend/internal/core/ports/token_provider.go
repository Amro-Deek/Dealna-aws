// internal/core/ports/token_provider.go
package ports

import "time"

type ITokenProvider interface {
    GenerateToken(
        userID string,
        role string,
        expiresAt time.Time,
    ) (string, error)
}
