package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Bookmark struct {
	Url         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

var bookmarks = map[int]Bookmark{}

var bookmarksMutex = new(sync.RWMutex)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /bookmarks", getBookmarksHandler)
	mux.HandleFunc("GET /bookmarks/{id}", getBookmarksByIdHandler)
	mux.HandleFunc("POST /bookmarks", postBookmarksHandler)
	mux.HandleFunc("PUT /bookmarks/{id}", putBookmarksHandler)
	mux.HandleFunc("DELETE /bookmarks", deleteBookmarksHandler)

	http.ListenAndServe(":8080", mux)
}

func getBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	bookmarksMutex.RLock()
	defer bookmarksMutex.RUnlock()

	j, err := json.Marshal(bookmarks)
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

	bookmarksMutex.RLock()
	defer bookmarksMutex.RUnlock()

	bookmark, ok := bookmarks[id]
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
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()

	var bookmark Bookmark
	err := json.NewDecoder(r.Body).Decode(&bookmark)
	if err != nil {
		http.Error(w, fmt.Sprint("Invalid request payload", err.Error()), http.StatusBadRequest)
		return
	}

	if bookmark.Name == "" {
		http.Error(w, "Bookmark name is required", http.StatusBadRequest)
		return
	}

	id := len(bookmarks) + 1
	bookmarks[id] = bookmark

	w.Write([]byte(strconv.Itoa(id)))
}

func putBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()

	var bookmark Bookmark
	err = json.NewDecoder(r.Body).Decode(&bookmark)
	if err != nil {
		http.Error(w, fmt.Sprint("Invalid request payload", err.Error()), http.StatusBadRequest)
		return
	}

	bookmarks[id] = bookmark

	w.Write([]byte(strconv.Itoa(id)))
}

func deleteBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()

	delete(bookmarks, id)

	w.Write([]byte(strconv.Itoa(id)))
}
