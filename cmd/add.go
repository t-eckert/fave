package cmd

import (
	"fmt"

	"github.com/t-eckert/fave/internal"
	"github.com/t-eckert/fave/internal/client"
)

var host = "http://localhost:8080"

func RunAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fave add <name> <url>")
	}

	name := args[0]
	url := args[1]

	bookmark := internal.Bookmark{
		Url:         url,
		Name:        name,
		Description: "",
		Tags:        []string{},
	}

	client := client.New(host)

	id, err := client.Add(bookmark)
	if err != nil {
		return err
	}
	fmt.Println(id)

	return nil
}
