package service

import (
	"context"
	"crypto/rsa"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
)

// tokenService used for injecting an implementation of TokenRepository
// for use in service methods along with keys and secrets for
// signing JWTs
type tokenService struct {
	// TokenRepository domain.TokenRepository
	PrivateKey            *rsa.PrivateKey
	PublicKey             *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

// TSConfig will hold repositories that will eventually be injected into this
// service layer
type TSConfig struct {
	// TokenRepository domain.TokenRepository
	PrivateKey            *rsa.PrivateKey
	PublicKey             *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

// NewTokenService is a factory function for
// initializing a userService with its repository layer dependencies
func NewTokenService(c *TSConfig) domain.TokenService {
	return &tokenService{
		// TokenRepository: c.TokenRepository,
		PrivateKey:            c.PrivateKey,
		PublicKey:             c.PublicKey,
		RefreshSecret:         c.RefreshSecret,
		IDExpirationSecs:      c.IDExpirationSecs,
		RefreshExpirationSecs: c.RefreshExpirationSecs,
	}
}

// NewPairFromUser creates fresh id and refresh tokens for the current user
// If a previous token is included, the previous token is removed from
// the tokens repository
func (t tokenService) NewPairFromUser(ctx context.Context, u *domain.User, prevTokenID string) (*domain.TokenPair, error) {
	l := logger.Get()

	// No need to use a repository for idToken as it is unrelated to any data source
	idToken, err := generateIDToken(u, t.PrivateKey, t.IDExpirationSecs)
	if err != nil {
		l.Error("Error generating ID Token for user",
			zap.Error(err),
		)
		return nil, apperrors.NewInternal()
	}

	refreshToken, err := generateRefreshToken(u.UID, t.RefreshSecret, t.RefreshExpirationSecs)
	if err != nil {
		l.Error("Error generating Refresh Token for user",
			zap.Error(err),
		)
		return nil, apperrors.NewInternal()
	}

	// TODO: store refresh tokens by calling TokenRepository methods

	return &domain.TokenPair{
		IDToken:      idToken,
		RefreshToken: refreshToken.SS,
	}, nil
}
