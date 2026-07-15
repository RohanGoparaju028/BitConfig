package main

import (
	"fmt"
	"os"

	L "github.com/RohanGoparaju028/BitConfig/cmds"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		fmt.Println("Initializing BitConfig in the current directory...")
		L.DoInit()
	case "help":
		L.Help()
	case "graph":
		handleGraph(os.Args[2:])
	case "get-context":
		fmt.Println("Building project knowledge graph...")
		L.Get_Context()
	case "push-context":
		fmt.Println("Sending knowledge graph to your terminal agent...")
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
		printUsage()
		os.Exit(1)
	}
}

func handleGraph(args []string) {
	if len(args) == 0 {
		fmt.Println("Building knowledge graph...")
		L.GraphBuild()
		return
	}

	switch args[0] {
	case "build":
		fmt.Println("Building knowledge graph...")
		L.GraphBuild()
	case "show":
		L.GraphShow()
	case "note":
		L.GraphNote()
	case "help":
		L.GraphHelp()
	default:
		fmt.Printf("Unknown graph subcommand: %s\n", args[0])
		L.GraphHelp()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: bitconfig <command>")
	fmt.Println("Run 'bitconfig help' to see available commands.")
}
