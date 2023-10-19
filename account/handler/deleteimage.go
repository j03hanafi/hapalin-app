package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"net/http"
)

// DeleteImage handler
func (h Handler) DeleteImage(c *gin.Context) {
	l := logger.Get()

	authUser := c.MustGet("user").(*domain.User)

	ctx := c.Request.Context()

	err := h.UserService.ClearProfileImage(ctx, authUser.UID)
	if err != nil {
		l.Error("Failed to delete profile image",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}
