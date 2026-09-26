import argparse
import datetime
import json
import math
import os
import sys
from typing import List, Dict, Any

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F

from dataset import KnowledgeGraphDataset
from model import RelationalGNN

def parse_args():
    parser = argparse.ArgumentParser(description="Train Graph Neural Network on BitConfig Knowledge Graph")
    parser.add_argument("--graph", type=str, default="./knowledge_graph.json", help="Path to input knowledge_graph.json")
    parser.add_argument("--output", type=str, default="./knowledge_graph_ml.json", help="Path to output ML enriched JSON")
    parser.add_argument("--epochs", type=int, default=60, help="Number of training epochs")
    parser.add_argument("--lr", type=float, default=0.015, help="Learning rate")
    parser.add_argument("--hidden-dim", type=int, default=64, help="Hidden embedding dimension")
    parser.add_argument("--out-dim", type=int, default=32, help="Output representation dimension")
    parser.add_argument("--seed", type=int, default=42, help="Random seed")
    return parser.parse_args()

def compute_roc_auc(pos_scores: np.ndarray, neg_scores: np.ndarray) -> float:
    """Computes ROC-AUC metric without requiring external heavy libraries."""
    y_true = np.concatenate([np.ones_like(pos_scores), np.zeros_like(neg_scores)])
    y_scores = np.concatenate([pos_scores, neg_scores])

    # Rank-based AUC computation (Mann-Whitney U statistic)
    order = np.argsort(y_scores)
    rank = np.empty_like(order, dtype=float)
    rank[order] = np.arange(len(y_scores)) + 1.0

    n_pos = len(pos_scores)
    n_neg = len(neg_scores)
    if n_pos == 0 or n_neg == 0:
        return 0.5

    pos_ranks = rank[:n_pos]
    auc = (np.sum(pos_ranks) - n_pos * (n_pos + 1) / 2.0) / (n_pos * n_neg)
    return float(auc)

def main():
    args = parse_args()

    if not os.path.exists(args.graph):
        print(f"Error: Graph file '{args.graph}' not found.")
        print("Run 'bitconfig graph build' first to generate the knowledge graph.")
        sys.exit(1)

    # Select best hardware device
    if torch.backends.mps.is_available():
        device = torch.device("mps")
        device_name = "Apple Silicon (MPS)"
    elif torch.cuda.is_available():
        device = torch.device("cuda")
        device_name = f"CUDA ({torch.cuda.get_device_name(0)})"
    else:
        device = torch.device("cpu")
        device_name = "CPU"

    print("=" * 60)
    print(" BitConfig GNN - Graph Neural Network Training Pipeline")
    print("=" * 60)
    print(f"Device:            {device_name}")
    print(f"Knowledge Graph:   {args.graph}")
    print(f"Epochs:            {args.epochs}")
    print(f"Learning Rate:     {args.lr}")

    # 1. Dataset loading & tensor conversion
    dataset = KnowledgeGraphDataset(json_path=args.graph, feature_dim=args.hidden_dim, seed=args.seed)
    print(f"Graph Statistics:  {dataset.num_nodes} nodes, {len(dataset.edges)} edges, {dataset.num_relations} relation types")

    x = dataset.x.to(device)
    edge_index = dataset.edge_index.to(device)
    edge_type = dataset.edge_type.to(device)

    # Split forward edges into train / val
    train_idx, val_idx, _, _ = dataset.get_train_val_split(val_ratio=0.2)
    train_idx = train_idx.to(device)
    val_idx = val_idx.to(device)

    # Compute target degree centrality for criticality head supervision
    degrees = torch.zeros(dataset.num_nodes, dtype=torch.float32, device=device)
    degrees.scatter_add_(0, edge_index[1], torch.ones_like(edge_index[1], dtype=torch.float32))
    target_crit = torch.log1p(degrees)
    max_crit = torch.max(target_crit)
    if max_crit > 0:
        target_crit = target_crit / max_crit

    # 2. Model initialization
    num_bases = min(8, dataset.num_relations)
    model = RelationalGNN(
        in_dim=dataset.x.size(1),
        hidden_dim=args.hidden_dim,
        out_dim=args.out_dim,
        num_relations=dataset.num_relations,
        num_bases=num_bases,
        dropout=0.1,
    ).to(device)

    optimizer = torch.optim.AdamW(model.parameters(), lr=args.lr, weight_decay=1e-4)
    scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=args.epochs, eta_min=1e-4)

    # 3. Training Loop
    print("\nTraining GNN layers and learning relational representations...")
    best_loss = float("inf")
    final_val_auc = 0.5

    for epoch in range(1, args.epochs + 1):
        model.train()
        optimizer.zero_grad()

        # Positive train edges
        pos_src = edge_index[0, train_idx]
        pos_dst = edge_index[1, train_idx]
        pos_rel = edge_type[train_idx]

        # Negative train edges (contrastive sampling)
        neg_edge_idx, neg_rel = dataset.sample_negative_edges(len(train_idx))
        neg_src = neg_edge_idx[0].to(device)
        neg_dst = neg_edge_idx[1].to(device)
        neg_rel = neg_rel.to(device)

        # Forward pass on message graph
        h = model.encode(x, edge_index, edge_type)

        pos_scores = model.decode_links(h, pos_src, pos_dst, pos_rel)
        neg_scores = model.decode_links(h, neg_src, neg_dst, neg_rel)

        crit = model.predict_criticality(h)

        # Contrastive BCE link prediction loss
        pos_loss = F.binary_cross_entropy_with_logits(pos_scores, torch.ones_like(pos_scores))
        neg_loss = F.binary_cross_entropy_with_logits(neg_scores, torch.zeros_like(neg_scores))
        link_loss = pos_loss + neg_loss

        # Criticality alignment loss
        crit_loss = F.mse_loss(crit, target_crit)

        total_loss = link_loss + 0.25 * crit_loss
        total_loss.backward()
        optimizer.step()
        scheduler.step()

        # Validation step
        if epoch % 10 == 0 or epoch == args.epochs:
            model.eval()
            with torch.no_grad():
                h_val = model.encode(x, edge_index, edge_type)
                val_pos_src = edge_index[0, val_idx]
                val_pos_dst = edge_index[1, val_idx]
                val_pos_rel = edge_type[val_idx]

                neg_val_idx, neg_val_rel = dataset.sample_negative_edges(len(val_idx))
                val_neg_src = neg_val_idx[0].to(device)
                val_neg_dst = neg_val_idx[1].to(device)
                val_neg_rel = neg_val_rel.to(device)

                val_pos = torch.sigmoid(model.decode_links(h_val, val_pos_src, val_pos_dst, val_pos_rel)).cpu().numpy()
                val_neg = torch.sigmoid(model.decode_links(h_val, val_neg_src, val_neg_dst, val_neg_rel)).cpu().numpy()

                val_auc = compute_roc_auc(val_pos, val_neg)
                final_val_auc = val_auc

                print(f"  Epoch [{epoch:02d}/{args.epochs:02d}] | Loss: {total_loss.item():.4f} (Link: {link_loss.item():.4f}) | Val ROC-AUC: {val_auc:.4f}")

    # 4. Inference & High-Value Predictions
    print("\nGenerating ML architectural inferences...")
    model.eval()
    with torch.no_grad():
        final_h = model.encode(x, edge_index, edge_type)
        final_crit = model.predict_criticality(final_h).cpu().numpy()

    # Rank nodes by criticality (blast radius)
    ranked_nodes = []
    for idx, score in enumerate(final_crit):
        node = dataset.idx_to_node[idx]
        ranked_nodes.append({
            "id": node["id"],
            "label": node.get("label", ""),
            "type": node.get("type", ""),
            "path": node.get("path", ""),
            "criticality_score": round(float(score), 4),
        })
    ranked_nodes.sort(key=lambda item: item["criticality_score"], reverse=True)

    # Predict hidden couplings between file nodes that have no direct edge
    file_indices = [
        i for i, n in enumerate(dataset.idx_to_node)
        if n.get("type") in ("file", "dependency_file")
    ]

    predicted_couplings = []
    if len(file_indices) >= 2:
        candidate_src = []
        candidate_dst = []
        candidate_rels = []

        # Check candidate pairs
        for i in range(len(file_indices)):
            for j in range(i + 1, len(file_indices)):
                u = file_indices[i]
                v = file_indices[j]
                # Default relation index for general coupling
                rel_idx = 0
                if (u, rel_idx, v) not in dataset.edge_set and (v, rel_idx, u) not in dataset.edge_set:
                    candidate_src.append(u)
                    candidate_dst.append(v)
                    candidate_rels.append(rel_idx)

        if candidate_src:
            cand_src_t = torch.tensor(candidate_src, dtype=torch.long, device=device)
            cand_dst_t = torch.tensor(candidate_dst, dtype=torch.long, device=device)
            cand_rel_t = torch.tensor(candidate_rels, dtype=torch.long, device=device)

            with torch.no_grad():
                cand_scores = torch.sigmoid(model.decode_links(final_h, cand_src_t, cand_dst_t, cand_rel_t)).cpu().numpy()

            for k in range(len(candidate_src)):
                score = cand_scores[k]
                u_node = dataset.idx_to_node[candidate_src[k]]
                v_node = dataset.idx_to_node[candidate_dst[k]]
                predicted_couplings.append({
                    "source": u_node.get("path") or u_node.get("label"),
                    "target": v_node.get("path") or v_node.get("label"),
                    "source_id": u_node["id"],
                    "target_id": v_node["id"],
                    "coupling_probability": round(float(score), 4),
                })

            predicted_couplings.sort(key=lambda item: item["coupling_probability"], reverse=True)
            predicted_couplings = predicted_couplings[:10]  # Keep top 10

    # 5. Export Enriched ML Results
    ml_output = {
        "project_name": dataset.project_name,
        "trained_at": datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "model_architecture": "Relational Graph Convolutional Network (R-GCN)",
        "metrics": {
            "device": device_name,
            "epochs": args.epochs,
            "final_loss": round(float(total_loss.item()), 4),
            "val_roc_auc": round(float(final_val_auc), 4),
            "num_nodes": dataset.num_nodes,
            "num_edges": len(dataset.edges),
            "num_relations": dataset.num_relations,
        },
        "critical_nodes": ranked_nodes[:15],
        "top_predicted_couplings": predicted_couplings,
    }

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(ml_output, f, indent=2)

    print("\n" + "=" * 60)
    print(" GNN Training Complete!")
    print("=" * 60)
    print(f"ML Artifact saved to: {args.output}")
    print("\nTop 5 Critical Files / Modules (Blast Radius Risk):")
    for item in ranked_nodes[:5]:
        print(f"  * [{item['type']}] {item['label']} (Score: {item['criticality_score']:.3f})")

    if predicted_couplings:
        print("\nTop Predicted Architectural Couplings:")
        for edge in predicted_couplings[:5]:
            print(f"  * {edge['source']} <---> {edge['target']} (Confidence: {edge['coupling_probability']*100:.1f}%)")

if __name__ == "__main__":
    main()
