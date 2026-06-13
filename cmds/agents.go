package cmds

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const contextPrompt = "You have been given project context below. Review it and help me understand this codebase. Suggest what to focus on next."

type terminalAgent struct {
	Name    string
	Binaries []string // try in order
}

var terminalAgents = map[string]terminalAgent{
	"Claude Code":  {Name: "Claude Code", Binaries: []string{"claude"}},
	"Cursor Agent": {Name: "Cursor Agent", Binaries: []string{"agent", "cursor-agent"}},
	"Ollama":       {Name: "Ollama", Binaries: []string{"ollama"}},
	"Gemini CLI":   {Name: "Gemini CLI", Binaries: []string{"gemini"}},
	"Aider":        {Name: "Aider", Binaries: []string{"aider"}},
}

var legacyAgentNames = map[string]string{
	"Claude":  "Claude Code",
	"Gemini":  "Gemini CLI",
	"ChatGpt": "",
	"Ollama":  "Ollama",
}

func resolveAgent(name string) (terminalAgent, error) {
	if agent, ok := terminalAgents[name]; ok {
		return agent, nil
	}
	if mapped, ok := legacyAgentNames[name]; ok {
		if mapped == "" {
			return terminalAgent{}, fmt.Errorf("'%s' is a web LLM — re-run 'bitconfig init' and pick a terminal agent", name)
		}
		return terminalAgents[mapped], nil
	}
	return terminalAgent{}, fmt.Errorf("unknown agent '%s' — re-run 'bitconfig init' to pick a terminal agent", name)
}

func findBinary(candidates []string) (string, error) {
	for _, bin := range candidates {
		if path, err := exec.LookPath(bin); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("none of [%s] found in PATH — install the CLI first", strings.Join(candidates, ", "))
}

func PushContext() {
	loadEnvFile()

	config, err := LoadBitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	contextData, err := os.ReadFile("./context.txt")
	if err != nil {
		fmt.Println("No context.txt found. Run 'bitconfig get-context' first.")
		os.Exit(1)
	}

	if len(strings.TrimSpace(string(contextData))) == 0 {
		fmt.Println("context.txt is empty. Run 'bitconfig get-context' to generate project context.")
		os.Exit(1)
	}

	payload := buildContextPayload(config, contextData)

	agent, err := resolveAgent(config.Model)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	binary, err := findBinary(agent.Binaries)
	if err != nil {
		fmt.Println(err)
		printInstallHint(agent.Name)
		os.Exit(1)
	}

	if err := validateAgentAuth(agent.Name); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sending context to %s...\n\n", agent.Name)
	if err := runTerminalAgent(binary, agent.Name, config.AgentModel, payload); err != nil {
		fmt.Printf("Failed to run %s: %v\n", agent.Name, err)
		os.Exit(1)
	}
}

func buildContextPayload(config BitConfigFile, contextData []byte) string {
	var builder strings.Builder
	builder.WriteString("=== BitConfig Project Context ===\n")
	builder.WriteString(fmt.Sprintf("Project: %s\n", config.ProjectName))
	builder.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(config.Languages, ", ")))
	builder.WriteString(fmt.Sprintf("Terminal Agent: %s\n", config.Model))
	builder.WriteString("\n--- Context ---\n")
	builder.Write(contextData)
	return builder.String()
}

func runTerminalAgent(binary, agentName, agentModel, payload string) error {
	switch agentName {
	case "Claude Code":
		return runWithStdin(binary, []string{"-p", contextPrompt, "--bare"}, payload)
	case "Cursor Agent":
		return runWithStdin(binary, []string{"-p", contextPrompt, "--trust"}, payload)
	case "Gemini CLI":
		return runWithStdin(binary, []string{"-p", contextPrompt, "--skip-trust"}, payload)
	case "Ollama":
		model := agentModel
		if model == "" {
			model = "llama3.2"
		}
		prompt := contextPrompt + "\n\n" + payload
		return runWithStdin(binary, []string{"run", model}, prompt)
	case "Aider":
		cmd := exec.Command(binary, "--read", "context.txt", "--message", contextPrompt, "--no-git")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("no runner configured for %s", agentName)
	}
}

func runWithStdin(binary string, args []string, stdin string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."
	return cmd.Run()
}

func printInstallHint(agentName string) {
	fmt.Println("\nInstall hints:")
	switch agentName {
	case "Claude Code":
		fmt.Println("  npm install -g @anthropic-ai/claude-code")
	case "Cursor Agent":
		fmt.Println("  curl https://cursor.com/install -fsS | bash")
	case "Ollama":
		fmt.Println("  https://ollama.com/download")
	case "Gemini CLI":
		fmt.Println("  npm install -g @google/gemini-cli")
	case "Aider":
		fmt.Println("  pip install aider-install && aider-install")
	}
}

func validateAgentAuth(agentName string) error {
	switch agentName {
	case "Gemini CLI":
		if os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") != "" || os.Getenv("GOOGLE_GENAI_USE_GCA") != "" {
			return nil
		}
		home, err := os.UserHomeDir()
		if err == nil {
			settingsPath := filepath.Join(home, ".gemini", "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				if data, err := os.ReadFile(settingsPath); err == nil && len(strings.TrimSpace(string(data))) > 2 {
					return nil
				}
			}
		}
		return fmt.Errorf("authentication is not configured.\nTo fix this, please do one of the following:\n  1. Set the GEMINI_API_KEY environment variable:\n     export GEMINI_API_KEY=\"your_key\"\n  2. Set GOOGLE_GENAI_USE_VERTEXAI=true or GOOGLE_GENAI_USE_GCA=true\n  3. Configure your API key/auth in ~/.gemini/settings.json")

	case "Claude Code":
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			return nil
		}
		home, err := os.UserHomeDir()
		if err == nil {
			configDirs := []string{
				filepath.Join(home, ".config", "claude-code"),
				filepath.Join(home, ".config", "@anthropic-ai", "claude-code"),
			}
			for _, dir := range configDirs {
				if _, err := os.Stat(dir); err == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("authentication is not configured.\nTo fix this, please do one of the following:\n  1. Set the ANTHROPIC_API_KEY environment variable:\n     export ANTHROPIC_API_KEY=\"your_key\"\n  2. Run 'claude login' to authenticate the CLI")

	case "Cursor Agent":
		if os.Getenv("CURSOR_API_KEY") != "" {
			return nil
		}
		home, err := os.UserHomeDir()
		if err == nil {
			configDirs := []string{
				filepath.Join(home, ".config", "cursor-agent"),
				filepath.Join(home, ".cursor-agent"),
				filepath.Join(home, ".cursor"),
			}
			for _, dir := range configDirs {
				if _, err := os.Stat(dir); err == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("authentication is not configured.\nTo fix this, please do one of the following:\n  1. Set the CURSOR_API_KEY environment variable:\n     export CURSOR_API_KEY=\"your_key\"\n  2. Run 'agent login' to authenticate the CLI")

	case "Aider":
		keys := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "GROQ_API_KEY", "COHERE_API_KEY"}
		for _, key := range keys {
			if os.Getenv(key) != "" {
				return nil
			}
		}
		return fmt.Errorf("no API key set. Aider requires at least one API key environment variable, such as:\n  export OPENAI_API_KEY=\"your_key\"\n  export ANTHROPIC_API_KEY=\"your_key\"\n  export GEMINI_API_KEY=\"your_key\"")

	case "Ollama":
		conn, err := net.DialTimeout("tcp", "127.0.0.1:11434", 500*time.Millisecond)
		if err != nil {
			conn, err = net.DialTimeout("tcp", "localhost:11434", 500*time.Millisecond)
		}
		if err != nil {
			return fmt.Errorf("Ollama service does not appear to be running.\nPlease make sure Ollama is started before running this command")
		}
		conn.Close()
		return nil
	}
	return nil
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
