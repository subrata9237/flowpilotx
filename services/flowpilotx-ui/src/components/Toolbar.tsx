import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { Node, NodeData } from '../types/workflow';
import { nodeDefinitions } from '../data/nodeDefinitions';
``
export const Toolbar: React.FC = () => {
  const addNode = useWorkflowStore((state) => state.addNode);

  const createNode = (type: 'add' | 'multiply') => {
    const nodeDefinition = nodeDefinitions[type];
    
    // Initialize config with default values from node definition
    const config: Record<string, any> = {};
    if (nodeDefinition) {
      Object.entries(nodeDefinition.settings.config).forEach(([key, setting]) => {
        config[key] = setting.default;
      });
    }

    const newNode: Node = {
      id: `${type}-${Date.now()}`,
      type,
      position: { x: 100, y: 100 },
      data: {
        name: type.charAt(0).toUpperCase() + type.slice(1),
        type,
        inputs: { a: 0, b: 0 },
        outputs: { result: 0 },
        config: config  // Add the initialized config
      } as NodeData,
    };
    addNode(newNode);
  };

  return (
    <div className="absolute top-5 left-5 z-10 flex gap-2">
      <button
        onClick={() => createNode('add')}
        className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors"
      >
        Add Node
      </button>
      <button
        onClick={() => createNode('multiply')}
        className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors"
      >
        Multiply Node
      </button>
    </div>
  );
}; 