package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ruziba3vich/soand/cmd/app"
)

func main() {
	// Initialize logger
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "[MusicLib] ", log.Ldate|log.Ltime|log.Lshortfile)

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
	logger.Printf("Received signal: %v, shutting down...\n", sig)

	// Ensure the context is properly canceled
	cancel()

	// Allow time for cleanup
	<-ctx.Done()
	logger.Println("Shutdown complete")
}
