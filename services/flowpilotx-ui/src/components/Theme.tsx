import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { SunIcon, MoonIcon } from '@heroicons/react/24/outline';

export const ThemeToggleX: React.FC = () => {
  const { theme, setTheme } = useWorkflowStore();

  return (
    <div className={`
      fixed top-4 right-4 z-50
      flex items-center gap-2 
      rounded-lg shadow-lg p-1.5 
      ${theme === 'vscode' ? 
        'bg-node-vscode-bg border border-node-vscode-border' : 
        'bg-node-miro-bg border border-node-miro-border'}
    `}>
      {/* VS Code Theme Button */}
      <button
        onClick={() => setTheme('vscode')}
        className={`
          p-2 rounded-md transition-all duration-200
          flex items-center gap-2
          ${theme === 'vscode'
            ? 'bg-node-vscode-selected text-white ring-2 ring-node-vscode-selected ring-opacity-50'
            : 'bg-node-vscode-bg border border-node-vscode-border text-node-vscode-text hover:bg-node-vscode-button'}
        `}
        title="VS Code Theme"
      >
        <MoonIcon className="w-4 h-4" />
        {theme === 'vscode' && <span className="text-xs font-medium">Dark</span>}
      </button>

      {/* Miro Theme Button */}
      <button
        onClick={() => setTheme('miro')}
        className={`
          p-2 rounded-md transition-all duration-200
          flex items-center gap-2
          ${theme === 'miro'
            ? 'bg-node-miro-selected text-white ring-2 ring-node-miro-selected ring-opacity-50'
            : 'bg-node-miro-bg border border-node-miro-border text-node-miro-text hover:bg-node-miro-button'}
        `}
        title="Miro Theme"
      >
        <SunIcon className="w-4 h-4" />
        {theme === 'miro' && <span className="text-xs font-medium">Light</span>}
      </button>
    </div>
  );
}; 