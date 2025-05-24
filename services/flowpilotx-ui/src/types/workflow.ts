import { Node as ReactFlowNode, Edge as ReactFlowEdge, NodeChange, EdgeChange } from 'reactflow';
import { ReactNode } from 'react';

export type NodeType = 'add' | 'multiply';

export interface NodeTemplate {
  type: NodeType;
  name: string;
  icon: ReactNode;
  description: string;
  color: string;
  category: 'Math';
  defaults?: {
    inputs?: Record<string, any>;
    outputs?: Record<string, any>;
  };
}

export interface NodeCategory {
  name: string;
  icon: ReactNode;
  nodes: NodeTemplate[];
}

export interface NodeData {
  name: string;
  type: NodeType;
  description?: string;
  inputs: { [key: string]: any };
  outputs: { [key: string]: any };
}

export type Node = ReactFlowNode<NodeData>;
export type Edge = ReactFlowEdge;

export interface WorkflowState {
  nodes: Node[];
  edges: Edge[];
}

export type NodeUpdater = ((nodes: Node[]) => Node[]) | Node[];
export type EdgeUpdater = ((edges: Edge[]) => Edge[]) | Edge[];

export interface NodeTheme {
  bg: string;
  border: string;
  selected: string;
  hover: string;
  text: string;
  add: string;
  multiply: string;
  handle: string;
}

export interface ThemeColors {
  vscode: NodeTheme;
  miro: NodeTheme;
} 