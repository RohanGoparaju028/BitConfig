import math
from typing import Optional, Tuple
import torch
import torch.nn as nn
import torch.nn.functional as F

class RGCNLayer(nn.Module):
    """
    Relational Graph Convolutional Layer (Schlichtkrull et al.).
    Performs relation-specific message passing and neighbor aggregation:
        h_i^{(l+1)} = ReLU( W_0 h_i^{(l)} + sum_{r in R} sum_{j in N_i^r} (1 / c_{i,r}) W_r h_j^{(l)} )
    """

    def __init__(
        self,
        in_channels: int,
        out_channels: int,
        num_relations: int,
        num_bases: Optional[int] = None,
        dropout: float = 0.1,
    ):
        super().__init__()
        self.in_channels = in_channels
        self.out_channels = out_channels
        self.num_relations = num_relations
        self.num_bases = num_bases
        self.dropout = nn.Dropout(dropout)

        # Self-loop weight
        self.w_self = nn.Linear(in_channels, out_channels, bias=True)

        if num_bases is not None and num_bases < num_relations:
            # Basis decomposition for parameter sharing across relations
            self.bases = nn.Parameter(torch.Tensor(num_bases, in_channels, out_channels))
            self.rel_coeffs = nn.Parameter(torch.Tensor(num_relations, num_bases))
            nn.init.xavier_uniform_(self.bases)
            nn.init.xavier_uniform_(self.rel_coeffs)
        else:
            # Independent weight matrix per relation
            self.rel_weights = nn.Parameter(torch.Tensor(num_relations, in_channels, out_channels))
            nn.init.xavier_uniform_(self.rel_weights)

        self.norm = nn.LayerNorm(out_channels)

    def _get_rel_weight(self, r: int) -> torch.Tensor:
        if self.num_bases is not None and self.num_bases < self.num_relations:
            return torch.einsum("b,bio->io", self.rel_coeffs[r], self.bases)
        return self.rel_weights[r]

    def forward(
        self,
        x: torch.Tensor,
        edge_index: torch.Tensor,
        edge_type: torch.Tensor,
    ) -> torch.Tensor:
        num_nodes = x.size(0)
        out = self.w_self(x)  # [num_nodes, out_channels]

        src, dst = edge_index[0], edge_index[1]

        # Aggregate messages per relation type
        unique_types = torch.unique(edge_type)
        for r_idx in unique_types:
            mask = (edge_type == r_idx)
            if not mask.any():
                continue

            r_src = src[mask]
            r_dst = dst[mask]

            W_r = self._get_rel_weight(r_idx.item())  # [in_channels, out_channels]
            msg = torch.matmul(x[r_src], W_r)         # [num_edges_r, out_channels]

            # In-degree normalization (c_{i,r})
            deg = torch.zeros(num_nodes, dtype=torch.float32, device=x.device)
            deg.scatter_add_(0, r_dst, torch.ones_like(r_dst, dtype=torch.float32))
            deg = torch.clamp(deg, min=1.0)
            norm_factor = 1.0 / deg[r_dst].unsqueeze(-1)
            norm_msg = msg * norm_factor

            # Accumulate into output
            out.scatter_add_(0, r_dst.unsqueeze(-1).expand(-1, self.out_channels), norm_msg)

        out = self.norm(out)
        out = F.relu(out)
        out = self.dropout(out)
        return out


class RelationalGNN(nn.Module):
    """
    End-to-end Graph Neural Network architecture for BitConfig.
    Computes:
      1. Node architectural representations via multi-layer RGCN.
      2. Link prediction probabilities via DistMult bilinear relational decoder.
      3. Node architectural criticality / blast-radius scores via MLP head.
    """

    def __init__(
        self,
        in_dim: int,
        hidden_dim: int = 64,
        out_dim: int = 32,
        num_relations: int = 15,
        num_bases: Optional[int] = 8,
        dropout: float = 0.1,
    ):
        super().__init__()
        self.conv1 = RGCNLayer(in_dim, hidden_dim, num_relations, num_bases, dropout)
        self.conv2 = RGCNLayer(hidden_dim, out_dim, num_relations, num_bases, dropout)

        # DistMult relation embeddings for link prediction
        self.rel_embeddings = nn.Embedding(num_relations, out_dim)
        nn.init.xavier_uniform_(self.rel_embeddings.weight)

        # Node importance / blast radius estimation head
        self.criticality_head = nn.Sequential(
            nn.Linear(out_dim, 16),
            nn.ReLU(),
            nn.Linear(16, 1),
            nn.Sigmoid(),
        )

    def encode(
        self,
        x: torch.Tensor,
        edge_index: torch.Tensor,
        edge_type: torch.Tensor,
    ) -> torch.Tensor:
        """Runs message passing and returns node embeddings."""
        h = self.conv1(x, edge_index, edge_type)
        h = self.conv2(h, edge_index, edge_type)
        return h

    def decode_links(
        self,
        h: torch.Tensor,
        src: torch.Tensor,
        dst: torch.Tensor,
        rel: torch.Tensor,
    ) -> torch.Tensor:
        """
        DistMult scoring function: score(u, r, v) = sum(h_u * R_r * h_v).
        Returns link existence logits / probabilities.
        """
        h_u = h[src]                         # [B, out_dim]
        h_v = h[dst]                         # [B, out_dim]
        r_emb = self.rel_embeddings(rel)      # [B, out_dim]

        scores = torch.sum(h_u * r_emb * h_v, dim=-1)
        return scores

    def predict_criticality(self, h: torch.Tensor) -> torch.Tensor:
        """Estimates architectural criticality score in [0, 1] for each node."""
        return self.criticality_head(h).squeeze(-1)

    def forward(
        self,
        x: torch.Tensor,
        edge_index: torch.Tensor,
        edge_type: torch.Tensor,
        link_src: torch.Tensor,
        link_dst: torch.Tensor,
        link_rel: torch.Tensor,
    ) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor]:
        """
        Returns:
          - link_scores: Logits for the evaluated (src, rel, dst) edges
          - criticality: Node importance scores [0, 1]
          - embeddings: Final continuous node representations
        """
        h = self.encode(x, edge_index, edge_type)
        scores = self.decode_links(h, link_src, link_dst, link_rel)
        crit = self.predict_criticality(h)
        return scores, crit, h
