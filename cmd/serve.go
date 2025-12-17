package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/t-eckert/fave/internal/server"
	"github.com/t-eckert/fave/internal/store"
)

func RunServe(args []string) error {
	// Load configuration
	config, err := server.LoadConfig(args)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Setup logger
	logger := setupLogger(config)

	logger.Info("starting fave server",
		"version", "0.1.0",
		"addr", config.Addr(),
	)

	// Ensure store directory exists
	storeDir := filepath.Dir(config.StoreFileName)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("creating store directory: %w", err)
	}

	// Create store
	bookmarkStore, err := store.NewStore(config.StoreFileName)
	if err != nil {
		return fmt.Errorf("creating store: %w", err)
	}

	logger.Info("store loaded", "file", config.StoreFileName)

	// Create server
	srv, err := server.New(config, bookmarkStore, logger)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		if err := srv.Close(); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func setupLogger(config server.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: config.LogLevelValue(),
	}

	var handler slog.Handler
	if config.LogJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
