package cmds

import "fmt"

func Help() {
	fmt.Println("Usage: bitconfig <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init          Initialize .bitconfig in the current directory")
	fmt.Println("  help          Show this help message")
	fmt.Println("  get-context   Summarize README and build context.txt for your terminal agent")
	fmt.Println("  push-context  Send context.txt to your configured terminal agent CLI")
	fmt.Println("  status        Show current project config and dependency snapshot")
	fmt.Println("  diff          Compare current project against saved .bitconfig")
	fmt.Println("  update        Save new languages and dependency changes to .bitconfig")
}
