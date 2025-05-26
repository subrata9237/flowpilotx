import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { HandRaisedIcon, CursorArrowRaysIcon } from '@heroicons/react/24/outline';
import { ThemeToggleX } from './Theme';

const AutomatedFlowIcon: React.FC<{ color: string }> = ({ color }) => (
  <svg
    width="28"
    height="28"
    viewBox="0 0 28 28"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="mr-2 flex-shrink-0"
  >
    <circle cx="7" cy="7" r="3" stroke={color} strokeWidth="2" />
    <circle cx="21" cy="7" r="3" stroke={color} strokeWidth="2" />
    <circle cx="7" cy="21" r="3" stroke={color} strokeWidth="2" />
    <circle cx="21" cy="21" r="3" stroke={color} strokeWidth="2" />
    <path d="M10 7h8" stroke={color} strokeWidth="2" strokeLinecap="round" />
    <path d="M7 10v8" stroke={color} strokeWidth="2" strokeLinecap="round" />
    <path d="M10 21h8" stroke={color} strokeWidth="2" strokeLinecap="round" />
    <path d="M21 10v8" stroke={color} strokeWidth="2" strokeLinecap="round" />
    <path d="M10 10l8 8" stroke={color} strokeWidth="1.5" strokeDasharray="2 2" />
    <path d="M10 18l8-8" stroke={color} strokeWidth="1.5" strokeDasharray="2 2" />
  </svg>
);

export const ToolbarHeader: React.FC = () => {
  const editorMode = useWorkflowStore((state) => state.editorMode);
  const setEditorMode = useWorkflowStore((state) => state.setEditorMode);
  const theme = useWorkflowStore((state) => state.theme);
  const iconColor = theme === 'vscode' ? '#9B51E0' : '#F2994A';

  return (
    <header
      className={`w-full h-16 border-b flex items-center justify-center px-2 sm:px-4 ${
        theme === 'vscode'
          ? 'bg-node-vscode-bg/50 border-node-vscode-border'
          : 'bg-node-miro-bg/50 border-node-miro-border'
      } backdrop-blur-sm`}
    >
      <div className="w-full max-w-7xl mx-auto grid grid-cols-3 items-center h-full">
        {/* Left: App Title */}
        <div className="flex items-center min-w-[140px] sm:min-w-[180px] pl-0">
          <AutomatedFlowIcon color={iconColor} />
          <h1
            className={`text-xl font-bold truncate max-w-[12rem] sm:max-w-xs ${
              theme === 'vscode' ? 'text-node-vscode-text' : 'text-node-miro-text'
            }`}
            title="FlowPilotX"
          >
            FlowPilotX
          </h1>
        </div>
        {/* Center: Mode Toggle */}
        <div className="flex items-center justify-center">
          <div
            className={`flex items-center gap-2 rounded-lg p-1 shadow-sm ${
              theme === 'vscode'
                ? 'bg-node-vscode-bg/50'
                : 'bg-node-miro-bg/50'
            }`}
          >
            <button
              onClick={() => setEditorMode('click')}
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 ${
                editorMode === 'click'
                  ? 'bg-blue-500 text-white'
                  : theme === 'vscode'
                  ? 'text-node-vscode-text hover:bg-node-vscode-button'
                  : 'text-node-miro-text hover:bg-node-miro-button'
              }`}
              title="Click Mode - Interact with nodes"
            >
              <CursorArrowRaysIcon className="w-5 h-5" />
            </button>
            <button
              onClick={() => setEditorMode('pan')}
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 ${
                editorMode === 'pan'
                  ? 'bg-blue-500 text-white'
                  : theme === 'vscode'
                  ? 'text-node-vscode-text hover:bg-node-vscode-button'
                  : 'text-node-miro-text hover:bg-node-miro-button'
              }`}
              title="Pan Mode - Move nodes and canvas"
            >
              <HandRaisedIcon className="w-5 h-5" />
            </button>
          </div>
        </div>
        {/* Right: Theme Toggle */}
        <div className="flex items-center justify-end min-w-[80px] pr-2 sm:pr-4">
          <ThemeToggleX />
        </div>
      </div>
    </header>
  );
}; 