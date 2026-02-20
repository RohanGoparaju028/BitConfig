package cmd

import (
	"fmt"
	"os"
)

type Initfile struct {
	Language     []string ` json:"language"`     // For the field  that the devs are coding in
	Dependencies []string ` json:"dependencies"` // The dependencies that are currently in use or installed
	Model        string   `json:"model" `        // LLM that the devolpers are using for the project

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
		fmt.Print("enter a choice between 0-4 for selectiing the model:")
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

	}
}
