# BitConfig

BitConfig is a developer CLI tool that captures project architecture, tracks dependency changes, and constructs a structured Knowledge Graph of software repositories. It features an integrated Machine Learning pipeline utilizing a **Relational Graph Convolutional Network (R-GCN)** to perform link prediction and blast-radius risk scoring, streaming rich context directly to configured AI providers via live HTTP (Ollama, Claude, ChatGPT, Gemini).

---

## Supported AI Providers

BitConfig connects directly to AI providers via native HTTP streaming — no external CLI wrappers or heavy dependencies required:

- **Ollama**: Local models (`llama3.2`, `gemma4:e4b`, `mistral`, etc.) at `http://localhost:11434` (free & private).
- **Claude**: Anthropic Messages API (`claude-3-5-sonnet-latest`, `claude-3-7-sonnet-latest`).
- **ChatGPT**: OpenAI API (`gpt-4o`, `o3-mini`, etc.).
- **Gemini**: Google Generative AI REST API (`gemini-1.5-flash`, `gemini-2.0-flash`).

---

## Installation

### Option 1: Install Prebuilt Binary Globally (Recommended)

Copy the prebuilt binary for your OS to your system PATH:

```bash
# macOS (Apple Silicon M1/M2/M3/M4)
sudo cp dist/bitconfig-darwin-arm64 /usr/local/bin/bitconfig

# macOS (Intel)
sudo cp dist/bitconfig-darwin-amd64 /usr/local/bin/bitconfig

# Linux (x86_64)
sudo cp dist/bitconfig-linux-amd64 /usr/local/bin/bitconfig

# Linux (ARM64)
sudo cp dist/bitconfig-linux-arm64 /usr/local/bin/bitconfig
```

Make sure it's executable:
```bash
sudo chmod +x /usr/local/bin/bitconfig
```

### Option 2: Build From Source

Requirements: [Go 1.22+](https://go.dev/dl/)

```bash
# Clone the repository
git clone https://github.com/RohanGoparaju028/BitConfig.git
cd BitConfig

# Build host binary
go build -o bitconfig main.go

# Or cross-compile for all platforms:
./build_release.sh
```

---

## Quick Start & Usage

Navigate to **any software project or repository** and run the following:

### 1. Initialize Project Configuration
```bash
bitconfig init
```
Scans for languages and dependency manifests (`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, etc.) and prompts you to select your AI provider and model.

### 2. Configure Authentication (if using Cloud APIs)
For cloud providers, set the API key in your environment or add it to a `.env` file in the project root:

```bash
# If using Claude
export ANTHROPIC_API_KEY="your_api_key"

# If using ChatGPT
export OPENAI_API_KEY="your_api_key"

# If using Gemini
export GEMINI_API_KEY="your_api_key"

# If using Ollama (no key needed; just make sure Ollama is running)
ollama serve
```

### 3. Build the Knowledge Graph
```bash
bitconfig graph build
```
Scans repository files, directory relationships, and summarizes README.md using native Go text processing to create `knowledge_graph.json`.

To inspect the graph in your terminal:
```bash
bitconfig graph show
```

### 4. (Optional) Train the Graph Neural Network (GNN)
Learn structural node embeddings and evaluate blast-radius risk:

```bash
# Train R-GCN link prediction model
bitconfig graph gnn

# View blast-radius risk scores and hidden couplings
bitconfig graph gnn show
```
*(Requires Python 3 with PyTorch. Automatically uses Apple Silicon MPS, CUDA, or CPU fallback.)*

### 5. Push Context to Your AI
```bash
bitconfig push-context
```
Streams the structured knowledge graph and GNN ML architectural predictions directly to your configured AI model in real time.

### 6. Track Dependency Changes
```bash
# Compare current dependencies against the saved snapshot
bitconfig diff

# Update .bitconfig with new dependencies
bitconfig update

# Check project status
bitconfig status
```

---

## Machine Learning Architecture (`gnn/`)

- **Model**: Multi-layer Relational Graph Convolutional Network (R-GCN) with basis decomposition (`gnn/model.py`).
- **Decoder**: DistMult bilinear scoring for link prediction ($P(\text{edge}(u, r, v)) = \sigma(h_u^\top R_r h_v)$).
- **Criticality Head**: Multi-Layer Perceptron (MLP) mapping node graph representations to blast-radius risk scores.
- **Hardware Acceleration**: Automatically leverages Apple Silicon (`mps`), NVIDIA CUDA, or CPU.

---

## Commands Summary

| Command | Description |
| :--- | :--- |
| `bitconfig init` | Initialize `.bitconfig` and select your AI provider |
| `bitconfig graph build` | Construct `knowledge_graph.json` from workspace files and README |
| `bitconfig graph show` | View nodes and architectural edges in terminal |
| `bitconfig graph note` | Append developer notes to the knowledge graph |
| `bitconfig graph gnn` | Train PyTorch R-GCN model on the graph |
| `bitconfig graph gnn show`| Inspect blast-radius criticality scores and predicted couplings |
| `bitconfig push-context` | Stream knowledge graph payload directly to configured AI |
| `bitconfig diff` | Compare workspace against `.bitconfig` snapshot |
| `bitconfig update` | Record dependency additions/removals into `.bitconfig` |
| `bitconfig status` | Display configuration, tracked files, and graph stats |
| `bitconfig help` | Show help menu and usage |
