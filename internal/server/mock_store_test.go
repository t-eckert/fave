package server_test

import (
	"errors"
	"maps"
	"sync"

	"github.com/t-eckert/fave/internal"
)

// MockStore implements StoreInterface for testing.
type MockStore struct {
	mu        sync.RWMutex
	bookmarks map[int]internal.Bookmark
	idCounter int

	// Hooks for testing error scenarios
	GetError          error
	AddError          error
	UpdateError       error
	DeleteError       error
	SaveSnapshotError error
}

func NewMockStore() *MockStore {
	return &MockStore{
		bookmarks: make(map[int]internal.Bookmark),
		idCounter: 0,
	}
}

func (m *MockStore) Get(id int) (internal.Bookmark, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetError != nil {
		return internal.Bookmark{}, m.GetError
	}

	bookmark, exists := m.bookmarks[id]
	if !exists {
		return internal.Bookmark{}, errors.New("bookmark not found")
	}

	return bookmark, nil
}

func (m *MockStore) List() map[int]internal.Bookmark {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy
	result := make(map[int]internal.Bookmark, len(m.bookmarks))
	maps.Copy(result, m.bookmarks)
	return result
}

func (m *MockStore) Add(bookmark internal.Bookmark) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.idCounter++
	m.bookmarks[m.idCounter] = bookmark
	return m.idCounter
}

func (m *MockStore) Update(id int, bookmark internal.Bookmark) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UpdateError != nil {
		return m.UpdateError
	}

	if _, exists := m.bookmarks[id]; !exists {
		return errors.New("bookmark not found")
	}

	m.bookmarks[id] = bookmark
	return nil
}

func (m *MockStore) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DeleteError != nil {
		return m.DeleteError
	}

	if _, exists := m.bookmarks[id]; !exists {
		return errors.New("bookmark not found")
	}

	delete(m.bookmarks, id)
	return nil
}

func (m *MockStore) SaveSnapshot() error {
	if m.SaveSnapshotError != nil {
		return m.SaveSnapshotError
	}
	return nil
}

// Helper methods for testing

func (m *MockStore) Seed(bookmarks map[int]internal.Bookmark) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bookmarks = make(map[int]internal.Bookmark, len(bookmarks))
	for k, v := range bookmarks {
		m.bookmarks[k] = v
		if k > m.idCounter {
			m.idCounter = k
		}
	}
}

func (m *MockStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.bookmarks)
}
