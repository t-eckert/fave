package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/t-eckert/fave/internal"
)

var host = "http://localhost:8080"

func RunAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fave add <name> <url>")
	}

	name := args[0]
	url := args[1]

	bookmark := internal.Bookmark{
		Url:         url,
		Name:        name,
		Description: "",
		Tags:        []string{},
	}

	j, err := json.Marshal(bookmark)
	if err != nil {
		return err
	}

	resp, err := http.Post(host+"/bookmarks", "application/json", bytes.NewReader(j))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	id, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(id))

	return nil
}
