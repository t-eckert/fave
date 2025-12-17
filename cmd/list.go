package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal/client"
)

func RunList(args []string) error {
	client := client.New(host)
	bookmarks, err := client.List()
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
