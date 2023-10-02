package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"net/http"
)

type tokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Tokens handler
func (h Handler) Tokens(c *gin.Context) {
	l := logger.Get()

	// bind request to struct
	var req tokenReq

	if ok := bindData(c, &req); !ok {
		return
	}

	ctx := c.Request.Context()

	// verify refresh JWT token
	refreshToken, err := h.TokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		l.Error("Unable to verify refresh token",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	// get up-to-date user
	u, err := h.UserService.Get(ctx, refreshToken.UID)
	if err != nil {
		l.Error("Unable to get user",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	// create fresh pair of tokens
	tokens, err := h.TokenService.NewPairFromUser(ctx, u, refreshToken.ID.String())
	if err != nil {
		l.Error("Unable to create new token pair",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	// send back response
	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}
