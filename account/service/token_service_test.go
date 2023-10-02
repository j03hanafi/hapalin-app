package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func TestNewPairFromUser(t *testing.T) {
	t.Parallel()

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

	t.Run("Prev token is not in repository", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		tokenIDNotInRepo := "not_in_repo"

		deleteArgs := mock.Arguments{
			ctx,
			user.UID.String(),
			tokenIDNotInRepo,
		}
		mockError := apperrors.NewAuthorization("Invalid refresh token")
		mockTokenRepository.
			On("DeleteRefreshToken", deleteArgs...).
			Return(mockError)

		_, err := tokenServ.NewPairFromUser(ctx, user, tokenIDNotInRepo)
		assert.Error(t, err)

		appError, ok := err.(*apperrors.Error)

		assert.True(t, ok)
		assert.Equal(t, apperrors.Authorization, appError.Type)

		mockTokenRepository.AssertCalled(t, "DeleteRefreshToken", deleteArgs...)
		mockTokenRepository.AssertNotCalled(t, "SetRefreshToken")
	})
}

func TestValidateIDToken(t *testing.T) {
	t.Parallel()

	var idExp int64 = 15 * 60

	private, _ := os.ReadFile("../rsa_private_test.pem")
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(private)
	public, _ := os.ReadFile("../rsa_public_test.pem")
	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(public)

	// instantiate a common token service to be used by all tests
	tokenServ := NewTokenService(&TSConfig{
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		IDExpirationSecs: idExp,
	})

	// include password to make sure it is not serialized
	// since json tag is "-"
	uid, _ := uuid.NewRandom()
	user := &domain.User{
		UID:      uid,
		Email:    "bob@bob.com",
		Password: "atLeast8Chars",
	}

	t.Run("Valid token", func(t *testing.T) {
		t.Parallel()

		ss, _ := generateIDToken(user, privateKey, idExp)

		uFromToken, err := tokenServ.ValidateIDToken(ss)
		assert.NoError(t, err)

		assert.ElementsMatch(t,
			[]interface{}{user.Email, user.Name, user.UID, user.Website, user.ImageURL},
			[]interface{}{uFromToken.Email, uFromToken.Name, uFromToken.UID, uFromToken.Website, uFromToken.ImageURL},
		)
	})

	t.Run("Expired token", func(t *testing.T) {
		t.Parallel()

		ss, _ := generateIDToken(user, privateKey, -1) // expired 1 second ago

		expectedError := apperrors.NewAuthorization("Invalid ID Token")
		_, err := tokenServ.ValidateIDToken(ss)
		assert.EqualError(t, err, expectedError.Message)
	})

	t.Run("Invalid signature", func(t *testing.T) {
		t.Parallel()

		ss, _ := generateIDToken(user, privateKey, -1) // expired 1 second ago

		expectedError := apperrors.NewAuthorization("Invalid ID Token")
		_, err := tokenServ.ValidateIDToken(ss)
		assert.EqualError(t, err, expectedError.Message)
	})
}

func TestValidateRefreshToken(t *testing.T) {
	t.Parallel()

	var refreshExp int64 = 3 * 24 * 2600
	secret := "aNotSoRandomSecret"

	tokenServ := NewTokenService(&TSConfig{
		RefreshSecret:         secret,
		RefreshExpirationSecs: refreshExp,
	})

	// include password to make sure it is not serialized
	// since json tag is "-"
	uid, _ := uuid.NewRandom()
	user := &domain.User{
		UID:      uid,
		Email:    "bob@bob.com",
		Password: "atLeast8Chars",
	}

	t.Run("Valid token", func(t *testing.T) {
		t.Parallel()

		testRefreshToken, _ := generateRefreshToken(user.UID, secret, refreshExp)

		validatedRefreshToken, err := tokenServ.ValidateRefreshToken(testRefreshToken.SS)
		assert.NoError(t, err)

		assert.Equal(t, user.UID, validatedRefreshToken.UID)
		assert.Equal(t, testRefreshToken.SS, validatedRefreshToken.SS)
	})

	t.Run("Expired Token", func(t *testing.T) {
		t.Parallel()

		testRefreshToken, _ := generateRefreshToken(user.UID, secret, -1)

		expectedError := apperrors.NewAuthorization("Unable to verify user from refresh token")

		_, err := tokenServ.ValidateRefreshToken(testRefreshToken.SS)
		assert.EqualError(t, err, expectedError.Message)
	})
}
