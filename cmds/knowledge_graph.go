package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const KnowledgeGraphPath = "./knowledge_graph.json"

var skipList = map[string]bool{
	".git": true, ".gitignore": true, ".DS_Store": true,
	"node_modules": true, "deno.lock": true, ".next": true, "dist": true, ".turbo": true,
	"vendor": true,
	".gradle": true, "gradle": true, "build": true, "target": true, ".idea": true, "out": true,
	"bin": true, "obj": true, ".vs": true, "packages": true,
	"__pycache__": true, ".venv": true, "venv": true, ".mypy_cache": true, ".pytest_cache": true,
	".bundle": true,
	".build": true, "DerivedData": true, "Pods": true,
	".terraform": true, "__MACOSX": true, "coverage": true, ".nyc_output": true,
	"logs": true, "tmp": true, "temp": true,
}

type KGNode struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Label      string         `json:"label"`
	Path       string         `json:"path,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type KGEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type KnowledgeGraph struct {
	ProjectName string   `json:"project_name"`
	BuiltAt     string   `json:"built_at"`
	Nodes       []KGNode `json:"nodes"`
	Edges       []KGEdge `json:"edges"`
}

var extensionLanguages = map[string]string{
	".go":   "go",
	".mod":  "go",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".jsx":  "javascript",
	".py":   "python",
	".rb":   "ruby",
	".rs":   "rust",
	".java": "java",
	".cs":   "csharp",
	".md":   "markdown",
	".json": "json",
}

type GraphBuilder struct {
	graph     KnowledgeGraph
	nodeIndex map[string]bool
}

func NewGraphBuilder(projectName string) *GraphBuilder {
	return &GraphBuilder{
		graph: KnowledgeGraph{
			ProjectName: projectName,
			BuiltAt:     time.Now().Format("2006-01-02 15:04:05"),
			Nodes:       []KGNode{},
			Edges:       []KGEdge{},
		},
		nodeIndex: map[string]bool{},
	}
}

func (b *GraphBuilder) addNode(node KGNode) {
	if b.nodeIndex[node.ID] {
		return
	}
	b.nodeIndex[node.ID] = true
	b.graph.Nodes = append(b.graph.Nodes, node)
}

func (b *GraphBuilder) addEdge(source, target, relation string) {
	if source == "" || target == "" || source == target {
		return
	}
	b.graph.Edges = append(b.graph.Edges, KGEdge{
		Source:   source,
		Target:   target,
		Relation: relation,
	})
}

func (b *GraphBuilder) BuildFromFilesystem(root string) error {
	projectID := "project:root"
	b.addNode(KGNode{
		ID:    projectID,
		Type:  "project",
		Label: b.graph.ProjectName,
		Path:  root,
	})

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}

		parts := strings.Split(rel, string(os.PathSeparator))
		for _, part := range parts {
			if skipList[part] {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			dirID := nodeID("directory", rel)
			b.addNode(KGNode{
				ID:    dirID,
				Type:  "directory",
				Label: d.Name(),
				Path:  rel,
			})

			parentID := parentNodeID(rel, projectID)
			b.addEdge(parentID, dirID, "contains")
			b.addEdge(dirID, projectID, "part_of")
			return nil
		}

		fileID := nodeID("file", rel)
		ext := strings.ToLower(filepath.Ext(d.Name()))
		lang := extensionLanguages[ext]

		props := map[string]any{"extension": ext}
		if lang != "" {
			props["language"] = lang
		}

		b.addNode(KGNode{
			ID:         fileID,
			Type:       "file",
			Label:      d.Name(),
			Path:       rel,
			Properties: props,
		})

		parentID := parentNodeID(rel, projectID)
		b.addEdge(parentID, fileID, "contains")
		b.addEdge(fileID, projectID, "part_of")

		if lang != "" {
			langID := nodeID("language", lang)
			b.addNode(KGNode{
				ID:    langID,
				Type:  "language",
				Label: lang,
			})
			b.addEdge(fileID, langID, "written_in")
			b.addEdge(langID, projectID, "used_by")
		}

		if isDependencyFile(d.Name()) {
			depID := nodeID("dependency_file", rel)
			b.addNode(KGNode{
				ID:    depID,
				Type:  "dependency_file",
				Label: d.Name(),
				Path:  rel,
			})
			b.addEdge(depID, projectID, "declares_dependencies_for")
			if lang != "" {
				b.addEdge(depID, nodeID("language", lang), "manages")
			}
		}

		return nil
	})
}

func (b *GraphBuilder) AddReadmeSummary(summary string) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return
	}

	nodeID := "knowledge:readme_summary"
	b.addNode(KGNode{
		ID:    nodeID,
		Type:  "readme_summary",
		Label: "README Summary",
		Properties: map[string]any{
			"summary": summary,
		},
	})
	b.addEdge(nodeID, "project:root", "documents")
}

func (b *GraphBuilder) AddUserContext(notes string) {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return
	}

	nodeID := "knowledge:user_context"
	b.addNode(KGNode{
		ID:    nodeID,
		Type:  "user_context",
		Label: "Developer Notes",
		Properties: map[string]any{
			"notes": notes,
		},
	})
	b.addEdge(nodeID, "project:root", "describes")
}

func (b *GraphBuilder) AddBitConfig(config BitConfigFile) {
	for lang := range uniqueStrings(config.Languages) {
		langID := nodeID("language", strings.ToLower(lang))
		b.addNode(KGNode{
			ID:    langID,
			Type:  "language",
			Label: lang,
			Properties: map[string]any{
				"source": "bitconfig",
			},
		})
		b.addEdge(langID, "project:root", "used_by")
	}

	for depPath, lines := range config.Dependencies {
		rel := depPath
		if filepath.IsAbs(depPath) {
			if wd, err := os.Getwd(); err == nil {
				if r, err := filepath.Rel(wd, depPath); err == nil {
					rel = r
				}
			}
		}

		depID := nodeID("dependency_file", rel)
		b.addNode(KGNode{
			ID:    depID,
			Type:  "dependency_file",
			Label: filepath.Base(rel),
			Path:  rel,
			Properties: map[string]any{
				"lines":  lines,
				"source": "bitconfig",
			},
		})
		b.addEdge(depID, "project:root", "declares_dependencies_for")
	}
}

func (b *GraphBuilder) Graph() KnowledgeGraph {
	return b.graph
}

func SaveKnowledgeGraph(graph KnowledgeGraph) error {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(KnowledgeGraphPath, data, 0644)
}

func LoadKnowledgeGraph() (KnowledgeGraph, error) {
	data, err := os.ReadFile(KnowledgeGraphPath)
	if err != nil {
		return KnowledgeGraph{}, fmt.Errorf("knowledge graph not found — run 'bitconfig get-context' first")
	}

	var graph KnowledgeGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return KnowledgeGraph{}, fmt.Errorf("could not read knowledge graph: %v", err)
	}
	return graph, nil
}

func (kg KnowledgeGraph) ToAgentPayload(config BitConfigFile) string {
	var b strings.Builder

	b.WriteString("=== BitConfig Knowledge Graph ===\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", config.ProjectName))
	b.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(config.Languages, ", ")))
	b.WriteString(fmt.Sprintf("Terminal Agent: %s\n", config.Model))
	b.WriteString(fmt.Sprintf("Graph built: %s\n", kg.BuiltAt))
	b.WriteString(fmt.Sprintf("Nodes: %d | Connections: %d\n\n", len(kg.Nodes), len(kg.Edges)))

	b.WriteString("--- NODES ---\n")
	for _, node := range kg.Nodes {
		b.WriteString(fmt.Sprintf("[%s] %s (%s)", node.Type, node.Label, node.ID))
		if node.Path != "" {
			b.WriteString(fmt.Sprintf(" path=%s", node.Path))
		}
		b.WriteString("\n")

		if node.Properties != nil {
			if summary, ok := node.Properties["summary"].(string); ok && summary != "" {
				b.WriteString("  summary: ")
				b.WriteString(truncate(summary, 500))
				b.WriteString("\n")
			}
			if notes, ok := node.Properties["notes"].(string); ok && notes != "" {
				b.WriteString("  notes: ")
				b.WriteString(truncate(notes, 500))
				b.WriteString("\n")
			}
			if lang, ok := node.Properties["language"].(string); ok && lang != "" {
				b.WriteString(fmt.Sprintf("  language: %s\n", lang))
			}
		}
	}

	b.WriteString("\n--- CONNECTIONS ---\n")
	for _, edge := range kg.Edges {
		source := kg.nodeLabel(edge.Source)
		target := kg.nodeLabel(edge.Target)
		b.WriteString(fmt.Sprintf("%s --[%s]--> %s\n", source, edge.Relation, target))
	}

	b.WriteString("\n--- GRAPH JSON ---\n")
	if data, err := json.MarshalIndent(kg, "", "  "); err == nil {
		b.Write(data)
	}

	b.WriteString("\n\nUse the nodes and connections above to understand how files, languages, docs, and dependencies relate in this project.\n")
	return b.String()
}

func (kg KnowledgeGraph) nodeLabel(id string) string {
	for _, node := range kg.Nodes {
		if node.ID == id {
			if node.Path != "" {
				return node.Path
			}
			return node.Label
		}
	}
	return id
}

func nodeID(nodeType, key string) string {
	return fmt.Sprintf("%s:%s", nodeType, key)
}

func parentNodeID(rel, projectID string) string {
	parent := filepath.Dir(rel)
	if parent == "." || parent == "" {
		return projectID
	}
	return nodeID("directory", parent)
}

func isDependencyFile(name string) bool {
	dependencyNames := map[string]bool{
		"go.mod":         true,
		"go.sum":         true,
		"package.json":   true,
		"package-lock.json": true,
		"yarn.lock":      true,
		"requirements.txt": true,
		"Pipfile":        true,
		"pyproject.toml": true,
		"pom.xml":        true,
		"build.gradle":   true,
		"Gemfile":        true,
		"Gemfile.lock":   true,
		"Cargo.toml":     true,
		"Cargo.lock":     true,
	}
	return dependencyNames[name]
}

func uniqueStrings(items []string) map[string]bool {
	set := map[string]bool{}
	for _, item := range items {
		set[strings.ToLower(item)] = true
	}
	return set
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
