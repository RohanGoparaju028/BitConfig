package main

import (
	"fmt"
	"os"

	L "github.com/RohanGoparaju028/BitConfig/cmds"
)

func main() {
	fmt.Println("Thank you for using BitConfig,Please use the Init command to start \n or use the help command to start you with the tool")
	if len(os.Args) != 2 {
		panic("The command that you have entered is wrong the correct way is BitConfig <command name> \n to get to know the commands use BitConfig help")
	}
	command := os.Args[1]
	switch command {
	case "init":
		fmt.Println("Initializing the BitConfig in the current directory")
		L.DoInit()

	case "help":
		fmt.Println("Here are the commands that are supported and used in the CLI")
		L.Help()
	case "get-context":
		fmt.Println("Getting the context of the application that you are devolping and storing in the json file")
		L.Get_Context()
	case "push-context":
		_, err := os.Open(".bitconfig")
		if err != nil {
			panic(err)
		} else {
			fmt.Println("Reading the context.json and passing it to the tool")
		}
	case "diff":
		fmt.Println("The changes with the last change is")
		L.Diff()
	case "update":
		fmt.Println("Updatating the changes in the .bitconfig")
		L.DoUpdate()
	default:
		panic("Unsupported command please try a valid command or view BitConfig help command to see the comands that you wanna use ")
	}
}
