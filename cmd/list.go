package cmd

import (
	"fmt"
	"sort"

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

	// Extract and sort IDs
	ids := make([]int, 0, len(bookmarks))
	for id := range bookmarks {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Print bookmarks in sorted order
	for _, id := range ids {
		bookmark := bookmarks[id]
		fmt.Println(utils.FormatBookmarkPreview(id, &bookmark))
	}

	return nil
}
