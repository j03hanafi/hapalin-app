package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"net/http"
)

// Image handler
func (h Handler) Image(c *gin.Context) {
	l := logger.Get()
	authUser := c.MustGet("user").(*domain.User)

	// limit overly large request body
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.MaxBodyBytes)

	imageFileHeader, err := c.FormFile("imageFile")

	// check for error before checking for non-nil header
	if err != nil {
		l.Error("Unable to parse mulipart/form-data",
			zap.Error(err),
		)

		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Max request body size is %v bytes\n", h.MaxBodyBytes),
			})
			return
		}

		e := apperrors.NewBadRequest("Unable to parse multipart/form-data")
		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	if imageFileHeader == nil {
		e := apperrors.NewBadRequest("Missing image file")
		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	mimeType := imageFileHeader.Header.Get("Content-Type")

	// Validate image mime-type is allowable
	if !isAllowedImageType(mimeType) {
		l.Error("Invalid image type",
			zap.String("mime_type", mimeType),
			zap.Error(err),
		)
		e := apperrors.NewBadRequest("imageFile must be 'image/jpeg' or 'image/png'")
		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	ctx := c.Request.Context()

	updatedUser, err := h.UserService.SetProfileImage(ctx, authUser.UID, imageFileHeader)
	if err != nil {
		l.Error("Unable to set profile image",
			zap.Error(err),
		)
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imageUrl": updatedUser.ImageURL,
		"message":  "Profile image updated successfully",
	})

}
