package cmd

import (
	"net/http"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/web"
)

func RunServer(args []string) error {
	store := internal.NewStore()
	bookmarksServer := web.NewBookmarksServer(store)
	mux := web.NewMux(bookmarksServer)
	return http.ListenAndServe(":8080", mux)
}
