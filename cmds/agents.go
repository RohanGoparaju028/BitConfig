package cmds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const contextPrompt = "You have been given a project knowledge graph below. Use the nodes and connections to understand how files, languages, docs, and dependencies relate. Suggest what to focus on next."

var supportedProviders = map[string]bool{
	"Ollama":  true,
	"Claude":  true,
	"ChatGPT": true,
	"Gemini":  true,
}

var legacyProviderNames = map[string]string{
	"Claude Code":  "Claude",
	"Gemini CLI":   "Gemini",
	"ChatGpt":      "ChatGPT",
	"Cursor Agent": "ChatGPT",
	"Aider":        "ChatGPT",
}

func resolveProvider(name string) (string, error) {
	if supportedProviders[name] {
		return name, nil
	}
	if mapped, ok := legacyProviderNames[name]; ok {
		return mapped, nil
	}
	switch strings.ToLower(name) {
	case "ollama":
		return "Ollama", nil
	case "claude", "anthropic":
		return "Claude", nil
	case "chatgpt", "openai":
		return "ChatGPT", nil
	case "gemini", "google":
		return "Gemini", nil
	}
	return "", fmt.Errorf("unknown provider '%s' — re-run 'bitconfig init' to pick an AI provider (Ollama, Claude, ChatGPT, Gemini)", name)
}

func defaultModelForProvider(provider string) string {
	switch provider {
	case "Ollama":
		return "llama3.2"
	case "Claude":
		return "claude-3-5-sonnet-latest"
	case "ChatGPT":
		return "gpt-4o"
	case "Gemini":
		return "gemini-1.5-flash"
	default:
		return ""
	}
}

func PushContext() {
	loadEnvFile()

	config, err := LoadBitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	graph, err := LoadKnowledgeGraph()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(graph.Nodes) == 0 {
		fmt.Println("Knowledge graph is empty. Run 'bitconfig get-context' first.")
		os.Exit(1)
	}

	provider, err := resolveProvider(config.Model)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	model := config.AgentModel
	if model == "" {
		model = defaultModelForProvider(provider)
	}

	if err := validateProviderAuth(provider); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	payload := graph.ToAgentPayload(config)

	fmt.Printf("Sending knowledge graph to %s (%s)...\n\n", provider, model)
	if err := streamToAI(provider, model, payload); err != nil {
		fmt.Printf("\nFailed to connect to %s: %v\n", provider, err)
		os.Exit(1)
	}
	fmt.Println()
}

func validateProviderAuth(provider string) error {
	switch provider {
	case "Ollama":
		conn, err := net.DialTimeout("tcp", "127.0.0.1:11434", 800*time.Millisecond)
		if err != nil {
			conn, err = net.DialTimeout("tcp", "localhost:11434", 800*time.Millisecond)
		}
		if err != nil {
			return fmt.Errorf("Ollama service does not appear to be running at localhost:11434.\nPlease make sure Ollama is started ('ollama serve') before running this command")
		}
		conn.Close()
		return nil

	case "Claude":
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			return nil
		}
		return fmt.Errorf("ANTHROPIC_API_KEY is not set.\nPlease set it via:\n  export ANTHROPIC_API_KEY=\"your_key\"\nor add ANTHROPIC_API_KEY=your_key in .env")

	case "ChatGPT":
		if os.Getenv("OPENAI_API_KEY") != "" {
			return nil
		}
		return fmt.Errorf("OPENAI_API_KEY is not set.\nPlease set it via:\n  export OPENAI_API_KEY=\"your_key\"\nor add OPENAI_API_KEY=your_key in .env")

	case "Gemini":
		if os.Getenv("GEMINI_API_KEY") != "" {
			return nil
		}
		home, err := os.UserHomeDir()
		if err == nil {
			settingsPath := filepath.Join(home, ".gemini", "settings.json")
			if data, err := os.ReadFile(settingsPath); err == nil {
				var s map[string]any
				if err := json.Unmarshal(data, &s); err == nil {
					if key, ok := s["apiKey"].(string); ok && key != "" {
						os.Setenv("GEMINI_API_KEY", key)
						return nil
					}
				}
			}
		}
		return fmt.Errorf("GEMINI_API_KEY is not set.\nPlease set it via:\n  export GEMINI_API_KEY=\"your_key\"\nor add GEMINI_API_KEY=your_key in .env")
	}
	return nil
}

func streamToAI(provider, model, payload string) error {
	client := &http.Client{Timeout: 180 * time.Second}

	switch provider {
	case "Ollama":
		return streamOllama(client, model, payload)
	case "Claude":
		return streamClaude(client, model, payload)
	case "ChatGPT":
		return streamChatGPT(client, model, payload)
	case "Gemini":
		return streamGemini(client, model, payload)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

func streamOllama(client *http.Client, model, payload string) error {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": contextPrompt},
			{"role": "user", "content": payload},
		},
		"stream": true,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := client.Post("http://localhost:11434/api/chat", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var chunk struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err == nil {
			fmt.Print(chunk.Message.Content)
			if chunk.Done {
				break
			}
		}
	}
	return scanner.Err()
}

func streamChatGPT(client *http.Client, model, payload string) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": contextPrompt},
			{"role": "user", "content": payload},
		},
		"stream": true,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		}
	}
	return scanner.Err()
}

func streamClaude(client *http.Client, model, payload string) error {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	reqBody := map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"system":     contextPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": payload},
		},
		"stream": true,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var chunk struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			if chunk.Type == "content_block_delta" && chunk.Delta.Text != "" {
				fmt.Print(chunk.Delta.Text)
			}
		}
	}
	return scanner.Err()
}

func streamGemini(client *http.Client, model, payload string) error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", model, apiKey)

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": contextPrompt + "\n\n" + payload},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var chunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			for _, cand := range chunk.Candidates {
				for _, part := range cand.Content.Parts {
					fmt.Print(part.Text)
				}
			}
		}
	}
	return scanner.Err()
}

func loadEnvFile() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return // .env doesn't exist, ignore
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			if len(val) >= 2 {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
