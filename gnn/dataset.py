import json
import os
import math
import hashlib
from typing import Dict, List, Tuple, Any, Optional
import numpy as np
import torch

class KnowledgeGraphDataset:
    """
    Parses knowledge_graph.json into PyTorch graph tensors with node featurization,
    relational adjacency, inverse relations, and negative edge sampling.
    """

    KNOWN_NODE_TYPES = [
        "project",
        "directory",
        "file",
        "language",
        "dependency_file",
        "readme_summary",
        "user_context",
    ]

    def __init__(self, json_path: str = "./knowledge_graph.json", feature_dim: int = 64, seed: int = 42):
        self.json_path = json_path
        self.feature_dim = feature_dim
        self.seed = seed
        np.random.seed(seed)
        torch.manual_seed(seed)

        if not os.path.exists(json_path):
            raise FileNotFoundError(f"Knowledge graph file not found at {json_path}")

        with open(json_path, "r", encoding="utf-8") as f:
            self.raw_data = json.load(f)

        self.project_name = self.raw_data.get("project_name", "Unknown")
        self.nodes = self.raw_data.get("nodes", [])
        self.edges = self.raw_data.get("edges", [])

        if not self.nodes:
            raise ValueError("Knowledge graph contains no nodes.")

        self._build_indices()
        self._build_node_features()
        self._build_edges()

    def _build_indices(self):
        self.node_to_idx: Dict[str, int] = {}
        self.idx_to_node: List[Dict[str, Any]] = []

        for i, node in enumerate(self.nodes):
            node_id = node["id"]
            self.node_to_idx[node_id] = i
            self.idx_to_node.append(node)

        self.num_nodes = len(self.nodes)

        # Collect distinct relations
        unique_relations = sorted(list({e["relation"] for e in self.edges if "relation" in e}))
        self.base_relations = unique_relations

        # Forward relations + inverse relations + self-loop relation
        self.rel_to_idx: Dict[str, int] = {}
        self.idx_to_rel: List[str] = []

        for r in unique_relations:
            self.rel_to_idx[r] = len(self.idx_to_rel)
            self.idx_to_rel.append(r)

        for r in unique_relations:
            inv_r = f"inv_{r}"
            self.rel_to_idx[inv_r] = len(self.idx_to_rel)
            self.idx_to_rel.append(inv_r)

        self.self_loop_rel = "self_loop"
        self.rel_to_idx[self.self_loop_rel] = len(self.idx_to_rel)
        self.idx_to_rel.append(self.self_loop_rel)

        self.num_relations = len(self.idx_to_rel)

    def _hash_string(self, text: str, dim: int) -> np.ndarray:
        """Projects a string to a dense continuous vector via character n-gram hashing."""
        vec = np.zeros(dim, dtype=np.float32)
        if not text:
            return vec
        # Character n-grams (sizes 2, 3, 4)
        for n in (2, 3, 4):
            for i in range(len(text) - n + 1):
                gram = text[i:i+n].encode("utf-8")
                h = int(hashlib.md5(gram).hexdigest(), 16)
                idx = h % dim
                sign = 1.0 if ((h >> 4) & 1) else -1.0
                vec[idx] += sign
        norm = np.linalg.norm(vec)
        if norm > 0:
            vec /= norm
        return vec

    def _build_node_features(self):
        """Constructs rich continuous feature vectors for every node."""
        type_dim = len(self.KNOWN_NODE_TYPES) + 1  # +1 for unknown type
        path_dim = self.feature_dim - type_dim - 2  # remaining dims for text hash, 2 dims for depth & size

        features = []
        for node in self.idx_to_node:
            feat = []

            # 1. One-hot node type
            ntype = node.get("type", "").lower()
            type_vec = [0.0] * type_dim
            if ntype in self.KNOWN_NODE_TYPES:
                type_vec[self.KNOWN_NODE_TYPES.index(ntype)] = 1.0
            else:
                type_vec[-1] = 1.0
            feat.extend(type_vec)

            # 2. Structural metrics (depth, line count if available)
            path_str = node.get("path", "")
            depth = float(path_str.count(os.sep) if path_str else 0)
            depth_scaled = math.tanh(depth / 5.0)

            lines = 0.0
            props = node.get("properties", {}) or {}
            if "lines" in props:
                try:
                    lines = float(props["lines"])
                except (ValueError, TypeError):
                    lines = 0.0
            lines_scaled = math.log1p(lines) / 10.0

            feat.append(depth_scaled)
            feat.append(lines_scaled)

            # 3. Label & path semantic text hashing
            label_text = f"{node.get('label', '')} {path_str} {props.get('language', '')}"
            text_vec = self._hash_string(label_text, path_dim)
            feat.extend(text_vec.tolist())

            features.append(feat)

        self.x = torch.tensor(features, dtype=torch.float32)

    def _build_edges(self):
        src_list = []
        dst_list = []
        type_list = []
        self.edge_set = set()

        for edge in self.edges:
            src_id = edge.get("source")
            dst_id = edge.get("target")
            rel = edge.get("relation")

            if src_id not in self.node_to_idx or dst_id not in self.node_to_idx:
                continue

            u = self.node_to_idx[src_id]
            v = self.node_to_idx[dst_id]
            r = self.rel_to_idx[rel]
            r_inv = self.rel_to_idx[f"inv_{rel}"]

            # Forward edge
            src_list.append(u)
            dst_list.append(v)
            type_list.append(r)
            self.edge_set.add((u, r, v))

            # Backward edge
            src_list.append(v)
            dst_list.append(u)
            type_list.append(r_inv)
            self.edge_set.add((v, r_inv, u))

        # Add self-loops
        self_r = self.rel_to_idx[self.self_loop_rel]
        for i in range(self.num_nodes):
            src_list.append(i)
            dst_list.append(i)
            type_list.append(self_r)

        self.edge_index = torch.tensor([src_list, dst_list], dtype=torch.long)
        self.edge_type = torch.tensor(type_list, dtype=torch.long)
        self.num_edges = len(src_list)

    def sample_negative_edges(self, num_samples: int) -> Tuple[torch.Tensor, torch.Tensor]:
        """
        Samples corrupt triples (u, r, v_neg) where (u, r, v_neg) does not exist in the graph.
        """
        neg_src = []
        neg_dst = []
        neg_types = []

        base_rel_indices = [self.rel_to_idx[r] for r in self.base_relations]
        if not base_rel_indices:
            base_rel_indices = list(range(self.num_relations))

        attempts = 0
        max_attempts = num_samples * 10
        while len(neg_src) < num_samples and attempts < max_attempts:
            attempts += 1
            u = np.random.randint(0, self.num_nodes)
            v = np.random.randint(0, self.num_nodes)
            r = np.random.choice(base_rel_indices)

            if u != v and (u, r, v) not in self.edge_set:
                neg_src.append(u)
                neg_dst.append(v)
                neg_types.append(r)

        if len(neg_src) < num_samples:
            # Fallback padding if graph is dense
            for _ in range(num_samples - len(neg_src)):
                neg_src.append(0)
                neg_dst.append(min(1, self.num_nodes - 1))
                neg_types.append(base_rel_indices[0])

        neg_edge_index = torch.tensor([neg_src, neg_dst], dtype=torch.long)
        neg_edge_type = torch.tensor(neg_types, dtype=torch.long)
        return neg_edge_index, neg_edge_type

    def get_train_val_split(self, val_ratio: float = 0.2) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor, torch.Tensor]:
        """Splits the forward non-self-loop edges into training and validation sets."""
        forward_mask = []
        for i in range(self.num_edges):
            r = self.edge_type[i].item()
            rel_name = self.idx_to_rel[r]
            forward_mask.append(not rel_name.startswith("inv_") and rel_name != self.self_loop_rel)

        forward_indices = np.where(forward_mask)[0]
        np.random.shuffle(forward_indices)

        n_val = max(1, int(len(forward_indices) * val_ratio))
        val_idx = forward_indices[:n_val]
        train_idx = forward_indices[n_val:]

        return (
            torch.tensor(train_idx, dtype=torch.long),
            torch.tensor(val_idx, dtype=torch.long),
            self.edge_index,
            self.edge_type
        )
