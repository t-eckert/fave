package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
	"gopkg.in/yaml.v3"
)

func RunEdit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fave edit <id>")
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

	// Get the current bookmark
	bookmark, err := c.Get(id)
	if err != nil {
		return err
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(bookmark)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmark to YAML: %w", err)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "fave-edit-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write YAML to temp file
	if _, err := tmpFile.Write(yamlData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Get editor from environment, default to vi
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Open editor
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run editor: %w", err)
	}

	// Read the edited file
	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Unmarshal YAML
	var editedBookmark internal.Bookmark
	if err := yaml.Unmarshal(editedData, &editedBookmark); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Validate required fields
	if editedBookmark.Url == "" {
		return fmt.Errorf("validation error: url is required")
	}
	if editedBookmark.Name == "" {
		return fmt.Errorf("validation error: name is required")
	}

	// Update the bookmark
	if err := c.Update(id, editedBookmark); err != nil {
		return err
	}

	fmt.Printf("Bookmark %d updated\n", id)

	return nil
}
