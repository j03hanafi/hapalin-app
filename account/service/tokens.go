package service

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"time"
)

// IDTokenCustomClaims holds structure of jwt claims of idToken
type IDTokenCustomClaims struct {
	User *domain.User `json:"user"`
	jwt.RegisteredClaims
}

// generateIDToken generates an IDToken which is a jwt with myCustomClaims
// Could call this GenerateIDTokenString, but the signature makes this fairly clear
func generateIDToken(u *domain.User, key *rsa.PrivateKey) (string, error) {
	l := logger.Get()

	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Minute * 15) // 15 minutes

	claims := IDTokenCustomClaims{
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

// RefreshToken holds the actual signed jwt string along with the ID
// We return the id, so it can be used without re-parsing the JWT from signed string
type RefreshToken struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}

// RefreshTokenCustomClaims holds the payload of a refresh token
// This can be used to extract user id for subsequent
// application operations (IE, fetch user in Redis)
type RefreshTokenCustomClaims struct {
	UID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

// generateRefreshToken creates a refresh token
// The refresh token stores only the user's ID, a string
func generateRefreshToken(uid uuid.UUID, key string) (*RefreshToken, error) {
	l := logger.Get()

	currentTime := time.Now()
	tokenExp := currentTime.AddDate(0, 0, 3) // 3 days
	tokenID, err := uuid.NewRandom()
	if err != nil {
		l.Error("Error generating token id",
			zap.Error(err),
		)
	}

	claims := RefreshTokenCustomClaims{
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

	return &RefreshToken{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}
