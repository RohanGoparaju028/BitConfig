package cmds

import (
	"encoding/json" // used to convert our struct into a JSON file
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"slices"
)

// languageFiles is a map that connects a language name to its known files

var languageFiles = map[string][]string{
	"typeScript": {"package.json", "package-lock.json", "yarn.lock"},
	"javaScript": {"package.json", "package-lock.json", "yarn.lock"},
	"python":     {"requirements.txt", "Pipfile", "pyproject.toml"},
	"c#":     {".csproj"},
	"java":       {"pom.xml", "build.gradle"},
	"go":         {"go.mod"},
	"ruby":       {"Gemfile", "Gemfile.lock"},
	"rust":       {"Cargo.toml"},
}

// Config is the structure of what gets saved into .bitconfig/config.json

type BitConfigFile struct {
    ProjectName   string   `json:"project_name"`
	Languages     []string `json:"languages"`     // For the field  that the devs are coding in
	Dependencies map[string]int `json:"dependencies"` // The dependencies that are currently in use or installed
	Model        string   `json:"model" `        // LLM that the devolpers are using for the project
	Initialized   string   `json:"initialized at"` // The moment the file is created 
	Dependencieschanged bool `json:"is dependencies changed or not between now and the previous iterations"` // to check is there any dependecy change
	DependecyListthathasChanged []string `json:"list of dependencies that has changed"`
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
var dependencyfile = []string {}
func iterateToFindLanguage() []string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error while getting current directory:", err)
		return []string{}
	}

	empty, err := isEmpty(currentDir)
	if err != nil {
		fmt.Println("Error while checking directory:", err)
		return []string{}
	}
	if empty {
		fmt.Println("Current directory is empty. No project detected.")
		return []string{}
	}
    detected := map[string]bool{}
	var detectedLanguages []string
     err = filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()

		for lang, files := range languageFiles {
			for _, f := range files {
				if name == f {
					detected[lang] = true
					dependencyfile = append(dependencyfile, path)
					if(slices.Contains(detectedLanguages,lang)){
						continue;
					} else{
						detectedLanguages = append(detectedLanguages,lang)
					}
				}
			}
		}

		return nil
	})

	return detectedLanguages
}
func countLines(files string) (int,error) {
	file,err := os.Open(files)
	if err != nil {
		fmt.Println("Error opening the file",err)
		return 0,err;
	}
	defer file.Close()

  const bufferSize = 4096
  buffer := make([]byte, bufferSize)
  lineCount := 0
  totalBytesRead := 0
  var lastByte byte

  for {
    n, err := file.Read(buffer)
    if n > 0 {
      totalBytesRead += n
      lastByte = buffer[n-1]

      for i := 0; i < n; i++ {
        if buffer[i] == '\n' {
          lineCount++
        }
      }
    }

    if err == io.EOF {
      break
    }
    if err != nil {
      return 0, err
    }
  }

  if totalBytesRead > 0 && lastByte != '\n' {
    lineCount++
  }

  return lineCount, nil
}
func TrackDependentFile() map[string]int {
	dep := map[string]int{}
	for _,file := range dependencyfile {
		lines,err := countLines(file)
		if err != nil {
			fmt.Println("Error while processing the file... The error encountered was",err)
			os.Exit(1)
		}
		dep[file] = lines
	}
	return dep
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
	detectedLanguages := iterateToFindLanguage()
	dep := TrackDependentFile()
	
	// Build the Config struct with everything we collected above

	config := BitConfigFile{
		ProjectName:   projectName,
		Model:         selectedModel,
		Languages:     detectedLanguages,
		Dependencies:  dep,
		Initialized: time.Now().Format("2006-01-02 15:04:05"),
		Dependencieschanged:false, // always for the first time when we are building the .bitconfig there is no dependencies that are changed becauase that is the first time we are setting the bitcconfig file
		DependecyListthathasChanged:[]string{}, // Since there is no dependencies that are changed then the array is empty
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Could not prepare config data: %v\n", err)
		os.Exit(1)
	}

	// 0644 means the file is readable by everyone but only writable by owner
	if err := os.WriteFile(".bitconfig", data, 0644); err != nil {
		fmt.Printf("Could not save config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(".bitconfig created successfully")
}
