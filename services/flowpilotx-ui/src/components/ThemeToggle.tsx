import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { ComputerDesktopIcon, PaintBrushIcon } from '@heroicons/react/24/outline';

export const ThemeToggle: React.FC = () => {
  const { theme, setTheme } = useWorkflowStore();

  return (
    <div className="fixed bottom-4 right-4 flex items-center gap-2 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-2 border border-gray-200 dark:border-gray-700">
      <button
        onClick={() => setTheme('vscode')}
        className={`p-2 rounded-md transition-colors ${
          theme === 'vscode'
            ? 'bg-node-vscode-selected text-white'
            : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400'
        }`}
        title="VS Code Theme"
      >
        <ComputerDesktopIcon className="w-5 h-5" />
      </button>
      <button
        onClick={() => setTheme('miro')}
        className={`p-2 rounded-md transition-colors ${
          theme === 'miro'
            ? 'bg-node-miro-selected text-white'
            : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400'
        }`}
        title="Miro Theme"
      >
        <PaintBrushIcon className="w-5 h-5" />
      </button>
    </div>
  );
}; 