import { Node as ReactFlowNode, Edge as ReactFlowEdge, NodeChange, EdgeChange, NodeProps } from 'reactflow';
import { ReactNode } from 'react';
import { nodeDefinitions } from '../data/nodeDefinitions';
import React from 'react';

export type NodeType = keyof typeof nodeDefinitions | 'sticky';

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
  text?: string;
  color?: string;
  opacity?: number;
  width?: number;
  height?: number;
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
  type: 'number' | 'text' | 'boolean' | 'select' | 'array';
  label: string;
  default?: any;
  description?: string;
  optional?: boolean;
  options?: string[];
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

export type NodeTypes = {
  [key: string]: React.ComponentType<NodeProps<NodeData>>;
}; 