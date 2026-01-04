package cmd

import (
	"fmt"
	"sort"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal/client"
)

func RunSearch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fave search <query>")
	}

	query := args[0]

	// Load configuration
	cfg, err := utils.LoadClientConfig(args[1:])
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	results, err := c.Search(query)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No bookmarks found matching query")
		return nil
	}

	// Extract and sort IDs
	ids := make([]int, 0, len(results))
	for id := range results {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Print results in sorted order
	for _, id := range ids {
		result := results[id]
		fmt.Printf("%d: %s (matched: %s)\n", id, result.Bookmark.Name, result.Match)
	}

	return nil
}
