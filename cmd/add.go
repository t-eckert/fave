package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

func RunAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fave add [flags] <name> <url>")
	}

	// Load configuration
	cfg, err := utils.LoadClientConfig(args[2:])
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	name := args[0]
	url := args[1]

	bookmark := internal.NewBookmark(url, name, "", []string{})

	id, err := c.Add(bookmark)
	if err != nil {
		return err
	}

	fmt.Printf("Bookmark added with ID: %d\n", id)

	return nil
}
