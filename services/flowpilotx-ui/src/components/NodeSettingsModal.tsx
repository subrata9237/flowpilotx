import React, { useState } from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { NodeData } from '../types/workflow';
import { useWorkflowStore } from '../store/workflowStore';

interface NodeSettingsModalProps {
  nodeId: string;
  data: NodeData;
  onClose: () => void;
}

export const NodeSettingsModal: React.FC<NodeSettingsModalProps> = ({ nodeId, data, onClose }) => {
  const updateNodeData = useWorkflowStore(state => state.updateNodeData);
  const [nodeData, setNodeData] = useState<NodeData>({ ...data });

  const handleSave = () => {
    updateNodeData(nodeId, nodeData);
    onClose();
  };

  const handleInputChange = (key: string, value: any) => {
    setNodeData(prev => ({
      ...prev,
      inputs: {
        ...prev.inputs,
        [key]: Number(value)
      }
    }));
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-[500px] max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Node Settings: {data.name}
          </h2>
          <button
            onClick={onClose}
            className="p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-6">
          {/* Description */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Description
            </label>
            <textarea
              value={nodeData.description || ''}
              onChange={(e) => setNodeData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Add a description..."
              className="w-full px-3 py-2 text-sm bg-gray-50 dark:bg-gray-700 border border-gray-200 
                dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
              rows={3}
            />
          </div>

          {/* Input Parameters */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Input Parameters
            </h3>
            <div className="space-y-3">
              {Object.entries(nodeData.inputs).map(([key, value]) => (
                <div key={key} className="flex items-center gap-4">
                  <label className="text-sm text-gray-600 dark:text-gray-400 w-24">
                    {key}:
                  </label>
                  <input
                    type="number"
                    value={value}
                    onChange={(e) => handleInputChange(key, e.target.value)}
                    className="flex-1 px-3 py-1.5 text-sm bg-gray-50 dark:bg-gray-700 border border-gray-200 
                      dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-accent-primary"
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Output Preview */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Output Preview
            </h3>
            <div className="p-3 bg-gray-50 dark:bg-gray-700 rounded-md">
              <div className="space-y-1">
                {Object.entries(nodeData.outputs).map(([key, value]) => (
                  <div key={key} className="flex items-center gap-2 text-sm">
                    <span className="text-gray-500 dark:text-gray-400">{key}:</span>
                    <span className="text-gray-900 dark:text-gray-100">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Node Type Info */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Node Information
            </h3>
            <div className="p-3 bg-gray-50 dark:bg-gray-700 rounded-md">
              <div className="space-y-1 text-sm">
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 dark:text-gray-400">Type:</span>
                  <span className="text-gray-900 dark:text-gray-100 capitalize">{nodeData.type}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 dark:text-gray-400">ID:</span>
                  <span className="text-gray-900 dark:text-gray-100 font-mono text-xs">{nodeId}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 
              hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-sm font-medium text-white bg-accent-primary 
              hover:bg-accent-primary/90 rounded-md transition-colors"
          >
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}; 