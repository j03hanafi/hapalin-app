package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/handler/middleware"
	"net/http"
	"time"
)

// Handler struct holds required services for handler to function
type Handler struct {
	UserService  domain.UserService
	TokenService domain.TokenService
	MaxBodyBytes int64
}

// Config will hold services that will eventually be injected into this
// handler layer on handler initialization
type Config struct {
	R               *gin.Engine
	UserService     domain.UserService
	TokenService    domain.TokenService
	BaseURL         string
	TimeoutDuration time.Duration
	MaxBodyBytes    int64
}

// NewHandler initializes the handler with required injected services along with http routes
// Does not return as it deals directly with a reference to the gin Engine
func NewHandler(c *Config) {
	// Create a handler (which will later have injected services)
	h := &Handler{
		UserService:  c.UserService,
		TokenService: c.TokenService,
		MaxBodyBytes: c.MaxBodyBytes,
	}

	// Create a group, or base url for all routes
	g := c.R.Group(c.BaseURL)

	if gin.Mode() != gin.TestMode {
		g.Use(middleware.Timeout(c.TimeoutDuration, apperrors.NewServiceUnavailable()))
		g.GET("/me", middleware.AuthUser(h.TokenService), h.Me)
		g.POST("/signout", middleware.AuthUser(h.TokenService), h.SignOut)
		g.PUT("/details", middleware.AuthUser(h.TokenService), h.Details)
		g.POST("/image", middleware.AuthUser(h.TokenService), h.Image)
	} else {
		g.GET("/me", h.Me)
		g.POST("/signout", h.SignOut)
		g.PUT("/details", h.Details)
		g.POST("/image", h.Image)
	}

	g.POST("/signup", h.SignUp)
	g.POST("/signin", h.SignIn)
	g.POST("/tokens", h.Tokens)
	g.GET("/deleteimage", h.DeleteImage)
}

// DeleteImage handler
func (h Handler) DeleteImage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's delete image",
	})
}
