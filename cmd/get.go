package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal/client"
)

func RunGet(args []string) error {
	if len(args) < 1 {
		return nil
	}

	id := args[0]

	client := client.New(host)
	bookmark, err := client.Get(id)
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
