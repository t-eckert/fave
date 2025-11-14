package internal

import (
	"errors"
	"maps"
	"sync"
)

type Store struct {
	Bookmarks  map[int]Bookmark `json:"bookmarks"`
	IdxCounter int              `json:"idx_counter"`

	mutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		Bookmarks:  make(map[int]Bookmark),
		IdxCounter: 0,
		mutex:      sync.RWMutex{},
	}
}

func (s *Store) Get(id int) (Bookmark, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bookmark, exists := s.Bookmarks[id]
	if !exists {
		return Bookmark{}, errors.New("bookmark not found")
	}

	return bookmark, nil
}

func (s *Store) List() map[int]Bookmark {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of bookmarks
	return maps.Clone(s.Bookmarks)
}

func (s *Store) Add(bookmark Bookmark) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.IdxCounter++
	s.Bookmarks[s.IdxCounter] = bookmark

	return s.IdxCounter
}

func (s *Store) Update(id int, bookmark Bookmark) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.Bookmarks[id]
	if !exists {
		return errors.New("bookmark not found")
	}

	s.Bookmarks[id] = bookmark
	return nil
}

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
