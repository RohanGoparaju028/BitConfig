package main

import (
	"fmt"
	"os"

	L "github.com/RohanGoparaju028/BitConfig/cmds"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: bitconfig <command>")
		fmt.Println("Run 'bitconfig help' to see available commands.")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		fmt.Println("Initializing BitConfig in the current directory...")
		L.DoInit()
	case "help":
		L.Help()
	case "get-context":
		fmt.Println("Building project context for your terminal agent...")
		L.Get_Context()
	case "push-context":
		fmt.Println("Sending context to your terminal agent...")
		L.PushContext()
	case "status":
		L.Status()
	case "diff":
		fmt.Println("Comparing project against .bitconfig...")
		L.Diff()
	case "update":
		fmt.Println("Updating .bitconfig with current project state...")
		L.DoUpdate()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'bitconfig help' to see available commands.")
		os.Exit(1)
	}
}
