package cmds

import (
	"encoding/json" // used to convert our struct into a JSON file
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// languageFiles is a map that connects a language name to its known files

var languageFiles = map[string][]string{
	"typeScript": {"package.json", "package-lock.json", "yarn.lock"},
	"javaScript": {"package.json", "package-lock.json", "yarn.lock"},
	"python":     {"requirements.txt", "Pipfile", "pyproject.toml"},
	"csharp":     {".csproj"},
	"java":       {"pom.xml", "build.gradle"},
	"go":         {"go.mod"},
	"ruby":       {"Gemfile", "Gemfile.lock"},
	"rust":       {"Cargo.toml"},
}

// Config is the structure of what gets saved into .bitconfig/config.json

type Config struct {
	ProjectName   string   `json:"project_name"`   // name of the project folder
	Model         string   `json:"model"`          // AI tool the developer is using
	Languages     []string `json:"languages"`      // detected programming languages
	Dependencies  []string `json:"dependencies"`   // dependency files found (go.mod, package.json etc)
	TrackedFiles  []string `json:"tracked_files"`  // config files BitConfig will watch
	InitializedAt string   `json:"initialized_at"` // timestamp of when init was run
}

// isEmpty checks if a directory has no files in it

func isEmpty(dirPath string) (bool, error) {
	// Open the directory
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}

	defer f.Close()

	_, err = f.ReadDir(1)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return false, nil
}

// iterateToFindLanguage walks through the current directory

func iterateToFindLanguage() ([]string, []string) {
	// Get the path of the current directory we are in
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error while getting current directory:", err)

		return []string{}, []string{}
	}

	// Check if the directory is completely empty

	empty, err := isEmpty(currentDir)
	if err != nil {
		fmt.Println("Error while checking directory:", err)
		return []string{}, []string{}
	}

	// If the folder is empty
	if empty {
		fmt.Println(" Current directory is empty. No project detected.")
		return []string{}, []string{}
	}

	// These will hold our results
	var detectedLanguages []string
	var foundDependencyFiles []string

	// Loop through every language in our map

	for language, files := range languageFiles {

		for _, file := range files {

			// os.Stat checks if a file exists
			// if err == nil → file EXISTS
			if _, err := os.Stat(file); err == nil {

				// Found a match! Add this language to our list
				detectedLanguages = append(detectedLanguages, language)

				// Also record which specific file we found
				foundDependencyFiles = append(foundDependencyFiles, file)

				break
			}
		}
	}

	return detectedLanguages, foundDependencyFiles
}

func findConfigFiles() []string {
	var found []string

	configExtensions := []string{".env", ".yaml", ".yml", ".toml", ".ini", ".conf"}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip folders we don't need
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == ".bitconfig" {
				return filepath.SkipDir
			}
			return nil
		}

		// Get the extension of this file
		ext := filepath.Ext(info.Name())

		// Loop through our list and check if the extension matches the list
		for _, validExt := range configExtensions {
			if ext == validExt {
				found = append(found, path)
				break // found a match, we can stop checking
			}
		}

		return nil
	})

	return found
}

// DoInit is the main function that runs when the developer types "bitconfig init".
// It gathers information about the project and saves it to .bitconfig/config.json
func DoInit() {
	fmt.Println(" Intializing BitConfig in the current directory...")
	fmt.Println("   Please answer a few questions about your project.")

	// Check if BitConfig is already initialized in this folder
	// os.Stat returns an error if the path does NOT exist
	// So if err == nil → .bitconfig folder ALREADY EXISTS → warn and stop
	if _, err := os.Stat(".bitconfig"); err == nil {
		fmt.Println("BitConfig is already initialized in this directory.")
		fmt.Println(" Run 'bitconfig update' to record new changes.")
		os.Exit(1)
	}

	// Get the project name automatically from the current folder name
	currentDir, _ := os.Getwd()
	projectName := filepath.Base(currentDir)
	fmt.Printf("\nProject detected: %s\n", projectName)
	// Ask the developer which AI tool they are using

	fmt.Println("\nWhich AI tool are you using for this project?")
	fmt.Println("  0 — Goose")
	fmt.Println("  1 — Ollama")
	fmt.Println("  2 — Gemini")

	fmt.Print("\nEnter a number (0-2): ")

	var modelChoice int
	fmt.Scanf("%d", &modelChoice)

	// List of supported models in the same order as shown above
	supportedModels := []string{"Goose", "Ollama", "Gemini"}

	// Validate that the number they typed is within the valid range
	if modelChoice < 0 || modelChoice > 2 {
		fmt.Println("Invalid choice. Please enter a number between 0 and 2.")
		os.Exit(1)
	}

	// Pick the model from the list using their choice as the index
	selectedModel := supportedModels[modelChoice]
	fmt.Printf("AI Model selected is : %s\n", selectedModel)

	// by looking for known files like go.mod, package.json etc
	fmt.Println("\n Detecting project language...")
	detectedLanguages, foundDependencyFiles := iterateToFindLanguage()

	// Tell the developer what we found
	if len(detectedLanguages) == 0 {
		fmt.Println(" Could not detect language automatically.")
		fmt.Println(" BitConfig looks for: go.mod, package.json, requirements.txt etc.")
	} else {
		fmt.Printf("Language(s) detected: %v\n", detectedLanguages)
		fmt.Printf("Dependency file(s) found: %v\n", foundDependencyFiles)
	}

	// Find config files in the project
	// These are the files BitConfig will watch for changes (like .env, .yaml)
	fmt.Println("\n Scanning for config files to track...")
	trackedFiles := findConfigFiles()

	if len(trackedFiles) == 0 {
		fmt.Println("No config files found.")
	} else {
		fmt.Printf(" Found %d config file(s):\n", len(trackedFiles))
		for _, file := range trackedFiles {
			fmt.Printf("   → %s\n", file)
		}
	}
	// Build the Config struct with everything we collected above

	config := Config{
		ProjectName:   projectName,
		Model:         selectedModel,
		Languages:     detectedLanguages,
		Dependencies:  foundDependencyFiles,
		TrackedFiles:  trackedFiles,
		InitializedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Create the .bitconfig folder on disk
	if err := os.MkdirAll(".bitconfig", 0755); err != nil {
		fmt.Printf("Could not create .bitconfig folder: %v\n", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Could not prepare config data: %v\n", err)
		os.Exit(1)
	}

	// 0644 means the file is readable by everyone but only writable by owner
	if err := os.WriteFile(".bitconfig/config.json", data, 0644); err != nil {
		fmt.Printf("Could not save config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n BitConfig initialized successfully!")
	fmt.Printf("   Project  : %s\n", projectName)
	fmt.Printf("   Model    : %s\n", selectedModel)
	fmt.Printf("   Language : %v\n", detectedLanguages)
	fmt.Printf("   Tracking : %d config file(s)\n", len(trackedFiles))
	fmt.Println("\n Next step: run 'bitconfig update' to record ")
}
