// Right now we are not planning to change the models in the update but we can do in the future version
package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

var changedDependenciesList []string

func compareBitConfig(bitconfig, updated_bitconfig BitConfigFile) bool {
	if !reflect.DeepEqual(bitconfig.Languages, updated_bitconfig.Languages) {
		return true
	}
	if !reflect.DeepEqual(bitconfig.Dependencies, updated_bitconfig.Dependencies) {
		return true
	}

	return false
}
func DoUpdate() {
	file, err := os.Open(".bitconfig")
	if err != nil {
		panic(".bitconfig doesnot exit in the directory,use init to initialize the bitconfig")
	}
	defer file.Close()
	var bitconfig BitConfigFile
	err = json.NewDecoder(file).Decode(&bitconfig) // Stores the current .bitconfig in the bitconfig structure
	var updated_bitconfig BitConfigFile
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Cannot get the project name")
		return
	}
	detectLanguage := iterateToFindLanguage()
	dep := TrackDependentFile()
	updated_bitconfig.ProjectName = currentDir
	updated_bitconfig.Model = bitconfig.Model
	updated_bitconfig.Languages = detectLanguage
	updated_bitconfig.Dependencies = dep
	ischanged := compareBitConfig(updated_bitconfig, bitconfig)
	if ischanged {
		updated_bitconfig.Dependencieschanged = true
		changedDependencies(bitconfig.Dependencies, updated_bitconfig.Dependencies)
		updated_bitconfig.DependecyListthathasChanged = changedDependenciesList
	} else {
		fmt.Println("There is no new update in the project")
		return
	}
	data, err := json.MarshalIndent(updated_bitconfig, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	err = os.WriteFile(file.Name(), data, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
func changedDependencies(previous, updated map[string]int) {
	for name, updatedline := range updated {
		currentline, exist := previous[name]
		if !exist || currentline != updatedline {
			changedDependenciesList = append(changedDependenciesList, name)
		}
	}
	for name := range previous {
		if _, exist := updated[name]; !exist {
			changedDependenciesList = append(changedDependenciesList, name+" removed from the list")
		}
	}
}
