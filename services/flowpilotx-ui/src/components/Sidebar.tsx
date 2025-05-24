import React, { useState } from 'react';
import { NodeTemplate, NodeCategory } from '../types/workflow';
import { useWorkflowStore } from '../store/workflowStore';
import {
  MagnifyingGlassIcon,
  CalculatorIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  PlusIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

const nodeTemplates: NodeTemplate[] = [
  {
    type: 'add',
    name: 'Add Numbers',
    icon: <PlusIcon className="w-5 h-5" />,
    description: 'Add two numbers together',
    color: '#9B51E0',
    category: 'Math',
    defaults: {
      inputs: { a: 0, b: 0 },
      outputs: { result: 0 }
    }
  },
  {
    type: 'multiply',
    name: 'Multiply Numbers',
    icon: <XMarkIcon className="w-5 h-5" />,
    description: 'Multiply two numbers together',
    color: '#F2994A',
    category: 'Math',
    defaults: {
      inputs: { a: 0, b: 0 },
      outputs: { result: 0 }
    }
  }
];

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const categories: NodeCategory[] = [
  {
    name: 'Math',
    icon: <CalculatorIcon className="w-5 h-5" />,
    nodes: nodeTemplates.filter(n => n.category === 'Math')
  }
];

const SearchInput: React.FC<{
  value: string;
  onChange: (value: string) => void;
}> = ({ value, onChange }) => {
  const theme = useWorkflowStore((state) => state.theme);
  
  return (
    <div className="relative">
      <input
        type="text"
        placeholder="Search nodes..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={`
          w-full pl-9 pr-3 py-2 rounded-md text-sm
          transition-all duration-200
          focus:outline-none
          ${theme === 'vscode' ? `
            bg-node-vscode-input-bg
            border border-node-vscode-input-border
            text-node-vscode-text
            placeholder-node-vscode-input-placeholder
            hover:border-node-vscode-input-hover-border
            focus:border-node-vscode-input-focus-border
            focus:ring-1 focus:ring-node-vscode-input-focus-border
          ` : `
            bg-node-miro-input-bg
            border border-node-miro-input-border
            text-node-miro-text
            placeholder-node-miro-input-placeholder
            hover:border-node-miro-input-hover-border
            focus:border-node-miro-input-focus-border
            focus:ring-1 focus:ring-node-miro-input-focus-border
          `}
        `}
      />
      <MagnifyingGlassIcon 
        className={`w-4 h-4 absolute left-3 top-2.5 
          ${theme === 'vscode' ? 'text-node-vscode-input-placeholder' : 'text-node-miro-input-placeholder'}`} 
      />
    </div>
  );
};

export const Sidebar: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const theme = useWorkflowStore((state) => state.theme);
  const isCollapsed = !useWorkflowStore((state) => state.isSidebarExpanded);
  const setSidebarExpanded = useWorkflowStore((state) => state.setSidebarExpanded);
  const sourceNodeId = useWorkflowStore((state) => state.sourceNodeId);
  const addNode = useWorkflowStore((state) => state.addNode);
  const nodes = useWorkflowStore((state) => state.nodes);

  const filteredNodes = nodeTemplates.filter(node => 
    node.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleDragStart = (event: React.DragEvent, node: NodeTemplate) => {
    event.dataTransfer.setData('application/json', JSON.stringify({
      type: node.type,
      name: node.name,
      description: node.description,
      defaults: node.defaults
    }));
    event.dataTransfer.effectAllowed = 'move';

    // Create a drag preview
    const preview = document.createElement('div');
    const isVSCode = theme === 'vscode';
    preview.className = `
      w-16 h-16 rounded-lg shadow-lg flex items-center justify-center border-2
      ${isVSCode ? 'bg-[#1e1e1e] border-[#454545]' : 'bg-white border-gray-200'}
    `;
    
    // Create icon element
    const iconDiv = document.createElement('div');
    iconDiv.className = 'w-8 h-8 flex items-center justify-center';
    const iconColor = node.type === 'add' 
      ? (isVSCode ? '#9B51E0' : '#9B51E0') 
      : (isVSCode ? '#F2994A' : '#F2994A');
    
    iconDiv.innerHTML = node.type === 'add' 
      ? `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="${iconColor}" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>`
      : `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="${iconColor}" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>`;
    
    preview.appendChild(iconDiv);
    document.body.appendChild(preview);
    event.dataTransfer.setDragImage(preview, 32, 32);
    
    // Remove the preview element after a short delay
    setTimeout(() => document.body.removeChild(preview), 100);
  };

  const handleNodeClick = (node: NodeTemplate) => {
    // Find the source node in our nodes array
    const sourceNode = nodes.find(n => n.id === sourceNodeId);
    
    // Calculate position relative to source node
    let position = { x: 100, y: 100 };
    if (sourceNode) {
      position = {
        x: sourceNode.position.x + 200, // Place 200 units to the right
        y: sourceNode.position.y, // Same vertical position
      };
    }

    // Create and add the new node
    const newNode = {
      id: `${node.type}-${Date.now()}`,
      type: 'custom',
      position,
      data: {
        type: node.type,
        name: node.name,
        description: node.description,
        inputs: node.defaults?.inputs || { a: 0, b: 0 },
        outputs: node.defaults?.outputs || { result: 0 },
      },
    };

    addNode(newNode);
  };

  return (
    <div className={`
      flex flex-col border-r
      ${isCollapsed ? 'w-16' : 'w-72'}
      transition-all duration-300 ease-in-out
      ${theme === 'vscode' ? 
        'bg-node-vscode-bg border-node-vscode-border' : 
        'bg-node-miro-bg border-node-miro-border'}
    `}>
      {/* Search and Toggle */}
      <div className={`p-4 border-b ${theme === 'vscode' ? 'border-node-vscode-border' : 'border-node-miro-border'}`}>
        {!isCollapsed && (
          <div className="mb-3">
            <SearchInput value={searchTerm} onChange={setSearchTerm} />
          </div>
        )}
        <button
          onClick={() => setSidebarExpanded(isCollapsed)}
          className={`
            w-full flex items-center justify-center p-2 rounded-md 
            transition-colors duration-200
            ${theme === 'vscode' ? 
              'bg-node-vscode-button hover:bg-node-vscode-button-hover text-node-vscode-text' : 
              'bg-node-miro-button hover:bg-node-miro-button-hover text-node-miro-text'}
          `}
        >
          {isCollapsed ? (
            <ChevronRightIcon className="w-5 h-5" />
          ) : (
            <ChevronLeftIcon className="w-5 h-5" />
          )}
        </button>
      </div>

      {/* Node List */}
      {isCollapsed ? (
        <div className="flex-1 p-2 space-y-2">
          {nodeTemplates.map((node) => (
            <div
              key={node.type}
              draggable
              onDragStart={(e) => handleDragStart(e, node)}
              onClick={() => handleNodeClick(node)}
              className={`
                w-12 h-12 rounded-lg cursor-grab active:cursor-grabbing
                border transition-all duration-200
                hover:shadow-lg group
                ${theme === 'vscode' ? 
                  'bg-node-vscode-bg border-node-vscode-border hover:border-node-vscode-selected' : 
                  'bg-node-miro-bg border-node-miro-border hover:border-node-miro-selected'}
                ${node.type === 'add' ? 
                  (theme === 'vscode' ? 'text-node-vscode-add' : 'text-node-miro-add') : 
                  (theme === 'vscode' ? 'text-node-vscode-multiply' : 'text-node-miro-multiply')}
              `}
              title={node.name}
            >
              <div className="w-full h-full flex items-center justify-center group-hover:scale-110 transition-transform duration-200">
                {node.icon}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto p-4">
          <div className="space-y-2">
            {filteredNodes.map((node) => (
              <div
                key={node.type}
                draggable
                onDragStart={(e) => handleDragStart(e, node)}
                onClick={() => handleNodeClick(node)}
                className={`
                  flex items-center gap-3 p-3 rounded-lg cursor-grab active:cursor-grabbing
                  border transition-all duration-200
                  hover:shadow-lg group
                  ${theme === 'vscode' ? 
                    'bg-node-vscode-bg border-node-vscode-border hover:border-node-vscode-selected' : 
                    'bg-node-miro-bg border-node-miro-border hover:border-node-miro-selected'}
                  ${sourceNodeId ? 'hover:ring-2 hover:ring-green-500' : ''}
                `}
              >
                <div className={`
                  w-10 h-10 rounded-lg flex items-center justify-center
                  group-hover:scale-110 transition-transform duration-200
                  ${node.type === 'add' ? 
                    (theme === 'vscode' ? 'text-node-vscode-add' : 'text-node-miro-add') : 
                    (theme === 'vscode' ? 'text-node-vscode-multiply' : 'text-node-miro-multiply')}
                `}>
                  {node.icon}
                </div>
                <div>
                  <h3 className={`text-sm font-medium mb-0.5
                    ${theme === 'vscode' ? 'text-node-vscode-text' : 'text-node-miro-text'}`}>
                    {node.name}
                  </h3>
                  <p className={`text-xs
                    ${theme === 'vscode' ? 'text-node-vscode-text opacity-60' : 'text-node-miro-text opacity-60'}`}>
                    {node.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}; 