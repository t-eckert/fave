package cmd

import (
	"net/http"

	"github.com/t-eckert/fave/web"
)

func RunServer(args []string) error {
	mux := web.NewMux()
	return http.ListenAndServe(":8080", mux)
}
