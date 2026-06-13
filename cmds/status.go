package cmds

import (
	"fmt"
	"os"
	"strings"
)

func Status() {
	config, err := LoadBitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("BitConfig project status")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Project:      %s\n", config.ProjectName)
	fmt.Printf("Initialized:  %s\n", config.Initialized)
	fmt.Printf("Terminal Agent: %s\n", config.Model)
	if config.AgentModel != "" {
		fmt.Printf("Agent Model:    %s\n", config.AgentModel)
	}
	fmt.Printf("Languages:    %s\n", strings.Join(config.Languages, ", "))
	fmt.Printf("Dependencies: %d tracked file(s)\n", len(config.Dependencies))

	for file, lines := range config.Dependencies {
		fmt.Printf("  - %s (%d lines)\n", file, lines)
	}

	if config.Dependencieschanged {
		fmt.Println("\nLast update detected dependency changes:")
		for _, item := range config.DependecyListthathasChanged {
			fmt.Printf("  - %s\n", item)
		}
	} else {
		fmt.Println("\nNo dependency changes recorded since init.")
	}

	if _, err := os.Stat("./context.txt"); err == nil {
		fmt.Println("context.txt exists — run push-context to send it to your terminal agent.")
	} else {
		fmt.Println("\nNo context.txt yet — run get-context to generate project context.")
	}
}
