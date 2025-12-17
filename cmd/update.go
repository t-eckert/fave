package cmd

import (
	"fmt"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

func RunUpdate(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: fave update [flags] <id> <name> <url>")
	}

	// Parse ID
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid bookmark ID: %w", err)
	}

	// Load configuration
	cfg, err := utils.LoadClientConfig(args[3:])
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	name := args[1]
	url := args[2]

	bookmark := internal.Bookmark{
		Url:         url,
		Name:        name,
		Description: "",
		Tags:        []string{},
	}

	err = c.Update(id, bookmark)
	if err != nil {
		return err
	}

	fmt.Printf("Bookmark %d updated\n", id)

	return nil
}
