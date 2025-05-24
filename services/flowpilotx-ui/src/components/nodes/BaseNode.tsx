import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { NodeData } from '../../types/workflow';

export const BaseNode: React.FC<NodeProps<NodeData>> = ({ data }) => {
  return (
    <div className="bg-gray-800 rounded-lg min-w-[200px] shadow-lg">
      <div className="p-4">
        <h6 className="text-lg font-semibold text-white">{data.name}</h6>
        <p className="text-gray-400">{data.type}</p>
        
        <div className="mt-4">
          <h6 className="text-sm font-medium text-gray-300 mb-2">Inputs</h6>
          {Object.entries(data.inputs).map(([key, value]) => (
            <div key={key} className="flex items-center mb-2 relative">
              <Handle
                type="target"
                position={Position.Left}
                id={`input-${key}`}
                className="node-handle node-handle-left"
              />
              <span className="text-gray-200">{key}: {value}</span>
            </div>
          ))}
        </div>

        <div className="mt-4">
          <h6 className="text-sm font-medium text-gray-300 mb-2">Outputs</h6>
          {Object.entries(data.outputs).map(([key, value]) => (
            <div key={key} className="flex items-center mb-2 relative">
              <span className="text-gray-200">{key}: {value}</span>
              <Handle
                type="source"
                position={Position.Right}
                id={`output-${key}`}
                className="node-handle node-handle-right"
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}; 