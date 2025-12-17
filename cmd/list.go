package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal/client"
)

func RunList(args []string) error {
	// Load configuration
	cfg, err := LoadClientConfig(args)
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	bookmarks, err := c.List()
	if err != nil {
		return err
	}

	if len(bookmarks) == 0 {
		fmt.Println("No bookmarks found")
		return nil
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
