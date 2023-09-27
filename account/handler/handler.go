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
}

// Config will hold services that will eventually be injected into this
// handler layer on handler initialization
type Config struct {
	R               *gin.Engine
	UserService     domain.UserService
	TokenService    domain.TokenService
	BaseURL         string
	TimeoutDuration time.Duration
}

// NewHandler initializes the handler with required injected services along with http routes
// Does not return as it deals directly with a reference to the gin Engine
func NewHandler(c *Config) {
	// Create a handler (which will later have injected services)
	h := &Handler{
		UserService:  c.UserService,
		TokenService: c.TokenService,
	}

	// Create a group, or base url for all routes
	g := c.R.Group(c.BaseURL)

	if gin.Mode() != gin.TestMode {
		g.Use(middleware.Timeout(c.TimeoutDuration, apperrors.NewServiceUnavailable()))
		g.GET("/me", middleware.AuthUser(h.TokenService), h.Me)
	} else {
		g.GET("/me", h.Me)
	}

	g.POST("/signup", h.SignUp)
	g.POST("/signin", h.SignIn)
	g.GET("/signout", h.SignOut)
	g.GET("/tokens", h.Tokens)
	g.GET("/image", h.Image)
	g.GET("/deleteimage", h.DeleteImage)
	g.GET("/details", h.Details)
}

// SignOut handler
func (h Handler) SignOut(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's sign out",
	})
}

// Tokens handler
func (h Handler) Tokens(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's tokens",
	})
}

// Image handler
func (h Handler) Image(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's image",
	})
}

// DeleteImage handler
func (h Handler) DeleteImage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's delete image",
	})
}

// Details handler
func (h Handler) Details(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's details",
	})
}
