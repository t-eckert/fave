package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t-eckert/fave/internal"
)

func BenchmarkGetBookmarks_Empty(b *testing.B) {
	mockStore := NewMockStore()
	// Create a mock testing.T for the helper
	srv := createTestServer(&testing.T{}, mockStore, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		srv.GetBookmarksHandler(w, req)
	}
}

func BenchmarkGetBookmarks_100Items(b *testing.B) {
	mockStore := NewMockStore()

	// Seed with 100 bookmarks
	bookmarks := make(map[int]internal.Bookmark, 100)
	for i := 1; i <= 100; i++ {
		bookmarks[i] = internal.Bookmark{
			Url:         "https://example.com",
			Name:        "Bookmark",
			Description: "Test bookmark",
			Tags:        []string{"test"},
		}
	}
	mockStore.Seed(bookmarks)

	srv := createTestServer(&testing.T{}, mockStore, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		srv.GetBookmarksHandler(w, req)
	}
}

func BenchmarkGetBookmarkByID(b *testing.B) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Test")})

	srv := createTestServer(&testing.T{}, mockStore, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/bookmarks/1", nil)
	req.SetPathValue("id", "1")

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		srv.GetBookmarkByIDHandler(w, req)
	}
}

func BenchmarkPostBookmarks(b *testing.B) {
	mockStore := NewMockStore()
	srv := createTestServer(&testing.T{}, mockStore, testConfig())

	bookmark := testBookmark("Test")
	body, _ := json.Marshal(bookmark)

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/bookmarks", bytes.NewReader(body))
		w := httptest.NewRecorder()
		srv.PostBookmarksHandler(w, req)
	}
}

func BenchmarkPutBookmarks(b *testing.B) {
	mockStore := NewMockStore()
	mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Original")})

	srv := createTestServer(&testing.T{}, mockStore, testConfig())

	updated := testBookmark("Updated")
	body, _ := json.Marshal(updated)

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPut, "/bookmarks/1", bytes.NewReader(body))
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()
		srv.PutBookmarksHandler(w, req)
	}
}

func BenchmarkDeleteBookmarks(b *testing.B) {
	mockStore := NewMockStore()

	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		// Re-add bookmark for each iteration
		mockStore.Seed(map[int]internal.Bookmark{1: testBookmark("Test")})
		srv := createTestServer(&testing.T{}, mockStore, testConfig())
		req := httptest.NewRequest(http.MethodDelete, "/bookmarks/1", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()
		b.StartTimer()

		srv.DeleteBookmarksHandler(w, req)
	}
}

// Benchmark concurrent requests
func BenchmarkConcurrentGets(b *testing.B) {
	mockStore := NewMockStore()
	bookmarks := make(map[int]internal.Bookmark, 100)
	for i := 1; i <= 100; i++ {
		bookmarks[i] = internal.Bookmark{
			Url:         "https://example.com",
			Name:        "Bookmark",
			Description: "Test bookmark",
			Tags:        []string{"test"},
		}
	}
	mockStore.Seed(bookmarks)

	srv := createTestServer(&testing.T{}, mockStore, testConfig())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/bookmarks", nil)
			w := httptest.NewRecorder()
			srv.GetBookmarksHandler(w, req)
		}
	})
}
