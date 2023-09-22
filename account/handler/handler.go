package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/domain"
	"net/http"
	"os"
)

// Handler struct holds required services for handler to function
type Handler struct {
	UserService domain.UserService
}

// Config will hold services that will eventually be injected into this
// handler layer on handler initialization
type Config struct {
	R           *gin.Engine
	UserService domain.UserService
}

// NewHandler initializes the handler with required injected services along with http routes
// Does not return as it deals directly with a reference to the gin Engine
func NewHandler(c *Config) {
	// Create a handler (which will later have injected services)
	h := &Handler{
		UserService: c.UserService,
	}

	// Create a group, or base url for all routes
	g := c.R.Group(os.Getenv("ACCOUNT_API_URL"))

	g.GET("/me", h.Me)
	g.GET("/signup", h.SignUp)
	g.GET("/signin", h.SignIn)
	g.GET("/signout", h.SignOut)
	g.GET("/tokens", h.Tokens)
	g.GET("/image", h.Image)
	g.GET("/deleteimage", h.DeleteImage)
	g.GET("/details", h.Details)
}

// SignUp handler
func (h Handler) SignUp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's sign up",
	})
}

// SignIn handler
func (h Handler) SignIn(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hello": "it's sign in",
	})
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
