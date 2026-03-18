package cmds

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// walkDir recursively writes the file tree into context.txt
func walkDir(path string, depth int, f *os.File) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	for _, entry := range entries {
		name := entry.Name()
		// Skip irrelevant/hidden folders
		if name == ".git" || name == "node_modules" || name == ".DS_Store" || name == "deno.lock" {
			continue
		}
		if entry.IsDir() {
			f.WriteString(fmt.Sprintf("%s├── %s/\n", indent, name))
			walkDir(path+"/"+name, depth+1, f)
		} else {
			f.WriteString(fmt.Sprintf("%s├── %s\n", indent, name))
		}
	}
}

// appendFileTree appends the project file structure to context.txt
func appendFileTree() {
	f, err := os.OpenFile("./context.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Warning: Could not append file tree:", err)
		return
	}
	defer f.Close()
	f.WriteString("\n\n--- PROJECT FILE TREE ---\n")
	walkDir(".", 0, f)
	fmt.Println("File tree appended to context.txt")
}

// appendBitConfig appends the .bitconfig tech stack info to context.txt
func appendBitConfig() {
	data, err := os.ReadFile(".bitconfig")
	if err != nil {
		fmt.Println("Warning: Could not read .bitconfig (run 'init' first):", err)
		return
	}
	f, err := os.OpenFile("./context.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening context file:", err)
		return
	}
	defer f.Close()
	f.WriteString("\n\n--- PROJECT TECH STACK (.bitconfig) ---\n" + string(data) + "\n")
	fmt.Println("Tech stack from .bitconfig appended to context.txt")
}

// get_more_context prompts the user for multi-line input until they type DONE
func get_more_context() {
	fmt.Println("Enter your project description (type DONE on a new line when finished):")

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // TrimSpace handles Windows \r\n endings
		if line == "DONE" {
			break
		}
		lines = append(lines, line)
	}

	userInput := strings.Join(lines, "\n")

	f, err := os.OpenFile("./context.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening context file:", err)
		return
	}
	defer f.Close()

	formattedInput := "\n\n--- USER CONTEXT ---\n" + userInput + "\n"
	if _, err := f.WriteString(formattedInput); err != nil {
		fmt.Println("Error appending context:", err)
		return
	}
	fmt.Println("Context added to context.txt")
}

func Get_Context() {
	// Check if context.txt already exists — protect old context from being overwritten
	if _, err := os.Stat("./context.txt"); err == nil {
		fmt.Print("context.txt already exists. Overwrite it with a fresh README summary? [y/n]: ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Keeping existing context. You can still add more context below.")
			get_more_context()
			return
		}
	}

	// Run Deno to summarize README.md → context.txt (this overwrites context.txt)
	cmd := exec.Command("deno", "run", "--allow-read", "--allow-write", "--allow-env", "Summarizer/summarize.ts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Execution failed: %v\n", err)
		return
	}

	fmt.Println("A summary of README.md has been generated into context.txt")

	// Append tech stack from .bitconfig
	appendBitConfig()

	// Append project file tree
	appendFileTree()

	// Ask user if they want to add more manual context
	fmt.Print("\nDo you want to add more context? [0 = exit, 1 = add context]: ")
	var choice int
	fmt.Scanln(&choice)
	switch choice {
	case 1:
		get_more_context()
	case 0:
		fmt.Println("Done! Use push-context to send the context to your LLM.")
	default:
		fmt.Println("Not a valid option.")
	}
}
