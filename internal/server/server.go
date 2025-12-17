package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/store"
)

type Server struct {
	store *store.Store

	ticker *time.Ticker
	done   chan any
}

func New(config Config) (*Server, error) {
	store, err := store.NewStore(config.StoreFileName)
	if err != nil {
		return nil, err
	}

	s := Server{
		store:  store,
		ticker: time.NewTicker(1 * time.Second),
		done:   make(chan any),
	}

	// Run the loop to save snapshots to disk in the background.
	go s.storeSnapshotLoop()

	return &s, nil
}

func (s *Server) Close() {
	panic("NOT IMPLEMENTED")
}

func (s *Server) GetBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	j, err := json.Marshal(s.store.List())
	if err != nil {
		http.Error(w, "Error marshalling bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func (s *Server) GetBookmarksByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bookmark, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	j, err := json.Marshal(bookmark)
	if err != nil {
		http.Error(w, "Error marshalling bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func (s *Server) PostBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	var bookmark internal.Bookmark
	err := json.NewDecoder(r.Body).Decode(&bookmark)
	if err != nil {
		http.Error(w, fmt.Sprint("Invalid request payload", err.Error()), http.StatusBadRequest)
		return
	}

	if bookmark.Name == "" {
		http.Error(w, "Bookmark name is required", http.StatusBadRequest)
		return
	}

	id := s.store.Add(bookmark)

	w.Write([]byte(strconv.Itoa(id)))
}

func (s *Server) PutBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bookmark internal.Bookmark
	err = json.NewDecoder(r.Body).Decode(&bookmark)
	if err != nil {
		http.Error(w, fmt.Sprint("Invalid request payload", err.Error()), http.StatusBadRequest)
		return
	}

	err = s.store.Update(id, bookmark)
	if err != nil {
		http.Error(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(strconv.Itoa(id)))
}

func (s *Server) DeleteBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.store.Delete(id)
	if err != nil {
		http.Error(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(strconv.Itoa(id)))
}

func (s *Server) storeSnapshotLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.store.SaveSnapshot()
		case <-s.done:
			s.ticker.Stop()
			return
		}
	}
}
