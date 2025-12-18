package cmd

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal/client"
)

func RunGet(args []string) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	output := fs.String("output", "text", "Output format: text or json")
	fs.String("o", "text", "Output format: text or json (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) < 1 {
		return fmt.Errorf("usage: fave get [flags] <id>")
	}

	// Handle shorthand -o flag
	if o := fs.Lookup("o").Value.String(); o != "text" {
		*output = o
	}

	// Parse ID
	id, err := strconv.Atoi(remaining[0])
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

	fmt.Println(utils.FormatBookmark(id, bookmark, *output))

	return nil
}
