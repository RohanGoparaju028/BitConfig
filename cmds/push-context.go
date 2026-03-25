package cmds

import (
	"encoding/json"
	"errors"
	"os"
)

func PushContext() {
	filename := ".bitconfig"

	if _, err := os.Stat(filename); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic(".bitconfig not exist in the current directory please use init or use help to see commands")
		}
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		panic("Error while opening the file")
	}
	var jsonmap map[string]interface{}
	if err := json.Unmarshal(file, &jsonmap); err != nil {
		panic("Error while reading .bitconfig")
	}
	var model_name string
	if val, ok := jsonmap["model"]; ok {
		model_name = val.(string)
	}

}
