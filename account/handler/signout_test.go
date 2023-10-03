package handler

import (
	"encoding/json"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSignOut(t *testing.T) {
	// Setup
	l := logger.Get()
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		uid, _ := uuid.NewRandom()
		ctxUser := &domain.User{
			UID:   uid,
			Email: "joe@joe.com",
		}

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))
		router.Use(func(c *gin.Context) {
			c.Set("user", ctxUser)
		})

		mockTokenService := new(mocks.MockTokenService)
		mockTokenService.
			On("SignOut", mock.AnythingOfType("context.backgroundCtx"), ctxUser.UID).
			Return(nil)

		NewHandler(&Config{
			R:            router,
			TokenService: mockTokenService,
		})

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		request, err := http.NewRequest(http.MethodPost, "/signout", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		respBody, _ := json.Marshal(gin.H{
			"message": "Successfully signed out",
		})

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
	})

	t.Run("error", func(t *testing.T) {

		uid, _ := uuid.NewRandom()
		ctxUser := &domain.User{
			UID:   uid,
			Email: "joe@joe.com",
		}

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))
		router.Use(func(c *gin.Context) {
			c.Set("user", ctxUser)
		})

		mockTokenService := new(mocks.MockTokenService)
		mockTokenService.
			On("SignOut", mock.AnythingOfType("context.backgroundCtx"), ctxUser.UID).
			Return(apperrors.NewInternal())

		NewHandler(&Config{
			R:            router,
			TokenService: mockTokenService,
		})

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		request, err := http.NewRequest(http.MethodPost, "/signout", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
