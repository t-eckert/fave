package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

// BenchmarkAdd benchmarks adding a bookmark.
func BenchmarkAdd(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	bookmark := testBookmark("Test")

	b.ResetTimer()
	for b.Loop() {
		_, err := c.Add(bookmark)
		if err != nil {
			b.Fatalf("Add failed: %v", err)
		}
	}
}

// BenchmarkList_Empty benchmarks listing with no bookmarks.
func BenchmarkList_Empty(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[int]internal.Bookmark{})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := c.List()
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}

// BenchmarkList_100Items benchmarks listing with 100 bookmarks.
func BenchmarkList_100Items(b *testing.B) {
	// Create 100 test bookmarks
	bookmarks := make(map[int]internal.Bookmark, 100)
	for i := 1; i <= 100; i++ {
		bookmarks[i] = testBookmark("Bookmark " + string(rune(i)))
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bookmarks)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := c.List()
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}

// BenchmarkGet benchmarks getting a bookmark by ID.
func BenchmarkGet(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testBookmark("Test"))
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := c.Get(1)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

// BenchmarkUpdate benchmarks updating a bookmark.
func BenchmarkUpdate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	bookmark := testBookmark("Updated")

	b.ResetTimer()
	for b.Loop() {
		err := c.Update(1, bookmark)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkDelete benchmarks deleting a bookmark.
func BenchmarkDelete(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	for b.Loop() {
		err := c.Delete(1)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// BenchmarkHealth benchmarks health check.
func BenchmarkHealth(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	for b.Loop() {
		err := c.Health()
		if err != nil {
			b.Fatalf("Health failed: %v", err)
		}
	}
}

// BenchmarkConcurrentLists benchmarks concurrent list operations.
func BenchmarkConcurrentLists(b *testing.B) {
	bookmarks := make(map[int]internal.Bookmark, 100)
	for i := 1; i <= 100; i++ {
		bookmarks[i] = testBookmark("Bookmark " + string(rune(i)))
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bookmarks)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	c, err := client.New(cfg)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := c.List()
			if err != nil {
				b.Fatalf("List failed: %v", err)
			}
		}
	})
}
