package cmd

import (
	"fmt"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
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

	bookmark, err := c.Get(id)
	if err != nil {
		return err
	}

	fmt.Println(utils.FormatBookmark(id, bookmark))

	return nil
}
