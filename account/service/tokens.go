package service

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"time"
)

// idTokenCustomClaims holds structure of jwt claims of idToken
type idTokenCustomClaims struct {
	User *domain.User `json:"user"`
	jwt.RegisteredClaims
}

// generateIDToken generates an IDToken which is a jwt with myCustomClaims
// Could call this GenerateIDTokenString, but the signature makes this fairly clear
func generateIDToken(u *domain.User, key *rsa.PrivateKey, exp int64) (string, error) {
	l := logger.Get()

	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(exp) * time.Second) // 15 minutes

	claims := idTokenCustomClaims{
		User: u,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(tokenExp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		l.Error("Error signing token",
			zap.Error(err),
		)
		return "", err
	}

	return ss, nil
}

// refreshTokenData holds the actual signed jwt string along with the ID
// We return the id, so it can be used without re-parsing the JWT from signed string
type refreshTokenData struct {
	SS        string
	ID        uuid.UUID
	ExpiresIn time.Duration
}

// refreshTokenCustomClaims holds the payload of a refresh token
// This can be used to extract user id for subsequent
// application operations (IE, fetch user in Redis)
type refreshTokenCustomClaims struct {
	UID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

// generateRefreshToken creates a refresh token
// The refresh token stores only the user's ID, a string
func generateRefreshToken(uid uuid.UUID, key string, exp int64) (*refreshTokenData, error) {
	l := logger.Get()

	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(exp) * time.Second) // 3 days
	tokenID, err := uuid.NewRandom()
	if err != nil {
		l.Error("Error generating token id",
			zap.Error(err),
		)
	}

	claims := refreshTokenCustomClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(tokenExp),
			ID:        tokenID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(key))
	if err != nil {
		l.Error("Error signing token",
			zap.Error(err),
		)
		return nil, err
	}

	return &refreshTokenData{
		SS:        ss,
		ID:        tokenID,
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}

// validateIDToken returns the token's claims if the token is valid
func validateIDToken(tokenString string, key *rsa.PublicKey) (*idTokenCustomClaims, error) {
	claims := &idTokenCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*idTokenCustomClaims)
	if !ok {
		return nil, fmt.Errorf("couldn't parse claims")
	}

	return claims, nil
}
