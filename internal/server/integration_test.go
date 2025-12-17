package server_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t-eckert/fave/internal"
)

// TestFullWorkflow tests a complete CRUD workflow
func TestFullWorkflow(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	// 1. List (empty)
	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
	w := httptest.NewRecorder()
	srv.GetBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List empty failed: %d", w.Code)
	}

	// 2. Create bookmark
	bookmark := testBookmark("Test")
	body, _ := json.Marshal(bookmark)
	req = httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader(body))
	w = httptest.NewRecorder()
	srv.PostBookmarksHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: %d", w.Code)
	}

	var createResp map[string]int
	json.NewDecoder(w.Body).Decode(&createResp)
	id := createResp["id"]

	// 3. Get by ID
	req = httptest.NewRequest(http.MethodGet, "/bookmarks/1", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()
	srv.GetBookmarkByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get by ID failed: %d", w.Code)
	}

	// 4. Update
	bookmark.Name = "Updated"
	body, _ = json.Marshal(bookmark)
	req = httptest.NewRequest(http.MethodPut, "/bookmarks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()
	srv.PutBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Update failed: %d", w.Code)
	}

	// 5. Verify update
	updated, _ := mockStore.Get(id)
	if updated.Name != "Updated" {
		t.Errorf("Update not applied: got '%s'", updated.Name)
	}

	// 6. Delete
	req = httptest.NewRequest(http.MethodDelete, "/bookmarks/1", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()
	srv.DeleteBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Delete failed: %d", w.Code)
	}

	// 7. Verify deletion
	if mockStore.Count() != 0 {
		t.Errorf("Bookmark not deleted")
	}
}

// TestWithAuthentication tests the full middleware chain with auth
func TestWithAuthentication(t *testing.T) {
	mockStore := NewMockStore()
	cfg := testConfig()
	cfg.AuthPassword = "secret123"
	srv := createTestServer(t, mockStore, cfg)

	// We need to test through the full middleware chain
	handler := srv.SetupRoutes()

	tests := []struct {
		name       string
		auth       string
		wantStatus int
	}{
		{
			name:       "no auth header",
			auth:       "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid auth",
			auth:       "Basic " + base64.StdEncoding.EncodeToString([]byte("user:wrongpass")),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid auth",
			auth:       "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret123")),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestHealthCheckNoAuth verifies health endpoint doesn't require auth
func TestHealthCheckNoAuth(t *testing.T) {
	mockStore := NewMockStore()
	cfg := testConfig()
	cfg.AuthPassword = "secret123" // Auth enabled
	srv := createTestServer(t, mockStore, cfg)

	handler := srv.SetupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// No Authorization header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Health check should work without auth even when auth is enabled
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d (health check should not require auth)", http.StatusOK, w.Code)
	}
}

// TestMultipleBookmarks tests creating and managing multiple bookmarks
func TestMultipleBookmarks(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	// Create 5 bookmarks
	for i := 1; i <= 5; i++ {
		bookmark := internal.Bookmark{
			Url:         "https://example.com",
			Name:        "Bookmark " + string(rune('0'+i)),
			Description: "Test bookmark",
			Tags:        []string{"test"},
		}
		body, _ := json.Marshal(bookmark)

		req := httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader(body))
		w := httptest.NewRecorder()
		srv.PostBookmarksHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Create bookmark %d failed: %d", i, w.Code)
		}
	}

	// List all
	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
	w := httptest.NewRecorder()
	srv.GetBookmarksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List failed: %d", w.Code)
	}

	var result map[int]internal.Bookmark
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 5 {
		t.Errorf("Expected 5 bookmarks, got %d", len(result))
	}
}

// TestCORSHeaders tests that CORS headers are set
func TestCORSHeaders(t *testing.T) {
	mockStore := NewMockStore()
	srv := createTestServer(t, mockStore, testConfig())

	handler := srv.SetupRoutes()

	// Test OPTIONS request (preflight)
	req := httptest.NewRequest(http.MethodOptions, "/bookmarks", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, w.Code)
	}

	// Check CORS headers
	if h := w.Header().Get("Access-Control-Allow-Origin"); h == "" {
		t.Error("Missing Access-Control-Allow-Origin header")
	}
	if h := w.Header().Get("Access-Control-Allow-Methods"); h == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	}
}
