import { create } from 'zustand';
import { Node, Edge, XYPosition, Viewport, Position } from 'reactflow';
import { WorkflowState, NodeUpdater, EdgeUpdater, NodeData, NodeType } from '../types/workflow';
import { nodeDefinitions } from '../data/nodeDefinitions';

export type Theme = 'vscode' | 'miro';

// Calculate node outputs based on type and inputs
const calculateNodeOutputs = (type: NodeType, inputs: { [key: string]: any }) => {
  if (type === 'sticky') return {};
  const nodeDefinition = nodeDefinitions[type];
  if (!nodeDefinition) return {};
  return nodeDefinition.calculate(inputs);
};

interface StickyNote {
  id: string;
  text: string;
  x: number;
  y: number;
  color: string;
  width?: number;
  height?: number;
}

interface SavedWorkflow {
  id: string;
  name: string;
  nodes: Node[];
  edges: Edge[];
  updatedAt: number;
}

interface WorkflowStore extends WorkflowState {
  theme: Theme;
  editorMode: 'click' | 'pan';
  stickyNotes: StickyNote[];
  isSidebarExpanded: boolean;
  sourceNodeId: string | null;
  viewport: { x: number; y: number; zoom: number } | null;
  showStickyNotes: boolean;
  workflows: SavedWorkflow[];
  currentProjectId: string | null;
  currentWorkflowId: string | null;
  setEditorMode: (mode: 'click' | 'pan') => void;
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
  setViewport: (viewport: { x: number; y: number; zoom: number } | null) => void;
  addStickyNote: (note: Omit<StickyNote, 'id'>) => void;
  updateStickyNote: (id: string, updates: Partial<Omit<StickyNote, 'id'>>) => void;
  deleteStickyNote: (id: string) => void;
  toggleStickyNotes: () => void;
  addOrUpdateWorkflow: (workflow: SavedWorkflow) => void;
  deleteWorkflow: (workflowId: string) => void;
  getWorkflows: () => SavedWorkflow[];
  setCurrentProjectId: (id: string | null) => void;
  setCurrentWorkflowId: (id: string | null) => void;
}

export const useWorkflowStore = create<WorkflowStore>((set, get) => ({
  nodes: [],
  edges: [],
  theme: 'vscode', // Default theme
  editorMode: 'click', // Default mode
  isSidebarExpanded: false, // Default sidebar state
  sourceNodeId: null, // Track source node for auto-connection
  viewport: null,
  stickyNotes: [],
  showStickyNotes: true, // Default to showing sticky notes
  workflows: [],
  currentProjectId: null,
  currentWorkflowId: null,
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
  addStickyNote: (note) => set((state) => ({
    stickyNotes: [...state.stickyNotes, { 
      ...note, 
      id: `sticky-${Date.now()}`,
      width: note.width ?? 200,
      height: note.height ?? 120
    }]
  })),
  updateStickyNote: (id, updates) => set((state) => ({
    stickyNotes: state.stickyNotes.map(note =>
      note.id === id ? { ...note, ...updates } : note
    )
  })),
  deleteStickyNote: (id) => set((state) => ({
    stickyNotes: state.stickyNotes.filter(note => note.id !== id)
  })),
  toggleStickyNotes: () => set((state) => ({ showStickyNotes: !state.showStickyNotes })),
  addOrUpdateWorkflow: (workflow) => set((state) => {
    const existing = state.workflows.find(w => w.id === workflow.id);
    if (existing) {
      return {
        workflows: state.workflows.map(w => w.id === workflow.id ? { ...workflow, updatedAt: Date.now() } : w)
      };
    } else {
      return {
        workflows: [
          ...state.workflows,
          { ...workflow, updatedAt: Date.now() }
        ]
      };
    }
  }),
  deleteWorkflow: (workflowId) => set((state) => ({
    workflows: state.workflows.filter(w => w.id !== workflowId)
  })),
  getWorkflows: () => get().workflows,
  setCurrentProjectId: (id) => set({ currentProjectId: id }),
  setCurrentWorkflowId: (id) => set({ currentWorkflowId: id }),
})); 