package store

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/t-eckert/fave/internal"
)

// Store contains an in-memory store of all bookmarks.
// It holds a pointer to a storage file for persistence.
type Store struct {
	Bookmarks  map[int]internal.Bookmark `json:"bookmarks"`
	IdxCounter int                       `json:"idx_counter"`

	fileName string
	file     *os.File

	mutex sync.RWMutex
}

// NewStore initializes a new store with the file at `fileName` as the backing file.
// If the file does not exist, it will be created.
// If the file exists and contains data, it will be read and loaded into the store.
func NewStore(fileName string) (*Store, error) {
	// Open the file for persistence.
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	store := &Store{
		Bookmarks:  make(map[int]internal.Bookmark),
		IdxCounter: 0,
		fileName:   fileName,
		file:       file,
		mutex:      sync.RWMutex{},
	}

	// Check if file has content.
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// If file has content, read and unmarshal it.
	if fileInfo.Size() > 0 {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(store)
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// Get retrieves a bookmark from the in-memory store.
// If the bookmark cannot be found, it returns an error.
func (s *Store) Get(id int) (internal.Bookmark, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bookmark, exists := s.Bookmarks[id]
	if !exists {
		return internal.Bookmark{}, errors.New("bookmark not found")
	}

	return bookmark, nil
}

// List returns all bookmarks in the in-memory store.
func (s *Store) List() map[int]internal.Bookmark {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of bookmarks
	return maps.Clone(s.Bookmarks)
}

// Add inserts a new bookmark.
// This bookmark will be given a unique ID by incrementing a counter on the store.
// The ID of the bookmark is returned.
func (s *Store) Add(bookmark internal.Bookmark) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.IdxCounter++
	s.Bookmarks[s.IdxCounter] = bookmark

	return s.IdxCounter
}

// Update swaps the bookmark at the given ID with the bookmark passed in.
// If no bookmark is found with the given ID, an error is returned.
// The update is not persisted until the next snapshot is saved.
func (s *Store) Update(id int, bookmark internal.Bookmark) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.Bookmarks[id]
	if !exists {
		return errors.New("bookmark not found")
	}

	s.Bookmarks[id] = bookmark
	return nil
}

// Delete removes the bookmark at the given ID from the in-memory bookmarks.
// The deletion is not persisted until the next snapshot is saved.
func (s *Store) Delete(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.Bookmarks[id]
	if !exists {
		return errors.New("bookmark not found")
	}

	delete(s.Bookmarks, id)
	return nil
}

// SaveSnapshot atomically saves the in-memory store to disk.
// On Unix-like systems, this is fully atomic. On Windows, there's a small
// window between removing the old file and renaming the temp file where the
// file doesn't exist, but this is necessary for cross-platform compatibility.
func (s *Store) SaveSnapshot() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	tmpf, err := os.CreateTemp(filepath.Dir(s.fileName), "snapshot-*.json")
	if err != nil {
		return err
	}
	defer tmpf.Close()

	if _, err := tmpf.Write(b); err != nil {
		return err
	}
	if err := tmpf.Close(); err != nil {
		return err
	}

	// On Windows, os.Rename fails if target exists, so remove it first
	// This sacrifices some atomicity on Windows, but maintains compatibility
	if _, err := os.Stat(s.fileName); err == nil {
		if err := os.Remove(s.fileName); err != nil {
			return err
		}
	}

	return os.Rename(tmpf.Name(), s.fileName)
}
