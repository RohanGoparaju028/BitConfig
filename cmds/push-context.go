package cmds

import (
	"bufio"
	"context"
	"log"
	"os"

	"cloud.google.com/go/vertexai/genai"
)

func getGeminiConnection() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, "")
	if err != nil {
		log.Fatal("Eror while establishing the connection to te gemini please try again")
	}
	defer client.Close()
	have_communication(ctx)
}
func getGooseConnection() {

}
func getOllamaConnection() {

}
func have_communication(ctx context.Context) {

}
func Push_Context() {
	file, err := os.Open(".bitconfig")
	if err != nil {
		panic(".bitconfig does not exist in the current directory please use init")
	}
	defer file.Close()
	Scanner := bufio.NewScanner(file)
	Scanner.Split(bufio.ScanWords)
	var model_name string
	for Scanner.Scan() {
		if Scanner.Text() == "model" {
			if Scanner.Scan() {
				model_name = Scanner.Text()
				break
			}

		}
	}
	switch model_name {
	case "Goose":
		getGooseConnection()
	case "Ollama":
		getOllamaConnection()
	case "Gemini":
		getGeminiConnection()

	}
}
