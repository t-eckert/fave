package main

import (
	"github.com/t-eckert/fave/cmd"

	"fmt"
	"os"
)

const usage = `fave
A tiny CLI for saving your bookmarks.

Usage: fave <subcommand> [flags]

Available subcommands:
(Server)
	serve	Starts a Fave server to store and share bookmarks.
(Client)
	add	Add a bookmark.
	list	List all bookmarks.
	get	Get a bookmark by ID.
	update	Update an existing bookmark.
	delete	Delete a bookmark by ID.
	health	Check server health.

Common flags:
	--host		Server URL (default: http://localhost:8080)
	--password	Authentication password`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	args := os.Args[1:]

	subcommand := args[0]
	rest := args[1:]

	var err error
	switch subcommand {
	case "serve":
		err = cmd.RunServe(rest)
	case "add":
		err = cmd.RunAdd(rest)
	case "list":
		err = cmd.RunList(rest)
	case "get":
		err = cmd.RunGet(rest)
	case "update":
		err = cmd.RunUpdate(rest)
	case "delete":
		err = cmd.RunDelete(rest)
	case "health":
		err = cmd.RunHealth(rest)
	default:
		fmt.Println("Unknown subcommand:", subcommand)
	}
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
