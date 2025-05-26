import React from 'react';
import { useDrag } from 'react-dnd';
import { NodeTemplate } from '../types/workflow';

interface DraggableNodeProps {
  node: NodeTemplate;
}

export const DraggableNode: React.FC<DraggableNodeProps> = ({ node }) => {
  const [{ isDragging }, dragRef] = useDrag(() => ({
    type: 'node',
    item: node,
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
    previewOptions: {
      captureDraggingState: true,
    },
  }));

  return (
    <div 
      ref={dragRef}
      className={`px-2 ${isDragging ? 'opacity-50' : 'opacity-100'}`}
      style={{ touchAction: 'none' }}
    >
      <div 
        className="group p-3 rounded-lg bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 
          border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-all 
          cursor-grab active:cursor-grabbing"
      >
        <div className="flex items-center gap-3">
          <div 
            className="w-8 h-8 flex items-center justify-center rounded-md"
            style={{ 
              backgroundColor: `${node.color}15`,
              color: node.color,
            }}
          >
            {React.createElement(node.icon, { className: "w-5 h-5" })}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
              {node.name}
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
              {node.description}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}; 