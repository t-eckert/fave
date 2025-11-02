package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/t-eckert/fave/internal"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /bookmarks", getBookmarksHandler)
	mux.HandleFunc("GET /bookmarks/{id}", getBookmarksByIdHandler)
	mux.HandleFunc("POST /bookmarks", postBookmarksHandler)
	mux.HandleFunc("PUT /bookmarks/{id}", putBookmarksHandler)
	mux.HandleFunc("DELETE /bookmarks", deleteBookmarksHandler)

	return mux
}

func getBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	internal.BookmarksMutex.RLock()
	defer internal.BookmarksMutex.RUnlock()

	j, err := json.Marshal(internal.Bookmarks)
	if err != nil {
		http.Error(w, "Error marshalling bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func getBookmarksByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	internal.BookmarksMutex.RLock()
	defer internal.BookmarksMutex.RUnlock()

	bookmark, ok := internal.Bookmarks[id]
	if !ok {
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

func postBookmarksHandler(w http.ResponseWriter, r *http.Request) {
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

	id := len(internal.Bookmarks) + 1
	internal.Bookmarks[id] = bookmark

	w.Write([]byte(strconv.Itoa(id)))
}

func putBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	internal.BookmarksMutex.Lock()
	defer internal.BookmarksMutex.Unlock()

	var bookmark internal.Bookmark
	err = json.NewDecoder(r.Body).Decode(&bookmark)
	if err != nil {
		http.Error(w, fmt.Sprint("Invalid request payload", err.Error()), http.StatusBadRequest)
		return
	}

	internal.Bookmarks[id] = bookmark

	w.Write([]byte(strconv.Itoa(id)))
}

func deleteBookmarksHandler(w http.ResponseWriter, r *http.Request) {
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
