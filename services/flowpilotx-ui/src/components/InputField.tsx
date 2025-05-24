import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';

interface InputFieldProps {
  value: string | number;
  onChange: (value: any) => void;
  placeholder?: string;
  type?: string;
  className?: string;
}

export const InputField: React.FC<InputFieldProps> = ({ 
  value, 
  onChange, 
  placeholder, 
  type = 'text', 
  className = '' 
}) => {
  const theme = useWorkflowStore((state) => state.theme);
  
  return (
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className={`
        w-full px-3 py-2 rounded-md text-sm
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
        ${className}
      `}
    />
  );
}; 