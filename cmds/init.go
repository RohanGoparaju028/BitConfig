package cmd

import (
	"fmt"
	"io"
	"os"
)
var language = map[string][]string {
	    "typeScript" : {"package.json","package-lock.json","yarn.lock"},
		"javaScript" :  {"package.json","package-lock.json","yarn.lock"},
		"python" : {"requirments.txt","Pipefile","pyproject.toml"},
		"c#" :  {".csproj"},
		"java" :  {"pom.xml","build.gradle"},
		"go" : {"go.mod"},
		"ruby":{"Gemfile","Gemfile.lock"},
		"rust" :{"cargo.toml"},
};
type Initfile struct {
	Language     []string `json:"language"`     // For the field  that the devs are coding in
	Dependencies []string `json:"dependencies"` // The dependencies that are currently in use or installed
	Model        string   `json:"model" `        // LLM that the devolpers are using for the project

}
func isEmpty(file string) (bool,error) {
	f,err := os.Open(file)
	if err != nil {
		return false,err
	}
	defer f.Close()
	_,err = f.ReadDir(1)
	if err == io.EOF {
		return true,nil
	}
	if err != nil {
		return false,err
	}
	return false,nil
}
func iterateToFindLanguage() []string {
    f,err := os.Getwd()
	if err != nil {
		fmt.Println("Error while getting the current directoy and the error is: ",err)
		fmt.Println("Exiting by returning an empty string")
		return []string{}
	}
	flag,err := isEmpty(f)
	if err != nil {
		fmt.Println("Error while currently checking the current directory",err)
		return []string{}
	}
	if flag {
		panic("Current directory is empty no project is initialized")
	} else{

	}
}
func DoInit() {
	fmt.Print("We are initializing the BitConfig in the current directory\nplease choose the appropritate answers that best decribes your project\n")
	var llm string = "Ollama"
	if _, err := os.Stat(".bitconfig"); err != nil {
		fmt.Println(".bitconfig already exist in the working  directory")
		os.Exit(1)
	} else {
		fmt.Println("There is no .bitconfig file so writing all the dependencies")
		supportedLLM := []string{"Goose","Ollama", "Gemini"}
		var choice int
		fmt.Print("enter a choice between 0-2 for selectiing the model:")
		fmt.Scanf("%d", &choice)
		switch choice {
		case 0:
			llm = supportedLLM[0]
		case 1:
			llm = supportedLLM[1]
		case 2:
			llm = supportedLLM[2]
		default:
			fmt.Println("Enter a valid choice")
			os.Exit(1)
		}
		language := iterateToFindLanguage()
	}
}
