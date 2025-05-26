import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { Node, NodeData } from '../types/workflow';
import { nodeDefinitions } from '../data/nodeDefinitions';

export const Toolbar: React.FC = () => {
  const addNode = useWorkflowStore((state) => state.addNode);

  const createNode = (type: string) => {
    const nodeDefinition = nodeDefinitions[type];
    if (!nodeDefinition) return;
    
    // Initialize config with default values from node definition
    const config: Record<string, any> = {};
    const inputs: Record<string, any> = {};
    const outputs: Record<string, any> = {};

    // Initialize inputs from node definition
    Object.entries(nodeDefinition.settings.inputs).forEach(([key, setting]) => {
      inputs[key] = setting.default;
    });

    // Initialize outputs from node definition
    Object.entries(nodeDefinition.settings.outputs).forEach(([key, setting]) => {
      outputs[key] = setting.default || 0;
    });

    // Initialize config
    Object.entries(nodeDefinition.settings.config).forEach(([key, setting]) => {
      config[key] = setting.default;
    });

    const newNode: Node = {
      id: `${type}-${Date.now()}`,
      type,
      position: { x: 100, y: 100 },
      data: {
        name: nodeDefinition.name,
        type,
        inputs,
        outputs,
        config
      } as NodeData,
    };
    addNode(newNode);
  };

  return (
    <div className="absolute top-5 left-5 z-10 flex gap-2">
      {Object.entries(nodeDefinitions).map(([type, def]) => (
        <button
          key={type}
          onClick={() => createNode(type)}
          className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors"
        >
          {def.name}
        </button>
      ))}
    </div>
  );
}; 