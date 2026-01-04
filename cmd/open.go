package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal/client"
)

func RunOpen(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fave open [flags] <id>")
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

	// Fetch bookmark
	bookmark, err := c.Get(id)
	if err != nil {
		return err
	}

	// Open URL in browser
	fmt.Printf("Opening bookmark %d: %s\n", id, bookmark.Url)

	if err := openURL(bookmark.Url); err != nil {
		return fmt.Errorf("failed to open URL: %w", err)
	}

	return nil
}

// openURL opens a URL in the default browser (cross-platform).
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
