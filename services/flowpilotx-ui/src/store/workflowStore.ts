import { create } from 'zustand';
import { Node, Edge, XYPosition } from 'reactflow';
import { WorkflowState, NodeUpdater, EdgeUpdater, NodeData, NodeType } from '../types/workflow';

export type Theme = 'vscode' | 'miro';

// Calculate node outputs based on type and inputs
const calculateNodeOutputs = (type: NodeType, inputs: { [key: string]: any }) => {
  switch (type) {
    case 'add': {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: a + b };
    }
    case 'multiply': {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: a * b };
    }
    default:
      return { result: 0 };
  }
};

interface WorkflowStore extends WorkflowState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  addNode: (node: Node) => void;
  addEdge: (edge: Edge) => void;
  updateNodePosition: (nodeId: string, position: XYPosition) => void;
  updateNodeData: (nodeId: string, data: NodeData) => void;
  deleteNode: (nodeId: string) => void;
  setNodes: (updater: NodeUpdater) => void;
  setEdges: (updater: EdgeUpdater) => void;
}

export const useWorkflowStore = create<WorkflowStore>((set) => ({
  nodes: [],
  edges: [],
  theme: 'vscode', // Default theme
  setTheme: (theme) => set({ theme }),
  addNode: (node) =>
    set((state) => ({
      nodes: [...state.nodes, node],
    })),
  addEdge: (edge) =>
    set((state) => ({
      edges: [...state.edges, edge],
    })),
  updateNodePosition: (nodeId, position) =>
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === nodeId ? { ...node, position } : node
      ),
    })),
  updateNodeData: (nodeId, data) =>
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === nodeId
          ? {
              ...node,
              data: {
                ...data,
                outputs: calculateNodeOutputs(data.type, data.inputs),
              },
            }
          : node
      ),
    })),
  deleteNode: (nodeId) =>
    set((state) => ({
      nodes: state.nodes.filter((node) => node.id !== nodeId),
      edges: state.edges.filter(
        (edge) => edge.source !== nodeId && edge.target !== nodeId
      ),
    })),
  setNodes: (updater) =>
    set((state) => ({
      nodes: typeof updater === 'function' ? updater(state.nodes) : updater,
    })),
  setEdges: (updater) =>
    set((state) => ({
      edges: typeof updater === 'function' ? updater(state.edges) : updater,
    })),
})); 