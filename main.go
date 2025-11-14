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
	rest := args[1:]

	var err error
	switch subcommand {
	case "serve":
		err = cmd.RunServer(rest)
	case "add":
		err = cmd.RunAdd(rest)
	case "list":
		err = cmd.RunList(rest)
	case "get":
		err = cmd.RunGet(rest)
	case "delete":
		err = cmd.RunDelete(rest)
	default:
		fmt.Println("Unknown subcommand:", subcommand)
	}
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
