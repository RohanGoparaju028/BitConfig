package cmds

import "fmt"

func Help() {
	fmt.Println("Usage: bitconfig <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init          Initialize .bitconfig in the current directory")
	fmt.Println("  help          Show this help message")
	fmt.Println("  graph         Build and inspect the project knowledge graph")
	fmt.Println("  push-context  Send the knowledge graph to your terminal agent CLI")
	fmt.Println("  status        Show current project config and dependency snapshot")
	fmt.Println("  diff          Compare current project against saved .bitconfig")
	fmt.Println("  update        Save new languages and dependency changes to .bitconfig")
	fmt.Println()
	fmt.Println("Graph subcommands:")
	fmt.Println("  graph build   Scan project files and build knowledge_graph.json")
	fmt.Println("  graph show    Print nodes and connections in the terminal")
	fmt.Println("  graph note    Add developer notes to the knowledge graph")
}
