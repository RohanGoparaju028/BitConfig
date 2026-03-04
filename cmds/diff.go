package cmds

import (
	"encoding/json"
	"fmt"
	"os"
)

func Diff() {
	// 1. Load the file
	file, _ := os.Open(".bitconfig")
	var config BitConfigFile
	json.NewDecoder(file).Decode(&config)

	// 2. Print the list that was saved during the update
	fmt.Println("Changes detected in last update:")
	for _, item := range config.DependecyListthathasChanged {
		fmt.Println(" -", item)
	}
}
