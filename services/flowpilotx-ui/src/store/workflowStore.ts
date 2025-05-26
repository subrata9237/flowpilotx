import { create } from 'zustand';
import { Node, Edge, XYPosition, Viewport } from 'reactflow';
import { WorkflowState, NodeUpdater, EdgeUpdater, NodeData, NodeType } from '../types/workflow';
import { nodeDefinitions } from '../data/nodeDefinitions';

export type Theme = 'vscode' | 'miro';

// Calculate node outputs based on type and inputs
const calculateNodeOutputs = (type: NodeType, inputs: { [key: string]: any }) => {
  const nodeDefinition = nodeDefinitions[type];
  if (!nodeDefinition) return {};
  return nodeDefinition.calculate(inputs);
};

interface WorkflowStore extends WorkflowState {
  theme: Theme;
  editorMode: 'click' | 'pan';
  setEditorMode: (mode: 'click' | 'pan') => void;
  isSidebarExpanded: boolean;
  sourceNodeId: string | null;
  viewport: Viewport | null;
  setTheme: (theme: Theme) => void;
  setSidebarExpanded: (expanded: boolean) => void;
  setSourceNodeId: (nodeId: string | null) => void;
  addNode: (node: Node<NodeData>) => void;
  addEdge: (edge: Edge) => void;
  updateNodePosition: (nodeId: string, position: XYPosition) => void;
  updateNodeData: (nodeId: string, data: NodeData) => void;
  deleteNode: (nodeId: string) => void;
  setNodes: (updater: NodeUpdater) => void;
  setEdges: (updater: EdgeUpdater) => void;
  connectNodes: (sourceId: string, targetNode: Node<NodeData>) => void;
  setViewport: (viewport: Viewport) => void;
}

export const useWorkflowStore = create<WorkflowStore>((set, get) => ({
  nodes: [],
  edges: [],
  theme: 'vscode', // Default theme
  editorMode: 'click', // Default mode
  isSidebarExpanded: false, // Default sidebar state
  sourceNodeId: null, // Track source node for auto-connection
  viewport: null,
  setTheme: (theme) => set({ theme }),
  setEditorMode: (mode) => set({ editorMode: mode }),
  setSidebarExpanded: (expanded) => {
    set({ isSidebarExpanded: expanded });
    if (!expanded) {
      // Clear source node when closing sidebar
      set({ sourceNodeId: null });
    }
  },
  setSourceNodeId: (nodeId) => set({ sourceNodeId: nodeId }),
  addNode: (node) =>
    set((state) => {
      const nodes = [...state.nodes, node];
      // If we have a source node, automatically connect it
      if (state.sourceNodeId) {
        const sourceNode = state.nodes.find(n => n.id === state.sourceNodeId);
        if (sourceNode) {
          const sourceOutput = Object.keys(sourceNode.data.outputs)[0];
          const targetInput = Object.keys(node.data.inputs)[0];
          const newEdge: Edge = {
            id: `${state.sourceNodeId}-${node.id}`,
            source: state.sourceNodeId,
            target: node.id,
            sourceHandle: `output-${sourceOutput}`,
            targetHandle: `input-${targetInput}`,
          };
          return {
            nodes,
            edges: [...state.edges, newEdge],
            sourceNodeId: null, // Clear source node after connection
            isSidebarExpanded: false, // Close sidebar after connection
          };
        }
      }
      return { nodes };
    }),
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
      sourceNodeId: state.sourceNodeId === nodeId ? null : state.sourceNodeId,
    })),
  setNodes: (updater) =>
    set((state) => ({
      nodes: typeof updater === 'function' ? updater(state.nodes) : updater,
    })),
  setEdges: (updater) =>
    set((state) => ({
      edges: typeof updater === 'function' ? updater(state.edges) : updater,
    })),
  connectNodes: (sourceId, targetNode) => {
    const state = get();
    const sourceNode = state.nodes.find(n => n.id === sourceId);
    if (sourceNode) {
      const sourceOutput = Object.keys(sourceNode.data.outputs)[0];
      const targetInput = Object.keys(targetNode.data.inputs)[0];
      const newEdge: Edge = {
        id: `${sourceId}-${targetNode.id}`,
        source: sourceId,
        target: targetNode.id,
        sourceHandle: `output-${sourceOutput}`,
        targetHandle: `input-${targetInput}`,
      };
      set((state) => ({
        edges: [...state.edges, newEdge],
      }));
    }
  },
  setViewport: (viewport) =>
    set(() => ({
      viewport,
    })),
})); 