package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/t-eckert/fave/internal"
)

type BookmarksServer struct {
	Store *internal.Store
}

func NewBookmarksServer(store *internal.Store) *BookmarksServer {
	return &BookmarksServer{
		Store: store,
	}
}

func NewMux(bookmarksServer *BookmarksServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /bookmarks", bookmarksServer.getBookmarksHandler)
	mux.HandleFunc("GET /bookmarks/{id}", bookmarksServer.getBookmarksByIdHandler)
	mux.HandleFunc("POST /bookmarks", bookmarksServer.postBookmarksHandler)
	mux.HandleFunc("PUT /bookmarks/{id}", bookmarksServer.putBookmarksHandler)
	mux.HandleFunc("DELETE /bookmarks/{id}", bookmarksServer.deleteBookmarksHandler)

	return mux
}

func (b *BookmarksServer) getBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	j, err := json.Marshal(b.Store.List())
	if err != nil {
		http.Error(w, "Error marshalling bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func (b *BookmarksServer) getBookmarksByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bookmark, err := b.Store.Get(id)
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

func (b *BookmarksServer) postBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	internal.BookmarksMutex.RLock()
	defer internal.BookmarksMutex.RUnlock()

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

	id := b.Store.Add(bookmark)

	w.Write([]byte(strconv.Itoa(id)))
}

func (b *BookmarksServer) putBookmarksHandler(w http.ResponseWriter, r *http.Request) {
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

	err = b.Store.Update(id, bookmark)
	if err != nil {
		http.Error(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(strconv.Itoa(id)))
}

func (b *BookmarksServer) deleteBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	internal.BookmarksMutex.Lock()
	defer internal.BookmarksMutex.Unlock()

	delete(internal.Bookmarks, id)

	w.Write([]byte(strconv.Itoa(id)))
}
