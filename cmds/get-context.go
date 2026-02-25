package cmds
import (
	"fmt"
	"log"
	"os"
	"goSummarizer/helper"
)
func get_context() {
	fmt.Println("Detecting the ReadMe.md and reading for context and after reading you can provide the context");


}
func detectReadMe() (string,error) {
	folder,err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot open the current directory")
		return "",err;
	}

}
