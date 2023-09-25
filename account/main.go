package main

import (
	"context"
	"errors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup zap logger
	l := logger.Get()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	l.Info("Server is starting...")

	ds, err := initDS()
	if err != nil {
		l.Fatal("Failed to initialize data sources: ",
			zap.Error(err),
		)
	}

	router, err := inject(ds)
	if err != nil {
		l.Fatal("Failed to initialize router: ",
			zap.Error(err),
		)
	}

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
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// block until kill signal received
	<-quit

	// Context for informing server to close the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down data sources
	if err := ds.close(); err != nil {
		l.Fatal("Failed to close data sources: ",
			zap.Error(err),
		)
	}

	// Shut down server
	l.Info("Server is shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		l.Fatal("Server forced to shut down",
			zap.Error(err),
		)
	}
}
