package client_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

func testConfig(host string) client.Config {
	cfg := client.DefaultConfig()
	cfg.Host = host
	cfg.RetryAttempts = 0 // Disable retries for most tests
	return cfg
}

func testBookmark(name string) internal.Bookmark {
	return internal.Bookmark{
		Url:         "https://example.com",
		Name:        name,
		Description: "Test bookmark",
		Tags:        []string{"test"},
	}
}

// TestAdd_Success tests successful bookmark creation.
func TestAdd_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/bookmarks" {
			t.Errorf("Expected POST /bookmarks, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 42})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	id, err := c.Add(testBookmark("Test"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}
}

// TestAdd_WithAuth tests bookmark creation with authentication.
func TestAdd_WithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for auth header
		auth := r.Header.Get("Authorization")
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret123"))

		if auth != expectedAuth {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.Password = "secret123"
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.Add(testBookmark("Test"))
	if err != nil {
		t.Fatalf("Add with auth failed: %v", err)
	}
}

// TestList_Success tests successful bookmark listing.
func TestList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/bookmarks" {
			t.Errorf("Expected GET /bookmarks, got %s %s", r.Method, r.URL.Path)
		}

		bookmarks := map[int]internal.Bookmark{
			1: testBookmark("First"),
			2: testBookmark("Second"),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bookmarks)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	bookmarks, err := c.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(bookmarks) != 2 {
		t.Errorf("Expected 2 bookmarks, got %d", len(bookmarks))
	}
}

// TestGet_Success tests successful bookmark retrieval.
func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/bookmarks/42" {
			t.Errorf("Expected GET /bookmarks/42, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testBookmark("Test"))
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	bookmark, err := c.Get(42)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if bookmark.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", bookmark.Name)
	}
}

// TestGet_NotFound tests 404 error handling.
func TestGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bookmark not found"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.Get(999)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, client.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

// TestUpdate_Success tests successful bookmark update.
func TestUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/bookmarks/42" {
			t.Errorf("Expected PUT /bookmarks/42, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": 42})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	err = c.Update(42, testBookmark("Updated"))
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

// TestDelete_Success tests successful bookmark deletion.
func TestDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/bookmarks/42" {
			t.Errorf("Expected DELETE /bookmarks/42, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": 42})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	err = c.Delete(42)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// TestHealth_Success tests successful health check.
func TestHealth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Errorf("Expected GET /health, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	err = c.Health()
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

// TestHealth_Unhealthy tests unhealthy server response.
func TestHealth_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	err = c.Health()
	if err == nil {
		t.Fatal("Expected error for unhealthy status, got nil")
	}
}

// TestUnauthorized tests 401 error handling.
func TestUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.List()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, client.ErrUnauthorized) {
		t.Errorf("Expected ErrUnauthorized, got: %v", err)
	}
}

// TestBadRequest tests 400 error handling.
func TestBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.Add(testBookmark("Test"))
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, client.ErrBadRequest) {
		t.Errorf("Expected ErrBadRequest, got: %v", err)
	}
}

// TestInternalServerError tests 500 error handling.
func TestInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.List()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, client.ErrInternalServerError) {
		t.Errorf("Expected ErrInternalServerError, got: %v", err)
	}
}

// TestRetryLogic_EventualSuccess tests retry with eventual success.
func TestRetryLogic_EventualSuccess(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// Fail first 2 attempts, succeed on 3rd
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "Service unavailable"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[int]internal.Bookmark{})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.RetryAttempts = 3
	cfg.RetryDelay = 10 * time.Millisecond
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.List()
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestRetryLogic_AllFailed tests retry with all attempts failing.
func TestRetryLogic_AllFailed(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Service unavailable"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.RetryAttempts = 2
	cfg.RetryDelay = 10 * time.Millisecond
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.List()
	if err == nil {
		t.Fatal("Expected error after all retries, got nil")
	}

	// Should attempt initial + 2 retries = 3 total
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestRetryLogic_NoRetryOn4xx tests that 4xx errors are not retried.
func TestRetryLogic_NoRetryOn4xx(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad request"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.RetryAttempts = 3
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.Add(testBookmark("Test"))
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should not retry 4xx errors
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retries), got %d", attempts)
	}
}

// TestClientError_Wrapping tests error wrapping and unwrapping.
func TestClientError_Wrapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bookmark not found"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	_, err = c.Get(999)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Test error unwrapping
	if !errors.Is(err, client.ErrNotFound) {
		t.Errorf("Expected error to wrap ErrNotFound")
	}

	// Test error message contains context
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}
