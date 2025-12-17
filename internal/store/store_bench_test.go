package store_test

import (
	"os"
	"testing"

	"github.com/t-eckert/fave/internal/store"
)

// BenchmarkAdd_WithoutSnapshot benchmarks Add without snapshotting
func BenchmarkAdd_WithoutSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	b.ResetTimer()
	for b.Loop() {
		s.Add(testBookmark())
	}
}

// BenchmarkAdd_WithSnapshot benchmarks Add with snapshot on every call
func BenchmarkAdd_WithSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	b.ResetTimer()
	for b.Loop() {
		s.Add(testBookmark())
		s.SaveSnapshot()
	}
}

// BenchmarkDelete_WithoutSnapshot benchmarks Delete without snapshotting
func BenchmarkDelete_WithoutSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	// Pre-populate with a pool of bookmarks
	poolSize := 10000
	ids := make([]int, poolSize)
	for i := 0; i < poolSize; i++ {
		ids[i] = s.Add(testBookmark())
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		s.Delete(ids[i%poolSize])
		i++
	}
}

// BenchmarkDelete_WithSnapshot benchmarks Delete with snapshot on every call
func BenchmarkDelete_WithSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	// Pre-populate with a pool of bookmarks
	poolSize := 10000
	ids := make([]int, poolSize)
	for i := 0; i < poolSize; i++ {
		ids[i] = s.Add(testBookmark())
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		s.Delete(ids[i%poolSize])
		s.SaveSnapshot()
		i++
	}
}

// BenchmarkUpdate_CurrentBehavior benchmarks Update (which does snapshot)
func BenchmarkUpdate_CurrentBehavior(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	id := s.Add(testBookmark())

	b.ResetTimer()
	for b.Loop() {
		s.Update(id, testBookmark())
	}
}

// BenchmarkSaveSnapshot benchmarks just the snapshot operation
func BenchmarkSaveSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	// Add some data
	for i := 0; i < 100; i++ {
		s.Add(testBookmark())
	}

	b.ResetTimer()
	for b.Loop() {
		s.SaveSnapshot()
	}
}

// BenchmarkSaveSnapshot_LargeDataset benchmarks snapshot with 1000 bookmarks
func BenchmarkSaveSnapshot_LargeDataset(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	// Add 1000 bookmarks
	for i := 0; i < 1000; i++ {
		s.Add(testBookmark())
	}

	b.ResetTimer()
	for b.Loop() {
		s.SaveSnapshot()
	}
}

// BenchmarkMixedOperations_WithSnapshot simulates realistic workload with snapshotting
func BenchmarkMixedOperations_WithSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	b.ResetTimer()
	for b.Loop() {
		id := s.Add(testBookmark())
		s.SaveSnapshot()
		s.Update(id, testBookmark())
		s.SaveSnapshot()
		s.Delete(id)
		s.SaveSnapshot()
	}
}

// BenchmarkMixedOperations_WithoutSnapshot simulates realistic workload without snapshotting
func BenchmarkMixedOperations_WithoutSnapshot(b *testing.B) {
	s, filename := createBenchStore(b)
	defer os.Remove(filename)

	b.ResetTimer()
	for b.Loop() {
		id := s.Add(testBookmark())
		s.Update(id, testBookmark())
		s.Delete(id)
	}
}

// Helper function for benchmarks
func createBenchStore(b *testing.B) (*store.Store, string) {
	b.Helper()
	tmpFile, err := os.CreateTemp("", "bench-store-*.json")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	s, err := store.NewStore(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		b.Fatalf("Failed to create store: %v", err)
	}

	return s, tmpFile.Name()
}
