package main

import (
	"github.com/t-eckert/fave/cmd"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fave <subcommand>")
		os.Exit(1)
	}

	args := os.Args[1:]
	subcommand := args[0]
	switch subcommand {
	case "serve":
		cmd.RunServer()
	}
}
