package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/j03hanafi/hapalin-app/handler"
	"github.com/j03hanafi/hapalin-app/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
)

func main() {
	// Setup zap logger
	l := logger.Get()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	l.Info("Server is starting...")

	router := gin.New()

	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))

	handler.NewHandler(&handler.Config{
		R: router,
	})

	server := &http.Server{
		Handler: router,
		Addr:    ":8008",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal("Failed to initialize server: ",
				zap.Error(err),
			)
		}
	}()

	l.Info("Starting server",
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
	l.Info("Server is shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		l.Fatal("Server forced to shut down",
			zap.Error(err),
		)
	}
}
