// Right now we are not planning to change the models in the update but we can do in the future version
package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

func compareBitConfig(stored, current BitConfigFile) bool {
	if !reflect.DeepEqual(stored.Languages, current.Languages) {
		return true
	}
	if !reflect.DeepEqual(stored.Dependencies, current.Dependencies) {
		return true
	}
	return false
}

func DoUpdate() {
	bitconfig, err := LoadBitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dependencyfile = []string{}
	detectLanguage := iterateToFindLanguage()
	dep := TrackDependentFile()

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Cannot get the current directory:", err)
		os.Exit(1)
	}

	updated_bitconfig := BitConfigFile{
		ProjectName:                 filepath.Base(currentDir),
		Model:                       bitconfig.Model,
		Languages:                   detectLanguage,
		Dependencies:                dep,
		Initialized:                 bitconfig.Initialized,
		Dependencieschanged:         false,
		DependecyListthathasChanged: []string{},
	}

	if !compareBitConfig(bitconfig, updated_bitconfig) {
		fmt.Println("No changes detected. .bitconfig is already up to date.")
		return
	}

	updated_bitconfig.Dependencieschanged = true
	updated_bitconfig.DependecyListthathasChanged = changedDependencies(bitconfig.Dependencies, updated_bitconfig.Dependencies)

	data, err := json.MarshalIndent(updated_bitconfig, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(".bitconfig", data, 0644); err != nil {
		fmt.Println("Error writing to file:", err)
		os.Exit(1)
	}

	fmt.Println(".bitconfig updated successfully.")
	fmt.Println("Changed dependencies:")
	for _, item := range updated_bitconfig.DependecyListthathasChanged {
		fmt.Printf("  - %s\n", item)
	}
}

func changedDependencies(previous, updated map[string]int) []string {
	var changedList []string

	for name, updatedline := range updated {
		currentline, exist := previous[name]
		if !exist || currentline != updatedline {
			changedList = append(changedList, name)
		}
	}
	for name := range previous {
		if _, exist := updated[name]; !exist {
			changedList = append(changedList, name+" (removed)")
		}
	}

	return changedList
}
