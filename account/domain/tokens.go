package domain

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// TokenPair used for returning pairs of id and refresh tokens
type TokenPair struct {
	IDToken
	RefreshToken
}

// RefreshToken stores token properties that
// are accessed in multiple application layer
type RefreshToken struct {
	ID  uuid.UUID `json:"-"`
	UID uuid.UUID `json:"-"`
	SS  string    `json:"refresh_token"`
}

// IDToken stores token properties that
// are accessed in multiple application layers
type IDToken struct {
	SS string `json:"id_token"`
}

// TokenService defines methods the handler layer expects to interact
// with in regard to producing JWTs as string
type TokenService interface {
	NewPairFromUser(ctx context.Context, u *User, prevTokenID string) (*TokenPair, error)
	ValidateIDToken(tokenString string) (*User, error)
	ValidateRefreshToken(refreshTokenString string) (*RefreshToken, error)
}

// TokenRepository defines methods it expects a repository
// it interacts with to implement
type TokenRepository interface {
	SetRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error
	DeleteRefreshToken(ctx context.Context, userID, prevTokenID string) error
}
