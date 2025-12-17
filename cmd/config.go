package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal/client"
)

// LoadClientConfig loads client configuration from all sources.
// This is a shared helper for all CLI commands that need a client.
func LoadClientConfig(args []string) (client.Config, error) {
	cfg, err := client.LoadConfig(args)
	if err != nil {
		return client.Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
