package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ruziba3vich/soand/cmd/app"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Channel to listen for termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run the application
	go func() {
		if err := app.Run(ctx, logger); err != nil {
			logger.Fatalf("Application failed: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	logger.Infof("Received signal: %v, shutting down...", sig)

	// Ensure the context is properly canceled
	cancel()

	// Allow time for cleanup
	<-ctx.Done()
	logger.Info("Shutdown complete")
}
