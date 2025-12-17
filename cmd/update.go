package cmd

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

func RunUpdate(args []string) error {
	// Parse command-specific flags
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	description := fs.String("description", "", "Bookmark description")
	fs.String("d", "", "Bookmark description (shorthand)")
	var tags utils.StringSlice
	fs.Var(&tags, "tag", "Tag (can be specified multiple times)")
	fs.Var(&tags, "t", "Tag (shorthand, can be specified multiple times)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get remaining args (id, name, url, and client config flags)
	remaining := fs.Args()
	if len(remaining) < 3 {
		return fmt.Errorf("usage: fave update [flags] <id> <name> <url>")
	}

	// Parse ID
	id, err := strconv.Atoi(remaining[0])
	if err != nil {
		return fmt.Errorf("invalid bookmark ID: %w", err)
	}

	name := remaining[1]
	url := remaining[2]

	// Handle shorthand -d flag
	if d := fs.Lookup("d").Value.String(); d != "" {
		*description = d
	}

	// Deduplicate tags
	uniqueTags := utils.DeduplicateStrings(tags)

	// Load client configuration from remaining args
	cfg, err := utils.LoadClientConfig(remaining[3:])
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

	err = c.Update(id, bookmark)
	if err != nil {
		return err
	}

	fmt.Printf("Bookmark %d updated\n", id)

	return nil
}
