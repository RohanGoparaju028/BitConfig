package cmds

import (
    "fmt"
    "os"
    "os/exec"
	"bufio"
)
func get_more_context() {
    fmt.Println("Enter your project description (Press Enter when done):")
        
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {
        userInput := scanner.Text()

        f, err := os.OpenFile("./context.txt", os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println("Error opening context file:", err)
            return
        }
        defer f.Close()
        formattedInput := "\n\n---  USER CONTEXT ---\n" + userInput + "\n"

        if _, err := f.WriteString(formattedInput); err != nil {
            fmt.Println("Error appending context:", err)
            return
        }
        
        fmt.Println("Context added to context.txt")
    }
}
func Get_Context() {
    cmd := exec.Command("deno", "run", "--allow-read", "--allow-write", "--allow-env", "Summarizer/summarize.ts")

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin 

    err := cmd.Run()
    if err != nil {
        fmt.Printf("Execution failed: %v\n", err)
        return
    }

    fmt.Println("A suummary of readme.md has been generated into the context.txt ")
	fmt.Print("Do u wanna add more context to the readme.md [choose 0 for exit 1 for enter more context]: ")
	var choice int
	fmt.Scanln(&choice);
	switch(choice) {
	  case 1: 
	      get_more_context()
		  break
	case 0: 
	     fmt.Println("okay if you are confident with the summarization of readme.txt you can use push-context to push the context to llm.")
		 break 
	default:
		fmt.Println("Not a valid option")
		break 
	}
}