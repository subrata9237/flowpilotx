import React, { memo, useState } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { NodeData } from '../types/workflow';
import { 
  Cog6ToothIcon,
  PlayIcon,
  PowerIcon,
  TrashIcon,
  PlusIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { useNavigate } from 'react-router-dom';
import { useWorkflowStore } from '../store/workflowStore';

const getNodeIcon = (type: string) => {
  switch (type) {
    case 'add':
      return <PlusIcon className="w-6 h-6" />;
    case 'multiply':
      return <XMarkIcon className="w-6 h-6" />;
    default:
      return null;
  }
};

export const BaseNode = memo<NodeProps<NodeData>>(({ data, selected, id }) => {
  const navigate = useNavigate();
  const deleteNode = useWorkflowStore(state => state.deleteNode);
  const theme = useWorkflowStore(state => state.theme);
  const [isActive, setIsActive] = useState(true);
  const [showActions, setShowActions] = useState(false);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    const nodeElement = document.querySelector(`[data-id="${id}"]`);
    if (nodeElement) {
      nodeElement.classList.add('scale-95', 'opacity-0');
      setTimeout(() => deleteNode(id), 150);
    }
  };

  const handleRun = (e: React.MouseEvent) => {
    e.stopPropagation();
    const nodeElement = document.querySelector(`[data-id="${id}"]`);
    if (nodeElement) {
      nodeElement.classList.add('animate-pulse');
      setTimeout(() => nodeElement.classList.remove('animate-pulse'), 1000);
    }
    console.log('Running node:', id);
  };

  const handleToggleActive = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsActive(!isActive);
  };

  const handleSelect = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/node/${id}/settings`);
  };

  const isVSCode = theme === 'vscode';
  const nodeTheme = isVSCode ? 'node-vscode' : 'node-miro';
  const nodeColor = data.type === 'add' ? 
    (isVSCode ? 'node-vscode-add' : 'node-miro-add') : 
    (isVSCode ? 'node-vscode-multiply' : 'node-miro-multiply');

  return (
    <div 
      className={`
        relative transition-all duration-150 
        ${!isActive ? 'opacity-50' : ''}
        group
      `}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      onDoubleClick={(e) => {
        e.stopPropagation();
        navigate(`/node/${id}/settings`);
      }}
    >
      {/* Node Body */}
      <div className={`
        w-16 h-16 
        ${isVSCode ? 'bg-node-vscode-bg border-node-vscode-border' : 'bg-node-miro-bg border-node-miro-border'}
        rounded-lg 
        ${selected ? 
          (isVSCode ? 'shadow-node-selected-vscode' : 'shadow-node-selected-miro') : 
          (isVSCode ? 'shadow-node-vscode' : 'shadow-node-miro')
        }
        border-2
        ${data.type === 'add' ? 
          (isVSCode ? 'text-node-vscode-add' : 'text-node-miro-add') : 
          (isVSCode ? 'text-node-vscode-multiply' : 'text-node-miro-multiply')
        }
        hover:animate-node-hover
        transition-all duration-150
        flex items-center justify-center
        relative
        cursor-pointer
      `}>
        {/* Main Icon */}
        <div className="p-2 rounded-md transition-all duration-150">
          {getNodeIcon(data.type)}
        </div>

        {/* Action Buttons - Show on Hover */}
        {showActions && (
          <div className={`
            absolute -top-8 left-1/2 transform -translate-x-1/2 
            flex items-center gap-1 p-1 rounded-md
            ${isVSCode ? 
              'bg-node-vscode-bg border-node-vscode-border text-node-vscode-text' : 
              'bg-node-miro-bg border-node-miro-border text-node-miro-text'
            }
            border shadow-lg
            opacity-0 group-hover:opacity-100
            transition-all duration-200
          `}>
            <button
              onClick={handleSelect}
              className="p-1 rounded hover:bg-gray-700/50"
              title="Settings"
            >
              <Cog6ToothIcon className="w-4 h-4" />
            </button>
            <button
              onClick={handleRun}
              className="p-1 rounded hover:bg-gray-700/50"
              title="Run"
            >
              <PlayIcon className="w-4 h-4" />
            </button>
            <button
              onClick={handleToggleActive}
              className={`p-1 rounded hover:bg-gray-700/50 ${!isActive ? 'text-gray-500' : ''}`}
              title={isActive ? 'Deactivate' : 'Activate'}
            >
              <PowerIcon className="w-4 h-4" />
            </button>
            <button
              onClick={handleDelete}
              className="p-1 rounded hover:bg-red-500/50 text-red-500"
              title="Delete"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* Handles */}
        {Object.keys(data.inputs).map((key) => (
          <Handle
            key={`input-${key}`}
            type="target"
            position={Position.Left}
            id={`input-${key}`}
            className="w-2 h-2 !bg-gray-400 border-2 border-white dark:border-gray-800"
          />
        ))}
        {Object.keys(data.outputs).map((key) => (
          <Handle
            key={`output-${key}`}
            type="source"
            position={Position.Right}
            id={`output-${key}`}
            className="w-2 h-2 !bg-gray-400 border-2 border-white dark:border-gray-800"
          />
        ))}
      </div>
    </div>
  );
}); 