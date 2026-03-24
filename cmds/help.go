package cmds

import "fmt"

func Help() {
	fmt.Println("bitconfig always comes first then followed by command")
	fmt.Println("For first time users and for one who are refresing thier memory the following command and thier description is sperated by -> ")
	fmt.Println("bitconfig init -> Initialized an empty .bitconfig file in the current directory taking note of the language being used and initial dependencies along with the llm model")
	fmt.Println("bitconfig get-context -> looks at the dependencies and code creates a context of whats hapenning")
	fmt.Println("bitconfig diff -> shows the changes made in the project with respect to dependencies")
	fmt.Println("bitconfig push-context -> pushes the context to the llm")
	fmt.Println("bitconfig update ->updates the .bitconfig")
}
