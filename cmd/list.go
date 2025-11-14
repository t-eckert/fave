package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/t-eckert/fave/internal"
)

func RunList(args []string) error {
	resp, err := http.Get(host + "/bookmarks")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bookmarks map[string]internal.Bookmark
	err = json.NewDecoder(resp.Body).Decode(&bookmarks)
	if err != nil {
		return err
	}

	for id, bookmark := range bookmarks {
		fmt.Println("id:", id)
		fmt.Println("name:", bookmark.Name)
		fmt.Println("url:", bookmark.Url)
		fmt.Println("description:", bookmark.Description)
		fmt.Println("tags:", bookmark.Tags)
		fmt.Println("---")
	}

	return nil
}
