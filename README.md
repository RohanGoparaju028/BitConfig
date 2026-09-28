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

### Option 1: Install a prebuilt release (recommended)

Open the project's [GitHub Releases](https://github.com/RohanGoparaju028/BitConfig/releases) and download the archive matching your operating system and CPU. Release archives are named `bitconfig-vX.Y.Z-OS-ARCH.tar.gz` and include the CLI, README, license, and GNN Python files.

On macOS or Linux, extract the archive, then install the binary. Replace the archive name below with the one you downloaded:

```bash
tar -xzf bitconfig-v0.1.0-darwin-arm64.tar.gz
cd bitconfig-v0.1.0-darwin-arm64
sudo install -m 755 bitconfig /usr/local/bin/bitconfig
```

Use `darwin-arm64` for Apple Silicon, `darwin-amd64` for Intel Macs, `linux-amd64` for x86_64 Linux, or `linux-arm64` for ARM64 Linux. For GNN support, keep the included `gnn/` directory beside the installed executable:

```bash
sudo mkdir -p /usr/local/bin/gnn
sudo install -m 644 gnn/*.py gnn/requirements.txt /usr/local/bin/gnn/
```

For Windows, extract the `windows-amd64` archive and add its directory to your `PATH`. Keep the `gnn` directory beside `bitconfig.exe` if you want GNN support.

To use the optional GNN feature, install its Python requirements:

```bash
python3 -m pip install -r /usr/local/bin/gnn/requirements.txt
```

Verify the CLI is available with `bitconfig help`.

### Option 2: Build From Source

Requirements: Go (see the `go` version in `go.mod`).

```bash
# Clone the repository
git clone https://github.com/RohanGoparaju028/BitConfig.git
cd BitConfig

# Build host binary
go build -o bitconfig main.go

# Or build versioned release archives for all supported platforms:
./build_release.sh v0.1.0
```

### Checksums

Release archives include a `SHA256SUMS.txt` file. On macOS or Linux, verify a downloaded archive from the directory containing the checksum file:

```bash
shasum -a 256 -c SHA256SUMS.txt
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

### 6. Keep an Interactive Agent Conversation
```bash
bitconfig chat
```
Ask follow-up questions about the project in one ongoing session. BitConfig includes the project graph as context and keeps earlier user and assistant messages in the conversation history. Type `exit` or `quit`, or send EOF (Ctrl-D), to end the session.

### 7. Track Dependency Changes
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
