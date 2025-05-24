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
  PlusCircleIcon
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

export const BaseNode = memo<NodeProps<NodeData>>(({ data, selected, id, type }) => {
  const navigate = useNavigate();
  const deleteNode = useWorkflowStore(state => state.deleteNode);
  const theme = useWorkflowStore(state => state.theme);
  const setSidebarExpanded = useWorkflowStore(state => state.setSidebarExpanded);
  const setSourceNodeId = useWorkflowStore(state => state.setSourceNodeId);
  const sourceNodeId = useWorkflowStore(state => state.sourceNodeId);
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

  const handleExpandSidebar = (e: React.MouseEvent) => {
    e.stopPropagation();
    setSourceNodeId(id);
    setSidebarExpanded(true);
  };

  const isVSCode = theme === 'vscode';
  const nodeTheme = isVSCode ? 'node-vscode' : 'node-miro';
  const nodeColor = data.type === 'add' ? 
    (isVSCode ? 'node-vscode-add' : 'node-miro-add') : 
    (isVSCode ? 'node-vscode-multiply' : 'node-miro-multiply');

  const handleStyle = {
    input: `w-3 h-3 !bg-gray-400 border-2 ${isVSCode ? 'border-[#1e1e1e]' : 'border-white'} rounded-full transition-all duration-200 hover:scale-125`,
    output: `w-4 h-4 !bg-gray-400 border-2 ${isVSCode ? 'border-[#1e1e1e]' : 'border-white'} rounded-full transition-all duration-200 hover:scale-125`
  };

  const isSourceNode = sourceNodeId === id;

  return (
    <div 
      className={`
        relative transition-all duration-150 
        ${!isActive ? 'opacity-50' : ''}
        ${isSourceNode ? 'ring-2 ring-green-500 ring-offset-2' : ''}
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

        {/* Add Button */}
        <button
          onClick={handleExpandSidebar}
          className={`
            absolute right-0 top-1/2 transform translate-x-[140%] -translate-y-1/2
            w-6 h-6 rounded-full
            flex items-center justify-center
            transition-all duration-200
            ${isVSCode ? 
              'bg-node-vscode-bg border-node-vscode-border text-node-vscode-text hover:text-white' : 
              'bg-node-miro-bg border-node-miro-border text-node-miro-text hover:text-gray-700'
            }
            border-2 shadow-lg
            hover:scale-110
            group-hover:opacity-100 opacity-0
            ${isSourceNode ? 'ring-2 ring-green-500' : ''}
          `}
          title="Add Connected Node"
        >
          <PlusIcon className="w-4 h-4" />
        </button>

        {/* Action Buttons - Show on Hover */}
        {showActions && (
          <div className={`
            absolute -top-7 left-1/2 transform -translate-x-1/2 
            flex items-center gap-0.5 p-0.5 rounded-md
            ${isVSCode ? 
              'bg-transparent text-node-vscode-text' : 
              'bg-transparent text-node-miro-text'
            }
            opacity-0 group-hover:opacity-100
            transition-all duration-200
          `}>
            <button
              onClick={handleSelect}
              className="p-0.5 rounded hover:text-white"
              title="Settings"
            >
              <Cog6ToothIcon className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleRun}
              className="p-0.5 rounded hover:text-white"
              title="Run"
            >
              <PlayIcon className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleToggleActive}
              className={`p-0.5 rounded hover:text-white ${!isActive ? 'text-gray-500' : ''}`}
              title={isActive ? 'Deactivate' : 'Activate'}
            >
              <PowerIcon className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleDelete}
              className="p-0.5 rounded hover:text-white text-red-500"
              title="Delete"
            >
              <TrashIcon className="w-3.5 h-3.5" />
            </button>
          </div>
        )}

        {/* Simple Info Footer */}
        <div className={`
          absolute -bottom-10 left-1/2 transform -translate-x-1/2
          w-max max-w-[180px] px-1.5 py-1
          text-center
        `}>
          <div className={`text-xs font-medium ${isVSCode ? 'text-node-vscode-text' : 'text-node-miro-text'}`}>
            {data.name}
          </div>
          {data.description && (
            <div className={`
              text-[10px] mt-0.5
              ${isVSCode ? 'text-node-vscode-text/70' : 'text-node-miro-text/70'}
            `}>
              {data.description}
            </div>
          )}
        </div>

        {/* Handles */}
        {Object.keys(data.inputs).map((key) => (
          <Handle
            key={`input-${key}`}
            type="target"
            position={Position.Left}
            id={`input-${key}`}
            className={handleStyle.input}
            style={{ left: -6 }}
          />
        ))}
        {Object.keys(data.outputs).map((key) => (
          <Handle
            key={`output-${key}`}
            type="source"
            position={Position.Right}
            id={`output-${key}`}
            className={handleStyle.output}
            style={{ right: -8 }}
          />
        ))}
      </div>
    </div>
  );
}); 