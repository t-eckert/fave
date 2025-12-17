package server_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/server"
)

// Test Helpers

func createTestServer(t *testing.T, mockStore *MockStore, config server.Config) *server.Server {
	t.Helper()

	if mockStore == nil {
		mockStore = NewMockStore()
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Quiet during tests
	}))

	srv, err := server.New(config, mockStore, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Cleanup(func() {
		srv.Close()
	})

	return srv
}

func testConfig() server.Config {
	cfg := server.DefaultConfig()
	cfg.Port = "0" // Random port
	cfg.SnapshotInterval = "1h" // Long interval for tests
	cfg.AuthPassword = "" // No auth by default
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

// GET /bookmarks Tests

func TestGetBookmarks_Empty(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
	w := httptest.NewRecorder()

	srv.GetBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[int]internal.Bookmark
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty map, got %d items", len(result))
	}
}

func TestGetBookmarks_WithData(t *testing.T) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{
		1: testBookmark("First"),
		2: testBookmark("Second"),
	})

	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
	w := httptest.NewRecorder()

	srv.GetBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[int]internal.Bookmark
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 bookmarks, got %d", len(result))
	}
}

// GET /bookmarks/{id} Tests

func TestGetBookmarkByID_Success(t *testing.T) {
	mockStore := NewMockStore()
	bookmark := testBookmark("Test")
	mockStore.Seed(map[int]internal.Bookmark{1: bookmark})

	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	srv.GetBookmarkByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result internal.Bookmark
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", result.Name)
	}
}

func TestGetBookmarkByID_NotFound(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	srv.GetBookmarkByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetBookmarkByID_InvalidID(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks/invalid", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	srv.GetBookmarkByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// POST /bookmarks Tests

func TestPostBookmarks_Success(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	bookmark := testBookmark("New Bookmark")
	body, _ := json.Marshal(bookmark)

	req := httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.PostBookmarksHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var result map[string]int
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["id"] != 1 {
		t.Errorf("Expected id 1, got %d", result["id"])
	}

	if mockStore.Count() != 1 {
		t.Errorf("Expected 1 bookmark in store, got %d", mockStore.Count())
	}
}

func TestPostBookmarks_MissingName(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	bookmark := internal.Bookmark{Url: "https://example.com"} // No name
	body, _ := json.Marshal(bookmark)

	req := httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.PostBookmarksHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostBookmarks_InvalidJSON(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	srv.PostBookmarksHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// PUT /bookmarks/{id} Tests

func TestPutBookmarks_Success(t *testing.T) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Original")})

	srv := createTestServer(t, mockStore, testConfig())

	updated := testBookmark("Updated")
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/bookmarks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	srv.PutBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify update
	bookmark, _ := mockStore.Get(1)
	if bookmark.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got '%s'", bookmark.Name)
	}
}

func TestPutBookmarks_NotFound(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	updated := testBookmark("Updated")
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/bookmarks/999", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	srv.PutBookmarksHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPutBookmarks_InvalidID(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	updated := testBookmark("Updated")
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/bookmarks/invalid", bytes.NewReader(body))
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	srv.PutBookmarksHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutBookmarks_InvalidJSON(t *testing.T) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Original")})
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodPut, "/bookmarks/1", bytes.NewReader([]byte("invalid json")))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	srv.PutBookmarksHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// DELETE /bookmarks/{id} Tests

func TestDeleteBookmarks_Success(t *testing.T) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Test")})

	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/bookmarks/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	srv.DeleteBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if mockStore.Count() != 0 {
		t.Errorf("Expected 0 bookmarks after delete, got %d", mockStore.Count())
	}
}

func TestDeleteBookmarks_NotFound(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/bookmarks/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	srv.DeleteBookmarksHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteBookmarks_InvalidID(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/bookmarks/invalid", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	srv.DeleteBookmarksHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Health Check Tests

func TestHealth(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", result["status"])
	}
}
