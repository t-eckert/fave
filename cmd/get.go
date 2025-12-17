package cmd

import (
	"fmt"
	"strconv"

	"github.com/t-eckert/fave/internal/client"
)

func RunGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fave get [flags] <id>")
	}

	// Parse ID
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid bookmark ID: %w", err)
	}

	// Load configuration
	cfg, err := LoadClientConfig(args[1:])
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	bookmark, err := c.Get(id)
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
