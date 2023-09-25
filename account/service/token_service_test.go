package service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNewPairFromUser(t *testing.T) {
	private, _ := os.ReadFile("../rsa_private_test.pem")
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(private)
	public, _ := os.ReadFile("../rsa_public_test.pem")
	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(public)
	secret := "aNotSoRandomSecret"

	// instantiate a common token service to be used by all tests
	tokenService := NewTokenService(&TSConfig{
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		RefreshSecret: secret,
	})

	// include password to make sure it is not serialized
	// since json tag is "-"
	uid, _ := uuid.NewRandom()
	user := &domain.User{
		UID:      uid,
		Email:    "bob@bob.com",
		Password: "atLeast8Chars",
	}

	t.Run("Returns a token pair with values", func(t *testing.T) {
		ctx := context.TODO()
		tokenPair, err := tokenService.NewPairFromUser(ctx, user, "")
		assert.NoError(t, err)

		var s string
		assert.IsType(t, s, tokenPair.IDToken)

		// decode the Base64URL encoded string
		// simpler to use jwt library which is already imported
		idTokenClaims := &IDTokenCustomClaims{}

		_, err = jwt.ParseWithClaims(tokenPair.IDToken, idTokenClaims, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		assert.NoError(t, err)

		// assert claims on idToken
		expectedClaims := []interface{}{
			user.UID,
			user.Email,
			user.Name,
			user.ImageURL,
			user.Website,
		}
		actualIDClaims := []interface{}{
			idTokenClaims.User.UID,
			idTokenClaims.User.Email,
			idTokenClaims.User.Name,
			idTokenClaims.User.ImageURL,
			idTokenClaims.User.Website,
		}

		assert.ElementsMatch(t, expectedClaims, actualIDClaims)
		assert.Empty(t, idTokenClaims.User.Password) // password should never be encoded to json

		expiresAt := time.Unix(idTokenClaims.RegisteredClaims.ExpiresAt.Unix(), 0)
		expectedExpiresAt := time.Now().Add(time.Minute * 15)
		assert.WithinDuration(t, expectedExpiresAt, expiresAt, 5*time.Second)

		refreshTokenClaims := &RefreshTokenCustomClaims{}
		_, err = jwt.ParseWithClaims(tokenPair.RefreshToken, refreshTokenClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		assert.NoError(t, err)

		assert.IsType(t, s, tokenPair.RefreshToken)
		assert.Equal(t, user.UID, refreshTokenClaims.UID)

		expiresAt = time.Unix(refreshTokenClaims.RegisteredClaims.ExpiresAt.Unix(), 0)
		expectedExpiresAt = time.Now().AddDate(0, 0, 3)
		assert.WithinDuration(t, expectedExpiresAt, expiresAt, 5*time.Second)
	})
}
