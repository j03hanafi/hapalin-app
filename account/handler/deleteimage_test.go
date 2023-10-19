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

func TestDeleteImage(t *testing.T) {
	// setup
	l := logger.Get()
	gin.SetMode(gin.TestMode)

	// authorized user middleware
	uid, _ := uuid.NewRandom()
	ctxUser := &domain.User{
		UID: uid,
	}

	router := gin.New()
	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))

	// mock user service
	mockUserService := new(mocks.MockUserService)

	NewHandler(&Config{
		R:           router,
		UserService: mockUserService,
	})

	t.Run("Clear profile image error", func(t *testing.T) {
		rr := httptest.NewRecorder()

		clearProfileImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			ctxUser.UID,
		}

		errResp := apperrors.NewInternal()
		mockUserService.
			On("ClearProfileImage", clearProfileImageArgs...).
			Return(errResp)

		request, err := http.NewRequest(http.MethodDelete, "/deleteimage", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"error": errResp,
		})

		assert.Equal(t, apperrors.Status(errResp), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockUserService.AssertCalled(t, "ClearProfileImage", clearProfileImageArgs...)
	})

	t.Run("Success", func(t *testing.T) {
		rr := httptest.NewRecorder()

		// authorized user middleware - overwriting for unique mock arguments
		uid, _ := uuid.NewRandom()
		ctxUser := &domain.User{
			UID: uid,
		}

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		// mock user service
		mockUserService := new(mocks.MockUserService)

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		clearProfileImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			ctxUser.UID,
		}

		mockUserService.
			On("ClearProfileImage", clearProfileImageArgs...).
			Return(nil)

		request, err := http.NewRequest(http.MethodDelete, "/deleteimage", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockUserService.AssertCalled(t, "ClearProfileImage", clearProfileImageArgs...)
	})
}
