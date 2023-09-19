package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup zap logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Server is starting...")

	router := gin.Default()

	router.GET("/api/account", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hello": "world!",
		})
	})

	server := &http.Server{
		Handler: router,
		Addr:    ":8008",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to initialize server: ",
				zap.Error(err),
			)
		}
	}()

	logger.Info("Starting server",
		zap.String("address", server.Addr),
	)

	// Kill signal channel to shut down
	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// block until kill signal received
	<-quit

	// Context for informing server to close the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down server
	logger.Info("Server is shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shut down",
			zap.Error(err),
		)
	}
}
