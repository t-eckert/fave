package cmd

import (
	"flag"
	"fmt"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

func RunAdd(args []string) error {
	// Parse command-specific flags
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	description := fs.String("description", "", "Bookmark description")
	fs.String("d", "", "Bookmark description (shorthand)")
	var tags utils.StringSlice
	fs.Var(&tags, "tag", "Tag (can be specified multiple times)")
	fs.Var(&tags, "t", "Tag (shorthand, can be specified multiple times)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get remaining args (name, url, and client config flags)
	remaining := fs.Args()
	if len(remaining) < 2 {
		return fmt.Errorf("usage: fave add [flags] <name> <url>")
	}

	name := remaining[0]
	url := remaining[1]

	// Handle shorthand -d flag
	if d := fs.Lookup("d").Value.String(); d != "" {
		*description = d
	}

	// Deduplicate tags
	uniqueTags := utils.DeduplicateStrings(tags)

	// Load client configuration from remaining args
	cfg, err := utils.LoadClientConfig(remaining[2:])
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	bookmark := internal.NewBookmark(url, name, *description, uniqueTags)

	id, err := c.Add(bookmark)
	if err != nil {
		return err
	}

	fmt.Printf("Bookmark added with ID: %d\n", id)

	return nil
}
