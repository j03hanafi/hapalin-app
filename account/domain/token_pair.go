package domain

import "context"

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
