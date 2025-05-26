import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const theme = useWorkflowStore((state) => state.theme);

  return (
    <div className={`min-h-screen ${theme === 'vscode' ? 'bg-node-vscode-bg' : 'bg-node-miro-bg'}`}>
      {children}
    </div>
  );
}; 