package ports

import "time"

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

type ParsedToken struct {
	UserID string
	Role   string
	JTI    string
	Type   TokenType
	Exp    time.Time
}

type ITokenProvider interface {
	GenerateAccessToken(userID, role, jti string, expiresAt time.Time) (string, error)
	GenerateRefreshToken(userID, role, jti string, expiresAt time.Time) (string, error)

	// Used by refresh endpoint (token comes from body, not Authorization header)
	ParseToken(token string) (*ParsedToken, error)
}
