package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/domain"
	"github.com/j03hanafi/hapalin-app/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/logger"
	"go.uber.org/zap"
	"net/http"
)

// Me handler calls services for getting
// a user's details
func (h Handler) Me(c *gin.Context) {
	l := logger.Get()

	user, exists := c.Get("user")
	if !exists {
		l.Info("Unable to extract user from request context for unknown reason",
			zap.Any("Gin Context", c),
		)
		err := apperrors.NewInternal()
		c.JSON(err.Status(), gin.H{
			"error": err,
		})

		return
	}
	uid := user.(*domain.User).UID

	u, err := h.UserService.Get(c, uid)
	if err != nil {
		l.Info("Unable to find user",
			zap.String("user", uid.String()),
			zap.Error(err),
		)

		e := apperrors.NewNotFound("user", uid.String())

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": u,
	})
}
