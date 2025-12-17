package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/t-eckert/fave/internal"
)

type Server struct {
	config Config
	logger *slog.Logger
	store  StoreInterface

	// HTTP server
	httpServer *http.Server

	// Background snapshot goroutine
	ticker       *time.Ticker
	snapshotDone chan struct{}

	// Graceful shutdown
	shutdownOnce sync.Once
	shutdownErr  error
}

// New creates a new Server with the given configuration and store.
func New(config Config, store StoreInterface, logger *slog.Logger) (*Server, error) {
	if store == nil {
		return nil, fmt.Errorf("store cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Parse snapshot interval
	interval, err := time.ParseDuration(config.SnapshotInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid snapshot interval: %w", err)
	}

	s := &Server{
		config:       config,
		logger:       logger,
		store:        store,
		ticker:       time.NewTicker(interval),
		snapshotDone: make(chan struct{}),
	}

	// Create HTTP server with routes
	mux := s.SetupRoutes()
	s.httpServer = &http.Server{
		Addr:         config.Addr(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start background snapshot loop
	go s.snapshotLoop()

	logger.Info("server created",
		"addr", config.Addr(),
		"snapshot_interval", interval,
		"auth_enabled", config.AuthPassword != "",
	)

	return s, nil
}

// SetupRoutes configures all HTTP routes and middleware.
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("GET /bookmarks", s.GetBookmarksHandler)
	mux.HandleFunc("GET /bookmarks/{id}", s.GetBookmarkByIDHandler)
	mux.HandleFunc("POST /bookmarks", s.PostBookmarksHandler)
	mux.HandleFunc("PUT /bookmarks/{id}", s.PutBookmarksHandler)
	mux.HandleFunc("DELETE /bookmarks/{id}", s.DeleteBookmarksHandler)

	// Health check endpoint (no auth required)
	mux.HandleFunc("GET /health", s.HealthHandler)

	// Build middleware chain
	middlewares := []Middleware{
		RecoveryMiddleware(s.logger),
		LoggingMiddleware(s.logger),
		CORSMiddleware([]string{"*"}), // Allow all origins for personal project
	}

	// Add auth middleware if password is configured
	if s.config.AuthPassword != "" {
		middlewares = append(middlewares, BasicAuthMiddleware(s.config.AuthPassword, s.logger))
	}

	return Chain(mux, middlewares...)
}

// Start begins listening for HTTP requests (blocking).
func (s *Server) Start() error {
	s.logger.Info("starting server", "addr", s.config.Addr())

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	s.shutdownOnce.Do(func() {
		s.logger.Info("shutting down server")

		// Stop snapshot loop
		close(s.snapshotDone)
		s.ticker.Stop()

		// Final snapshot before shutdown
		s.logger.Info("saving final snapshot")
		if err := s.store.SaveSnapshot(); err != nil {
			s.logger.Error("failed to save final snapshot", "error", err)
			s.shutdownErr = fmt.Errorf("final snapshot: %w", err)
			return
		}

		// Shutdown HTTP server with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("http server shutdown error", "error", err)
			s.shutdownErr = fmt.Errorf("http shutdown: %w", err)
			return
		}

		s.logger.Info("server shutdown complete")
	})

	return s.shutdownErr
}

// snapshotLoop periodically saves store snapshots to disk.
func (s *Server) snapshotLoop() {
	s.logger.Debug("snapshot loop started")

	for {
		select {
		case <-s.ticker.C:
			if err := s.store.SaveSnapshot(); err != nil {
				s.logger.Error("snapshot save failed", "error", err)
			} else {
				s.logger.Debug("snapshot saved")
			}
		case <-s.snapshotDone:
			s.logger.Debug("snapshot loop stopped")
			return
		}
	}
}

// HTTP Handlers

func (s *Server) GetBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks := s.store.List()
	writeJSON(w, bookmarks, http.StatusOK)
}

func (s *Server) GetBookmarkByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	bookmark, err := s.store.Get(id)
	if err != nil {
		writeJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	writeJSON(w, bookmark, http.StatusOK)
}

func (s *Server) PostBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	var bookmark internal.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
		writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if bookmark.Name == "" {
		writeJSONError(w, "Bookmark name is required", http.StatusBadRequest)
		return
	}

	id := s.store.Add(bookmark)

	s.logger.Info("bookmark added", "id", id, "name", bookmark.Name)

	writeJSON(w, map[string]int{"id": id}, http.StatusCreated)
}

func (s *Server) PutBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	var bookmark internal.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
		writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := s.store.Update(id, bookmark); err != nil {
		writeJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	s.logger.Info("bookmark updated", "id", id)

	writeJSON(w, map[string]int{"id": id}, http.StatusOK)
}

func (s *Server) DeleteBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	if err := s.store.Delete(id); err != nil {
		writeJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	s.logger.Info("bookmark deleted", "id", id)

	writeJSON(w, map[string]int{"id": id}, http.StatusOK)
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "healthy"}, http.StatusOK)
}

// ============================================================================
// Helper functions for JSON responses
// ============================================================================

func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Can't change status code at this point, just log
		fmt.Printf("error encoding JSON: %v\n", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}
