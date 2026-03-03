package cmds;
import (
	"fmt"
	"os"
	"errors"
	"encoding/json"
	"log"
)
func Diff() {
	bitconfigfile := ".bitconfig"
	if _,err := os.Stat(bitconfigfile);errors.Is(err,os.ErrNotExist) {
		panic("BitConfig doesnot exits in the current directory please use init to initialize and then run the command.For more info please use help command")
	}
	file,err := os.Open(bitconfigfile)
	if err != nil {
		fmt.Println("Error while opening  the file")
		return;
	}
	defer file.Close()
	var configFile BitConfigFile;
	err = json.NewDecoder(file).Decode(&configFile)
	if err != nil {
		log.Fatal(err)
		return 
	}
	currentDependentLanguage := TrackDependentFile()
	hasChanged := false 
	deletedFiles :=  []string{}
	changedFile := []string{}
	addedFiles := []string{}
	for filepath,lines := range configFile.Dependencies {
		currentLine,err := countLines(filepath)
		if err != nil {
			fmt.Printf("Error while opening the dependencies file: %s",filepath)
			continue
		}
		if (currentLine - lines) == 0 {
			continue;
		} else {
			hasChanged = true;
			if currentLine - lines < 0 {
				if currentDependentLanguage[filepath] == 0 {
					deletedFiles = append(deletedFiles,filepath)
				}
			} else {
				changedFile = append(changedFile,filepath)
			}
		}
	}
	for currentPath := range currentDependentLanguage {
        if _, found := configFile.Dependencies[currentPath]; !found {
            hasChanged = true
            addedFiles = append(addedFiles, currentPath)
        }
    }

	if !hasChanged {
		fmt.Println("The project did not add any new dependencies")
	} else{
		fmt.Println("Project state has changed since last update:")
        if len(addedFiles) > 0 { fmt.Printf("  [+] Added: %v\n", addedFiles) }
        if len(deletedFiles) > 0 { fmt.Printf("  [-] Deleted: %v\n", deletedFiles) }
        if len(changedFile) > 0 { fmt.Printf("  [*] Modified: %v\n", changedFile) }
        configFile.Dependencieschanged = true
        configFile.DependecyListthathasChanged = append(addedFiles, changedFile...)
	}

}