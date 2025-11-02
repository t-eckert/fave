package cmd

import (
	"net/http"

	"github.com/t-eckert/fave/web"
)

func RunServer() {
	mux := web.NewMux()
	http.ListenAndServe(":8080", mux)
}
