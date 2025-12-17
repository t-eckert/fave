package store_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/store"
)

// ============================================================================
// Test Utilities
// ============================================================================

// createTempStore creates a temporary store for testing.
// It automatically cleans up the temp file when the test completes.
func createTempStore(t *testing.T) (*store.Store, string) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "store-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	s, err := store.NewStore(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return s, tmpFile.Name()
}

// reloadStore reloads a store from disk to verify persistence.
func reloadStore(t *testing.T, filename string) *store.Store {
	t.Helper()
	s, err := store.NewStore(filename)
	if err != nil {
		t.Fatalf("Failed to reload store: %v", err)
	}
	return s
}

// assertBookmarkEqual checks deep equality of bookmarks.
func assertBookmarkEqual(t *testing.T, expected, actual internal.Bookmark) {
	t.Helper()
	if expected.Url != actual.Url ||
		expected.Name != actual.Name ||
		expected.Description != actual.Description ||
		!slicesEqual(expected.Tags, actual.Tags) {
		t.Errorf("Bookmarks not equal.\nExpected: %+v\nActual: %+v",
			expected, actual)
	}
}

// slicesEqual checks string slice equality.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// testBookmark creates a test bookmark with optional overrides.
func testBookmark(overrides ...func(*internal.Bookmark)) internal.Bookmark {
	b := internal.Bookmark{
		Url:         "https://example.com",
		Name:        "Example",
		Description: "Test bookmark",
		Tags:        []string{"test"},
	}
	for _, override := range overrides {
		override(&b)
	}
	return b
}

// ============================================================================
// Tests
// ============================================================================

// NewStore Tests

func TestNewStore_WithExistingEmptyFile(t *testing.T) {
	// Ensure the file exists
	f, err := os.CreateTemp("./", "existing-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	store, err := store.NewStore(f.Name())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if store.Bookmarks == nil {
		t.Fatal("Bookmarks map should not be nil")
	}

	if len(store.Bookmarks) != 0 {
		t.Fatalf("Expected empty bookmarks map, got %d items", len(store.Bookmarks))
	}

	if store.IdxCounter != 0 {
		t.Fatalf("Expected IdxCounter to be 0, got %d", store.IdxCounter)
	}
}

func TestNewStore_WithExistingFileWithData(t *testing.T) {
	// Create a temp file with existing bookmark data
	f, err := os.CreateTemp("./", "existing-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// Write test data to the file
	testData := struct {
		Bookmarks  map[int]internal.Bookmark `json:"bookmarks"`
		IdxCounter int                       `json:"idx_counter"`
	}{
		Bookmarks: map[int]internal.Bookmark{
			1: {
				Url:         "https://example.com",
				Name:        "Example",
				Description: "An example bookmark",
				Tags:        []string{"test", "example"},
			},
			2: {
				Url:         "https://github.com",
				Name:        "GitHub",
				Description: "Code hosting platform",
				Tags:        []string{"dev", "git"},
			},
		},
		IdxCounter: 2,
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	f.Close()

	// Load the store from the file
	store, err := store.NewStore(f.Name())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Verify the data was loaded correctly
	if store.Bookmarks == nil {
		t.Fatal("Bookmarks map should not be nil")
	}

	if len(store.Bookmarks) != 2 {
		t.Fatalf("Expected 2 bookmarks, got %d", len(store.Bookmarks))
	}

	if store.IdxCounter != 2 {
		t.Fatalf("Expected IdxCounter to be 2, got %d", store.IdxCounter)
	}

	// Verify specific bookmark data
	bookmark1, exists := store.Bookmarks[1]
	if !exists {
		t.Fatal("Bookmark with ID 1 should exist")
	}

	if bookmark1.Url != "https://example.com" {
		t.Errorf("Expected bookmark 1 URL to be 'https://example.com', got '%s'", bookmark1.Url)
	}

	if bookmark1.Name != "Example" {
		t.Errorf("Expected bookmark 1 Name to be 'Example', got '%s'", bookmark1.Name)
	}

	bookmark2, exists := store.Bookmarks[2]
	if !exists {
		t.Fatal("Bookmark with ID 2 should exist")
	}

	if bookmark2.Url != "https://github.com" {
		t.Errorf("Expected bookmark 2 URL to be 'https://github.com', got '%s'", bookmark2.Url)
	}
}

// Get Tests

func TestGet_ExistingBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	bookmark := testBookmark()
	id := s.Add(bookmark)

	result, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result)
}

func TestGet_NonExistentBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	_, err := s.Get(999)
	if err == nil {
		t.Fatal("Expected error for non-existent bookmark, got nil")
	}
}

func TestGet_AfterDelete(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark())
	err := s.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.Get(id)
	if err == nil {
		t.Fatal("Expected error when getting deleted bookmark, got nil")
	}
}

// List Tests

func TestList_EmptyStore(t *testing.T) {
	s, _ := createTempStore(t)

	bookmarks := s.List()
	if bookmarks == nil {
		t.Fatal("List() returned nil, expected empty map")
	}

	if len(bookmarks) != 0 {
		t.Fatalf("Expected empty map, got %d items", len(bookmarks))
	}
}

func TestList_MultipleBookmarks(t *testing.T) {
	s, _ := createTempStore(t)

	id1 := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "First" }))
	id2 := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Second" }))
	id3 := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Third" }))

	bookmarks := s.List()
	if len(bookmarks) != 3 {
		t.Fatalf("Expected 3 bookmarks, got %d", len(bookmarks))
	}

	if bookmarks[id1].Name != "First" {
		t.Errorf("Expected bookmark 1 Name='First', got '%s'", bookmarks[id1].Name)
	}
	if bookmarks[id2].Name != "Second" {
		t.Errorf("Expected bookmark 2 Name='Second', got '%s'", bookmarks[id2].Name)
	}
	if bookmarks[id3].Name != "Third" {
		t.Errorf("Expected bookmark 3 Name='Third', got '%s'", bookmarks[id3].Name)
	}
}

func TestList_ReturnsCopy(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Original" }))

	bookmarks := s.List()
	// Modify the returned map
	bookmarks[id] = testBookmark(func(b *internal.Bookmark) { b.Name = "Modified" })

	// Verify the store wasn't affected
	result, _ := s.Get(id)
	if result.Name != "Original" {
		t.Errorf("Expected Name='Original' (store should not be modified), got '%s'", result.Name)
	}
}

// Add Tests

func TestAdd_SingleBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	bookmark := testBookmark()
	id := s.Add(bookmark)

	if id != 1 {
		t.Fatalf("Expected first ID to be 1, got %d", id)
	}

	result, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result)
}

func TestAdd_MultipleBookmarks(t *testing.T) {
	s, _ := createTempStore(t)

	id1 := s.Add(testBookmark())
	id2 := s.Add(testBookmark())
	id3 := s.Add(testBookmark())

	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Fatalf("Expected sequential IDs 1,2,3, got %d,%d,%d", id1, id2, id3)
	}
}

func TestAdd_IDIncrement(t *testing.T) {
	s, _ := createTempStore(t)

	var ids []int
	for i := 0; i < 10; i++ {
		ids = append(ids, s.Add(testBookmark()))
	}

	// Verify all IDs are unique and sequential
	for i, id := range ids {
		expected := i + 1
		if id != expected {
			t.Errorf("Expected ID %d at position %d, got %d", expected, i, id)
		}
	}
}

func TestAdd_WithSpecialCharacters(t *testing.T) {
	s, _ := createTempStore(t)

	bookmark := testBookmark(func(b *internal.Bookmark) {
		b.Name = "Test 日本語 ñ © ® ™"
		b.Url = "https://example.com/パス?query=日本語"
		b.Description = "Unicode: 你好 Здравствуй مرحبا"
		b.Tags = []string{"日本語", "español", "русский"}
	})

	id := s.Add(bookmark)
	result, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result)
}

// Update Tests

func TestUpdate_ExistingBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Original" }))

	updated := testBookmark(func(b *internal.Bookmark) {
		b.Name = "Updated"
		b.Description = "New description"
	})

	err := s.Update(id, updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	result, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, updated, result)
}

func TestUpdate_NonExistentBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	err := s.Update(999, testBookmark())
	if err == nil {
		t.Fatal("Expected error when updating non-existent bookmark, got nil")
	}
}

func TestUpdate_Persistence(t *testing.T) {
	s, filename := createTempStore(t)

	id := s.Add(testBookmark())
	updated := testBookmark(func(b *internal.Bookmark) { b.Name = "Updated Name" })

	err := s.Update(id, updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Manually save snapshot (Update doesn't auto-save anymore)
	err = s.SaveSnapshot()
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Reload from disk
	s2 := reloadStore(t, filename)

	// Verify persisted
	bookmark, err := s2.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, updated, bookmark)
}

// Delete Tests

func TestDelete_ExistingBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark())

	err := s.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = s.Get(id)
	if err == nil {
		t.Fatal("Expected error when getting deleted bookmark, got nil")
	}
}

func TestDelete_NonExistentBookmark(t *testing.T) {
	s, _ := createTempStore(t)

	err := s.Delete(999)
	if err == nil {
		t.Fatal("Expected error when deleting non-existent bookmark, got nil")
	}
}

func TestDelete_NotInList(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark())
	err := s.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	bookmarks := s.List()
	if _, exists := bookmarks[id]; exists {
		t.Fatal("Deleted bookmark should not appear in List()")
	}
}

// Persistence Tests

func TestSaveSnapshot_BasicPersistence(t *testing.T) {
	s, filename := createTempStore(t)

	bookmark := testBookmark()
	id := s.Add(bookmark)

	err := s.SaveSnapshot()
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Reload from disk
	s2 := reloadStore(t, filename)

	result, err := s2.Get(id)
	if err != nil {
		t.Fatalf("Get failed after reload: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result)

	if s2.IdxCounter != s.IdxCounter {
		t.Errorf("Expected IdxCounter=%d after reload, got %d", s.IdxCounter, s2.IdxCounter)
	}
}

func TestSaveSnapshot_AtomicWrite(t *testing.T) {
	s, filename := createTempStore(t)

	s.Add(testBookmark())

	err := s.SaveSnapshot()
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Verify no temp files left behind
	dir := filepath.Dir(filename)
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "snapshot-") && strings.HasSuffix(file.Name(), ".json") {
			t.Errorf("Temporary snapshot file not cleaned up: %s", file.Name())
		}
	}
}

func TestReloadAfterAdd(t *testing.T) {
	s, filename := createTempStore(t)

	bookmark1 := testBookmark(func(b *internal.Bookmark) { b.Name = "First" })
	bookmark2 := testBookmark(func(b *internal.Bookmark) { b.Name = "Second" })

	id1 := s.Add(bookmark1)
	id2 := s.Add(bookmark2)

	s.SaveSnapshot()

	// Reload
	s2 := reloadStore(t, filename)

	// Verify both bookmarks
	result1, err := s2.Get(id1)
	if err != nil {
		t.Fatalf("Get bookmark 1 failed: %v", err)
	}
	assertBookmarkEqual(t, bookmark1, result1)

	result2, err := s2.Get(id2)
	if err != nil {
		t.Fatalf("Get bookmark 2 failed: %v", err)
	}
	assertBookmarkEqual(t, bookmark2, result2)
}

func TestReloadAfterUpdate(t *testing.T) {
	s, filename := createTempStore(t)

	id := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Original" }))

	// Update (no longer persists automatically)
	updated := testBookmark(func(b *internal.Bookmark) { b.Name = "Updated" })
	err := s.Update(id, updated)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Manually save
	s.SaveSnapshot()

	// Reload
	s2 := reloadStore(t, filename)

	result, err := s2.Get(id)
	if err != nil {
		t.Fatalf("Get failed after reload: %v", err)
	}

	assertBookmarkEqual(t, updated, result)
}

func TestReloadAfterDelete_WithSnapshot(t *testing.T) {
	s, filename := createTempStore(t)

	id := s.Add(testBookmark())
	s.SaveSnapshot()

	err := s.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	s.SaveSnapshot()

	// Reload
	s2 := reloadStore(t, filename)

	_, err = s2.Get(id)
	if err == nil {
		t.Fatal("Deleted bookmark should not exist after reload")
	}
}

func TestReloadAfterDelete_WithoutSnapshot(t *testing.T) {
	s, filename := createTempStore(t)

	bookmark := testBookmark()
	id := s.Add(bookmark)
	s.SaveSnapshot()

	// Delete but don't snapshot
	err := s.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Reload without saving
	s2 := reloadStore(t, filename)

	// Bookmark should still exist (delete wasn't persisted)
	result, err := s2.Get(id)
	if err != nil {
		t.Fatal("Expected bookmark to still exist (delete not persisted)")
	}

	assertBookmarkEqual(t, bookmark, result)
}

func TestNewStoreWithInvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid JSON
	tmpFile.WriteString("{invalid json content")
	tmpFile.Close()

	_, err = store.NewStore(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected error when loading invalid JSON, got nil")
	}
}

// Concurrency Tests

func TestConcurrent_MultipleReads(t *testing.T) {
	s, _ := createTempStore(t)

	// Add some test data
	id1 := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "First" }))
	id2 := s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Second" }))

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 50 concurrent readers for each bookmark
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := s.Get(id1)
				if err != nil {
					errors <- err
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := s.Get(id2)
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent read failed: %v", err)
	}
}

func TestConcurrent_ReadWrite(t *testing.T) {
	s, _ := createTempStore(t)

	// Add initial data
	id := s.Add(testBookmark())

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 10 concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := s.Get(id)
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	// 5 concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				err := s.Update(id, testBookmark(func(b *internal.Bookmark) {
					b.Name = fmt.Sprintf("Writer %d Update %d", n, j)
				}))
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

func TestConcurrent_MultipleAdds(t *testing.T) {
	s, _ := createTempStore(t)

	var wg sync.WaitGroup
	errors := make(chan error, 100)
	ids := make(chan int, 100)

	// 10 concurrent goroutines each adding 10 bookmarks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				id := s.Add(testBookmark(func(b *internal.Bookmark) {
					b.Name = fmt.Sprintf("Worker %d Bookmark %d", n, j)
				}))
				ids <- id
			}
		}(i)
	}

	wg.Wait()
	close(ids)

	// Verify all IDs are unique
	idMap := make(map[int]bool)
	for id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID detected: %d", id)
		}
		idMap[id] = true
	}

	// Should have exactly 100 unique IDs
	if len(idMap) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(idMap))
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent add failed: %v", err)
	}
}

func TestConcurrent_MultipleUpdates(t *testing.T) {
	s, _ := createTempStore(t)

	id := s.Add(testBookmark())

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	// 10 goroutines concurrently updating the same bookmark
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				err := s.Update(id, testBookmark(func(b *internal.Bookmark) {
					b.Name = fmt.Sprintf("Update from goroutine %d iteration %d", n, j)
				}))
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent update failed: %v", err)
	}

	// Verify bookmark still exists and is valid
	bookmark, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed after concurrent updates: %v", err)
	}

	if bookmark.Name == "" {
		t.Error("Bookmark name should not be empty after updates")
	}
}

func TestConcurrent_AddUpdateDelete(t *testing.T) {
	s, _ := createTempStore(t)

	// Seed with some data
	initialIDs := []int{
		s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Initial 1" })),
		s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Initial 2" })),
		s.Add(testBookmark(func(b *internal.Bookmark) { b.Name = "Initial 3" })),
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent adds
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				s.Add(testBookmark(func(b *internal.Bookmark) {
					b.Name = fmt.Sprintf("Adder %d Item %d", n, j)
				}))
			}
		}(i)
	}

	// Concurrent updates on initial data
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := initialIDs[n]
			for j := 0; j < 10; j++ {
				err := s.Update(id, testBookmark(func(b *internal.Bookmark) {
					b.Name = fmt.Sprintf("Updater %d Iteration %d", n, j)
				}))
				if err != nil {
					// OK if bookmark was deleted by another goroutine
					return
				}
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.List()
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent mixed operations failed: %v", err)
	}
}

// Edge Case Tests

func TestBookmarkWithEmptyFields(t *testing.T) {
	s, filename := createTempStore(t)

	bookmark := internal.Bookmark{
		Url:         "",
		Name:        "",
		Description: "",
		Tags:        []string{},
	}

	id := s.Add(bookmark)
	result, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result)

	// Verify persistence with empty fields
	s.SaveSnapshot()

	s2 := reloadStore(t, filename)
	result2, err := s2.Get(id)
	if err != nil {
		t.Fatalf("Get failed after reload: %v", err)
	}

	assertBookmarkEqual(t, bookmark, result2)
}

func TestLargeDataset(t *testing.T) {
	s, filename := createTempStore(t)

	const count = 1000

	// Add 1000 bookmarks
	for i := 0; i < count; i++ {
		s.Add(testBookmark(func(b *internal.Bookmark) {
			b.Name = fmt.Sprintf("Bookmark %d", i)
			b.Url = fmt.Sprintf("https://example.com/bookmark/%d", i)
		}))
	}

	// Verify count
	bookmarks := s.List()
	if len(bookmarks) != count {
		t.Fatalf("Expected %d bookmarks, got %d", count, len(bookmarks))
	}

	// Save and reload
	err := s.SaveSnapshot()
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	s2 := reloadStore(t, filename)
	bookmarks2 := s2.List()
	if len(bookmarks2) != count {
		t.Fatalf("Expected %d bookmarks after reload, got %d", count, len(bookmarks2))
	}

	// Verify counter
	if s2.IdxCounter != count {
		t.Errorf("Expected IdxCounter=%d after reload, got %d", count, s2.IdxCounter)
	}
}

func TestAdd_CounterPersistence(t *testing.T) {
	s, filename := createTempStore(t)

	// Add 5 bookmarks
	for i := 0; i < 5; i++ {
		id := s.Add(testBookmark())
		if id != i+1 {
			t.Errorf("Expected ID %d, got %d", i+1, id)
		}
	}

	s.SaveSnapshot()

	// Reload
	s2 := reloadStore(t, filename)

	// Next ID should be 6
	nextID := s2.Add(testBookmark())
	if nextID != 6 {
		t.Errorf("Expected next ID to be 6, got %d", nextID)
	}

	// Verify all 6 bookmarks exist
	bookmarks := s2.List()
	if len(bookmarks) != 6 {
		t.Fatalf("Expected 6 bookmarks, got %d", len(bookmarks))
	}
}
