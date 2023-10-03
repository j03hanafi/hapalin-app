package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"net/http"
)

type detailsReq struct {
	Name    string `json:"name" binding:"omitempty,max=40"`
	Email   string `json:"email" binding:"omitempty,email"`
	Website string `json:"website" binding:"omitempty,url"`
}

// Details handler
func (h Handler) Details(c *gin.Context) {
	l := logger.Get()
	authUser := c.MustGet("user").(*domain.User)

	var req detailsReq

	if ok := bindData(c, &req); !ok {
		return
	}

	// Should be returned with current imageURL
	u := &domain.User{
		UID:     authUser.UID,
		Name:    req.Name,
		Email:   req.Email,
		Website: req.Website,
	}

	ctx := c.Request.Context()
	err := h.UserService.UpdateDetails(ctx, u)
	if err != nil {
		l.Error("Failed to update user",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": u,
	})
}
