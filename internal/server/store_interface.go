package server

import "github.com/t-eckert/fave/internal"

// StoreInterface defines the contract for bookmark storage operations.
// This interface allows for easier testing via mocks and decouples the
// server from the concrete store implementation.
type StoreInterface interface {
	// Get retrieves a bookmark by its ID.
	// Returns an error if the bookmark does not exist.
	Get(id int) (internal.Bookmark, error)

	// List returns all bookmarks in the store.
	// The returned map is keyed by bookmark ID.
	List() map[int]internal.Bookmark

	// Add creates a new bookmark and returns its assigned ID.
	Add(bookmark internal.Bookmark) int

	// Update modifies an existing bookmark.
	// Returns an error if the bookmark does not exist.
	Update(id int, bookmark internal.Bookmark) error

	// Delete removes a bookmark from the store.
	// Returns an error if the bookmark does not exist.
	Delete(id int) error

	// SaveSnapshot persists the current store state to disk.
	SaveSnapshot() error
}
