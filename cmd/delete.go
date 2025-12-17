package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal/client"
)

func RunDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fave delete <id>")
	}

	id := args[0]

	client := client.New(host)
	delId, err := client.Delete(id)
	if err != nil {
		return err
	}
	fmt.Println(delId)

	return nil
}
