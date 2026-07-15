package cmds

import "fmt"

func Get_Context() {
	fmt.Println("get-context now builds the knowledge graph.")
	fmt.Println("Use: bitconfig graph build")
	fmt.Println()
	GraphBuild()
}
