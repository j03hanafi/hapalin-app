package service

import (
	"context"
	"crypto/rsa"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
)

// tokenService used for injecting an implementation of TokenRepository
// for use in service methods along with keys and secrets for
// signing JWTs
type tokenService struct {
	TokenRepository       domain.TokenRepository
	PrivateKey            *rsa.PrivateKey
	PublicKey             *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

// TSConfig will hold repositories that will eventually be injected into this
// service layer
type TSConfig struct {
	TokenRepository       domain.TokenRepository
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
		TokenRepository:       c.TokenRepository,
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

	// delete user's current refresh token (used when refreshing idToken)
	if prevTokenID != "" {
		if err := t.TokenRepository.DeleteRefreshToken(ctx, u.UID.String(), prevTokenID); err != nil {
			l.Error("Error deleting previous refresh token for user",
				zap.Error(err),
			)
			return nil, err
		}
	}

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

	// set refresh tokens by calling TokenRepository methods
	if err = t.TokenRepository.SetRefreshToken(ctx, u.UID.String(), refreshToken.ID.String(), refreshToken.ExpiresIn); err != nil {
		l.Error("Error saving refresh token for user",
			zap.Error(err),
		)
		return nil, apperrors.NewInternal()
	}

	return &domain.TokenPair{
		IDToken:      domain.IDToken{SS: idToken},
		RefreshToken: domain.RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: u.UID},
	}, nil
}

// ValidateIDToken validates the id token jwt string
// It returns the user extract from the idTokenCustomClaims
func (t tokenService) ValidateIDToken(tokenString string) (*domain.User, error) {
	l := logger.Get()

	claims, err := validateIDToken(tokenString, t.PublicKey)
	if err != nil {
		l.Error("Error validating ID Token",
			zap.Error(err),
		)
		return nil, apperrors.NewAuthorization("Invalid ID Token")
	}

	return claims.User, nil
}

// ValidateRefreshToken checks to make sure the JWT provided by a string is valid
// and returns a RefreshToken if valid
func (t tokenService) ValidateRefreshToken(tokenString string) (*domain.RefreshToken, error) {
	l := logger.Get()

	// validate actual JWT with string a secret
	claims, err := validateRefreshToken(tokenString, t.RefreshSecret)
	if err != nil {
		l.Error("Error validating Refresh Token",
			zap.String("tokenString", tokenString),
			zap.Error(err),
		)
		return nil, apperrors.NewAuthorization("Unable to verify user from refresh token")
	}

	// Registered claims store ID as a string
	// parse claims.ID as a uuid
	tokenUUID, err := uuid.Parse(claims.ID)
	if err != nil {
		l.Error("Error parsing token UUID",
			zap.String("tokenUUID", claims.ID),
			zap.Error(err),
		)
		return nil, apperrors.NewAuthorization("Unable to verify user from refresh token")
	}

	return &domain.RefreshToken{
		ID:  tokenUUID,
		UID: claims.UID,
		SS:  tokenString,
	}, nil
}

// SignOut signs out a user by deleting all refresh tokens associated with a user
func (t tokenService) SignOut(ctx context.Context, uid uuid.UUID) error {
	return t.TokenRepository.DeleteUserRefreshTokens(ctx, uid.String())
}
