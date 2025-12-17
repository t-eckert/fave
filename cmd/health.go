package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/cmd/utils"
	"github.com/t-eckert/fave/internal/client"
)

func RunHealth(args []string) error {
	// Load configuration
	cfg, err := utils.LoadClientConfig(args)
	if err != nil {
		return err
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()

	err = c.Health()
	if err != nil {
		return err
	}

	fmt.Println("Server is healthy")

	return nil
}
