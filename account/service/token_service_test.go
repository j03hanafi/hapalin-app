package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func TestNewPairFromUser(t *testing.T) {
	var idExp int64 = 15 * 60
	var refreshExp int64 = 3 * 24 * 2600

	private, _ := os.ReadFile("../rsa_private_test.pem")
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(private)
	public, _ := os.ReadFile("../rsa_public_test.pem")
	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(public)
	secret := "aNotSoRandomSecret"

	mockTokenRepository := new(mocks.MockTokenRepository)

	// instantiate a common token service to be used by all tests
	tokenServ := NewTokenService(&TSConfig{
		PrivateKey:            privateKey,
		PublicKey:             publicKey,
		RefreshSecret:         secret,
		IDExpirationSecs:      idExp,
		RefreshExpirationSecs: refreshExp,
		TokenRepository:       mockTokenRepository,
	})

	// include password to make sure it is not serialized
	// since json tag is "-"
	uid, _ := uuid.NewRandom()
	user := &domain.User{
		UID:      uid,
		Email:    "bob@bob.com",
		Password: "atLeast8Chars",
	}

	// Setup mock call response in setup
	uidErrorCase, _ := uuid.NewRandom()
	uErrorCase := &domain.User{
		UID:      uidErrorCase,
		Email:    "failure@failure.com",
		Password: "atLeast8Chars",
	}
	prevID := "a_previous_id"

	setSuccessArguments := mock.Arguments{
		mock.AnythingOfType("context.backgroundCtx"),
		user.UID.String(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
	}

	setErrorArguments := mock.Arguments{
		mock.AnythingOfType("context.backgroundCtx"),
		uErrorCase.UID.String(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
	}

	deleteWithPrevIDArguments := mock.Arguments{
		mock.AnythingOfType("context.backgroundCtx"),
		user.UID.String(),
		prevID,
	}

	// mock call argument/responses
	mockTokenRepository.
		On("SetRefreshToken", setSuccessArguments...).
		Return(nil)
	mockTokenRepository.
		On("SetRefreshToken", setErrorArguments...).
		Return(fmt.Errorf("error setting refresh token"))
	mockTokenRepository.
		On("DeleteRefreshToken", deleteWithPrevIDArguments...).
		Return(nil)

	t.Run("Returns a token pair with values", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		tokenPair, err := tokenServ.NewPairFromUser(ctx, user, prevID)
		assert.NoError(t, err)

		// SetRefreshToken should be called with setSuccessArguments
		mockTokenRepository.AssertCalled(t, "SetRefreshToken", setSuccessArguments...)
		// DeleteRefreshToken should not be called since prevID is ""
		mockTokenRepository.AssertCalled(t, "DeleteRefreshToken", deleteWithPrevIDArguments...)

		var s string
		assert.IsType(t, s, tokenPair.IDToken.SS)

		// decode the Base64URL encoded string
		// simpler to use jwt library which is already imported
		idTokenClaims := &idTokenCustomClaims{}

		_, err = jwt.ParseWithClaims(tokenPair.IDToken.SS, idTokenClaims, func(token *jwt.Token) (interface{}, error) {
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
		expectedExpiresAt := time.Now().Add(time.Duration(idExp) * time.Second)
		assert.WithinDuration(t, expectedExpiresAt, expiresAt, 5*time.Second)

		refreshTokenClaims := &refreshTokenCustomClaims{}
		_, err = jwt.ParseWithClaims(tokenPair.RefreshToken.SS, refreshTokenClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		assert.NoError(t, err)

		assert.IsType(t, s, tokenPair.RefreshToken.SS)
		assert.Equal(t, user.UID, refreshTokenClaims.UID)

		expiresAt = time.Unix(refreshTokenClaims.RegisteredClaims.ExpiresAt.Unix(), 0)
		expectedExpiresAt = time.Now().Add(time.Duration(refreshExp) * time.Second)
		assert.WithinDuration(t, expectedExpiresAt, expiresAt, 5*time.Second)
	})

	t.Run("Error setting refresh token", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		_, err := tokenServ.NewPairFromUser(ctx, uErrorCase, "")
		assert.Error(t, err)

		// SetRefreshToken should be called with setSuccessArguments
		mockTokenRepository.AssertCalled(t, "SetRefreshToken", setErrorArguments...)
		// DeleteRefreshToken should not be called since prevID is ""
		mockTokenRepository.AssertNotCalled(t, "DeleteRefreshToken")
	})

	t.Run("Empty string provided for prevID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		_, err := tokenServ.NewPairFromUser(ctx, user, "")
		assert.NoError(t, err)

		// SetRefreshToken should be called with setSuccessArguments
		mockTokenRepository.AssertCalled(t, "SetRefreshToken", setSuccessArguments...)
		// DeleteRefreshToken should not be called since prevID is ""
		mockTokenRepository.AssertNotCalled(t, "DeleteRefreshToken")
	})
}
