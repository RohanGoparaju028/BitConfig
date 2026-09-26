"""
BitConfig GNN Package
Graph Neural Network for Architectural Link Prediction and Node Criticality Scoring.
"""

from .dataset import KnowledgeGraphDataset
from .model import RelationalGNN

__all__ = ["KnowledgeGraphDataset", "RelationalGNN"]
