BitConfig is a developer CLI tool that captures project architecture, tracks dependency changes, and constructs a structured Knowledge Graph of software repositories. It features an integrated Machine Learning pipeline utilizing a **Relational Graph Convolutional Network (R-GCN)** to perform link prediction and blast-radius risk scoring, pushing rich context directly to terminal AI agents (Claude Code, Cursor Agent, Ollama, Gemini CLI, Aider).

### Core Features

1. **`init`**: Initialize `.bitconfig` by detecting programming languages, dependency manifests, and configuring your target terminal AI agent.
2. **`graph build`** (or `get-context`): Scans the repository filesystem and README summary to generate a heterogeneous `knowledge_graph.json`.
3. **`graph gnn`**: **[ML Feature]** Trains a PyTorch Relational Graph Convolutional Network (R-GCN) with negative sampling on Apple Silicon / CUDA to learn node representations, predict hidden architectural couplings, and evaluate change blast-radius risk.
4. **`graph gnn show`**: Inspect GNN training metrics (ROC-AUC, loss), top critical files/modules, and predicted implicit couplings.
5. **`graph show`**: Print nodes and graph connections directly in your terminal.
6. **`push-context`**: Formats the Knowledge Graph and GNN ML architectural predictions into an optimized payload and launches your terminal AI agent with standard I/O streaming.
7. **`diff`**: Compares current workspace dependencies against `.bitconfig` snapshots to report additions or removals.
8. **`update`**: Updates `.bitconfig` with new dependencies and records dependency change history.
9. **`status`**: Shows current configuration, tracked dependency snapshots, and graph status.

### Machine Learning Architecture (`gnn/`)

- **Model**: Multi-layer Relational Graph Convolutional Network (R-GCN) with basis decomposition (`gnn/model.py`).
- **Decoder**: DistMult bilinear scoring for link prediction ($P(\text{edge}(u, r, v)) = \sigma(h_u^\top R_r h_v)$).
- **Criticality Head**: Multi-Layer Perceptron (MLP) mapping node graph representations to blast-radius risk scores.
- **Hardware Acceleration**: Automatically leverages Apple Silicon (`mps`), NVIDIA CUDA, or CPU fallback.

### Building & Running

```bash
# Build the binary
go build -o bitconfig main.go

# Initialize and scan project
./bitconfig init

# Build Knowledge Graph
./bitconfig graph build

# Train GNN on Knowledge Graph
./bitconfig graph gnn

# View GNN predictions
./bitconfig graph gnn show

# Push graph & ML insights to your AI agent
./bitconfig push-context
```
