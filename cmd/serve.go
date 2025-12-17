package cmd

import (
	"net/http"

	"github.com/t-eckert/fave/internal/server"
)

func RunServer(args []string) error {
	config := server.NewConfig("winona", "./data/bookmarks.json")
	server, err := server.New(config)
	if err != nil {
		return err
	}
	defer server.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /bookmarks", server.GetBookmarksHandler)
	mux.HandleFunc("GET /bookmarks/{id}", server.GetBookmarksByIDHandler)
	mux.HandleFunc("POST /bookmarks", server.PostBookmarksHandler)
	mux.HandleFunc("PUT /bookmarks/{id}", server.PutBookmarksHandler)
	mux.HandleFunc("DELETE /bookmarks/{id}", server.DeleteBookmarksHandler)

	return http.ListenAndServe(":8080", mux)
}
