package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal/client"
)

func RunList(args []string) error {
	// Load configuration
	cfg, err := utils.LoadClientConfig(args)
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
		fmt.Println(utils.FormatBookmark(id, &bookmark, "text"))
		fmt.Println("---")
	}

	return nil
}
