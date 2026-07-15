package cmds

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GraphBuild() {
	var existingNotes string

	if _, err := os.Stat(KnowledgeGraphPath); err == nil {
		fmt.Print("knowledge_graph.json already exists. Rebuild it? [y/n]: ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Cancelled. Run 'bitconfig graph show' to view the current graph.")
			return
		}

		if data, err := os.ReadFile(KnowledgeGraphPath); err == nil {
			os.WriteFile("./knowledge_graph.json.bak", data, 0644)
			fmt.Println("Previous knowledge graph backed up to knowledge_graph.json.bak")
		}

		existingNotes = extractUserNotesFromGraph()
	}

	readmeSummary := summarizeReadmeForGraph()

	fmt.Print("Add developer notes to the graph? [0 = skip, 1 = add]: ")
	var choice int
	fmt.Scanln(&choice)

	userNotes := existingNotes
	if choice == 1 {
		userNotes = promptDeveloperNotes()
	} else if choice != 0 {
		fmt.Println("Not a valid option. Skipping notes.")
	}

	if err := buildKnowledgeGraph(readmeSummary, userNotes); err != nil {
		fmt.Printf("Failed to build knowledge graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done! Run 'bitconfig graph show' to inspect it or 'bitconfig push-context' to send it to your agent.")
}

func GraphShow() {
	graph, err := LoadKnowledgeGraph()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Knowledge Graph")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Project:     %s\n", graph.ProjectName)
	fmt.Printf("Built:       %s\n", graph.BuiltAt)
	fmt.Printf("Nodes:       %d\n", len(graph.Nodes))
	fmt.Printf("Connections: %d\n\n", len(graph.Edges))

	fmt.Println("NODES")
	fmt.Println(strings.Repeat("-", 50))
	for _, node := range graph.Nodes {
		fmt.Printf("  [%s] %s", node.Type, node.Label)
		if node.Path != "" {
			fmt.Printf(" (%s)", node.Path)
		}
		fmt.Println()

		if node.Properties != nil {
			if lang, ok := node.Properties["language"].(string); ok {
				fmt.Printf("    language: %s\n", lang)
			}
			if summary, ok := node.Properties["summary"].(string); ok {
				fmt.Printf("    summary: %s\n", truncate(summary, 120))
			}
			if notes, ok := node.Properties["notes"].(string); ok {
				fmt.Printf("    notes: %s\n", truncate(notes, 120))
			}
		}
	}

	fmt.Println("\nCONNECTIONS")
	fmt.Println(strings.Repeat("-", 50))
	for _, edge := range graph.Edges {
		fmt.Printf("  %s --[%s]--> %s\n",
			graph.nodeLabel(edge.Source),
			edge.Relation,
			graph.nodeLabel(edge.Target),
		)
	}
}

func GraphNote() {
	notes := promptDeveloperNotes()
	if notes == "" {
		fmt.Println("No notes entered.")
		return
	}

	readmeSummary := extractReadmeSummaryFromGraph()
	if readmeSummary == "" {
		readmeSummary = summarizeReadmeForGraph()
	}

	mergedNotes := mergeNotes(extractUserNotesFromGraph(), notes)
	if err := buildKnowledgeGraph(readmeSummary, mergedNotes); err != nil {
		fmt.Printf("Failed to update knowledge graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Developer notes added to the knowledge graph.")
}

func summarizeReadmeForGraph() string {
	if _, err := os.Stat("./README.md"); err != nil {
		fmt.Println("No README.md found. Skipping README summarization.")
		return extractReadmeSummaryFromGraph()
	}

	cmd := exec.Command("deno", "run", "--allow-read", "--allow-write", "--allow-env", "Summarizer/summarize.ts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to summarize README.md: %v\n", err)
		return extractReadmeSummaryFromGraph()
	}

	data, err := os.ReadFile("./context.txt")
	if err != nil {
		return extractReadmeSummaryFromGraph()
	}

	fmt.Println("README summary captured for the knowledge graph.")
	return strings.TrimSpace(string(data))
}

func buildKnowledgeGraph(readmeSummary, userNotes string) error {
	projectName := filepath.Base(mustGetwd())
	if config, err := LoadBitConfig(); err == nil {
		projectName = config.ProjectName
	}

	builder := NewGraphBuilder(projectName)
	if err := builder.BuildFromFilesystem("."); err != nil {
		return err
	}

	if config, err := LoadBitConfig(); err == nil {
		builder.AddBitConfig(config)
	}

	builder.AddReadmeSummary(readmeSummary)
	builder.AddUserContext(userNotes)

	graph := builder.Graph()
	if err := SaveKnowledgeGraph(graph); err != nil {
		return err
	}

	fmt.Printf("Knowledge graph saved to %s (%d nodes, %d connections)\n",
		KnowledgeGraphPath, len(graph.Nodes), len(graph.Edges))
	return nil
}

func promptDeveloperNotes() string {
	fmt.Println("Enter developer notes (type DONE on a new line when finished):")

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "DONE" || line == "Done" || line == "done" {
			break
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func extractReadmeSummaryFromGraph() string {
	graph, err := LoadKnowledgeGraph()
	if err != nil {
		return ""
	}
	for _, node := range graph.Nodes {
		if node.Type == "readme_summary" {
			if summary, ok := node.Properties["summary"].(string); ok {
				return summary
			}
		}
	}
	return ""
}

func extractUserNotesFromGraph() string {
	graph, err := LoadKnowledgeGraph()
	if err != nil {
		return ""
	}
	for _, node := range graph.Nodes {
		if node.Type == "user_context" {
			if notes, ok := node.Properties["notes"].(string); ok {
				return notes
			}
		}
	}
	return ""
}

func mergeNotes(existing, additional string) string {
	existing = strings.TrimSpace(existing)
	additional = strings.TrimSpace(additional)
	if existing == "" {
		return additional
	}
	if additional == "" {
		return existing
	}
	return existing + "\n" + additional
}

func GraphHelp() {
	fmt.Println("Usage: bitconfig graph <subcommand>")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  build   Scan the project and build knowledge_graph.json")
	fmt.Println("  show    Print nodes and connections in the terminal")
	fmt.Println("  note    Add developer notes to the knowledge graph")
}
