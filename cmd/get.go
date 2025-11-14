package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/t-eckert/fave/internal"
)

func RunGet(args []string) error {
	if len(args) < 1 {
		return nil
	}

	id := args[0]

	resp, err := http.Get(host + "/bookmarks/" + id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bookmark internal.Bookmark
	err = json.NewDecoder(resp.Body).Decode(&bookmark)
	if err != nil {
		return err
	}

	fmt.Println("id:", id)
	fmt.Println("name:", bookmark.Name)
	fmt.Println("url:", bookmark.Url)
	fmt.Println("description:", bookmark.Description)
	fmt.Println("tags:", bookmark.Tags)

	return nil
}
