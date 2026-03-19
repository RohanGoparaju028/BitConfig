package cmds

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// skipList contains directory/file names that are noise for AI context:
// build artifacts, dependency caches, and generated output across many languages.
var skipList = map[string]bool{
	// Version control
	".git":       true,
	".gitignore": true,

	// macOS
	".DS_Store": true,

	// ── JavaScript / TypeScript / Node ──
	"node_modules": true,
	"deno.lock":    true,
	".next":        true,
	"dist":         true,
	".turbo":       true,

	// ── Go ──
	"vendor": true,

	// ── Java / Kotlin / Gradle ──
	".gradle": true,
	"gradle":  true, // gradle wrapper dir
	"build":   true, // Gradle build output
	"target":  true, // Maven build output
	".idea":   true, // IntelliJ
	"out":     true, // IntelliJ output

	// ── C# / .NET ──
	"bin":      true, // compiled binaries
	"obj":      true, // intermediate objects
	".vs":      true, // Visual Studio IDE dir
	"packages": true, // NuGet packages (legacy)

	// ── Python ──
	"__pycache__":   true,
	".venv":         true,
	"venv":          true,
	".mypy_cache":   true,
	".pytest_cache": true,
	"*.egg-info":    true,
	"dist-info":     true,

	// ── Rust ──
	// "target" is already covered above

	// ── Ruby ──
	".bundle": true,

	// ── PHP / Composer ──
	// "vendor" is already covered above

	// ── Swift / Xcode ──
	".build":        true, // Swift Package Manager
	"DerivedData":   true,
	"*.xcworkspace": true,
	"Pods":          true, // CocoaPods

	// ── Miscellaneous ──
	".terraform":  true, // Terraform
	"__MACOSX":    true, // macOS zip artifacts
	"coverage":    true, // test coverage output
	".nyc_output": true, // NYC coverage
	"logs":        true,
	"tmp":         true,
	"temp":        true,
}

// walkDir recursively writes the file tree into context.txt
func walkDir(path string, depth int, f *os.File) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	for _, entry := range entries {
		name := entry.Name()
		// Skip build artifacts, dependency dirs, and IDE folders
		if skipList[name] {
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
	walkDir(".", 2, f)
	fmt.Println("File tree appended to context.txt")
}

// get_more_context prompts the user for multi-line input until they type DONE
func get_more_context() {
	fmt.Println("Enter your project description (type DONE, Done, or done on a new line when finished):")

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // TrimSpace handles Windows \r\n endings
		if line == "DONE" || line == "Done" || line == "done" {
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

		// Backup the existing context in case the user wants to revert later
		data, err := os.ReadFile("./context.txt")
		if err == nil {
			os.WriteFile("./context.txt.bak", data, 0644)
			fmt.Println("Note: Your previous context has been backed up to context.txt.bak")
		}
	}

	// Check if README.md exists before attempting to summarize it
	if _, err := os.Stat("./README.md"); err == nil {
		// Run Deno to summarize README.md → context.txt (this overwrites context.txt)
		cmd := exec.Command("deno", "run", "--allow-read", "--allow-write", "--allow-env", "Summarizer/summarize.ts")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Warning: Failed to summarize README.md: %v\n", err)
			os.WriteFile("./context.txt", []byte(""), 0644) // ensure context.txt is reset
		} else {
			fmt.Println("A summary of README.md has been generated into context.txt")
		}
	} else {
		fmt.Println("No README.md found. Skipping README summarization.")
		os.WriteFile("./context.txt", []byte(""), 0644) // ensure context.txt is reset
	}

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
