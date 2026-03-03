package cmds
import (
	"fmt"
	"os"
	"errors"
)
func DoUpdate() {
	file = ".bitconfig"
	if _,err := os.Open(file);errors.Is(err,os.ErrNotExist) {
		panic(".bitconfig doesnot exit in the directory,use init to initialize the bitconfig")
	}
	
}