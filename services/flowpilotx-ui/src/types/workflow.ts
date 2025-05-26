import { Node as ReactFlowNode, Edge as ReactFlowEdge, NodeChange, EdgeChange } from 'reactflow';
import { ReactNode } from 'react';
import { nodeDefinitions } from '../data/nodeDefinitions';

export type NodeType = keyof typeof nodeDefinitions;

export interface NodeTemplate {
  type: NodeType;
  name: string;
  icon: React.ComponentType<{ className?: string }>;
  description: string;
  color: string;
  category: string;
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
  config?: { [key: string]: any };
  isActive?: boolean;
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

export interface NodeSettingField {
  type: 'number' | 'text' | 'boolean';
  label: string;
  default?: any;
  description?: string;
  optional?: boolean;
}

export interface NodeSettings {
  inputs: Record<string, NodeSettingField>;
  outputs: Record<string, NodeSettingField>;
  config: Record<string, NodeSettingField>;
}

export interface NodeDefinition {
  type: NodeType;
  name: string;
  icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>;
  description: string;
  color: string;
  category: string;
  calculate: (inputs: { [key: string]: any }) => { [key: string]: any };
  settings: NodeSettings;
} 