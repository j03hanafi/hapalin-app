package domain

import (
	"context"
	"time"
)

// TokenPair used for returning pairs of id and refresh tokens
type TokenPair struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenService defines methods the handler layer expects to interact
// with in regard to producing JWTs as string
type TokenService interface {
	NewPairFromUser(ctx context.Context, u *User, prevTokenID string) (*TokenPair, error)
}

// TokenRepository defines methods it expects a repository
// it interacts with to implement
type TokenRepository interface {
	SetRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error
	DeleteRefreshToken(ctx context.Context, userID, prevTokenID string) error
}
