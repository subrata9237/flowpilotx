import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWorkflowStore } from '../store/workflowStore';
import {
  ArrowLeftIcon,
  PlusIcon,
  ChevronDownIcon,
  AdjustmentsHorizontalIcon,
  DocumentDuplicateIcon,
  XMarkIcon,
  PencilIcon,
  CheckCircleIcon,
  CheckIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { InputField } from '../components/InputField';

const TextArea: React.FC<{
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  rows?: number;
  className?: string;
}> = ({ value, onChange, placeholder, rows = 3, className = '' }) => {
  const theme = useWorkflowStore((state) => state.theme);
  
  return (
    <textarea
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      rows={rows}
      className={`
        w-full px-3 py-2 rounded-md text-sm
        transition-all duration-200
        focus:outline-none resize-none
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

const ParameterRow: React.FC<{
  name: string;
  value: any;
  onValueChange: (value: any) => void;
  onNameChange: (name: string) => void;
  onDelete: () => void;
  isEditing: boolean;
  onEditToggle: () => void;
  theme: string;
}> = ({ name, value, onValueChange, onNameChange, onDelete, isEditing, onEditToggle, theme }) => {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        {isEditing ? (
          <InputField
            value={name}
            onChange={onNameChange}
            className="flex-1 mr-2"
            placeholder="Parameter name"
          />
        ) : (
          <span className={`text-sm font-medium ${theme === 'vscode' ? 
            'text-node-vscode-text' : 'text-node-miro-text'}`}>
            {name}
          </span>
        )}
        <div className="flex items-center gap-1">
          <button
            onClick={onEditToggle}
            className={`p-1.5 rounded-md transition-colors ${theme === 'vscode' ? 
              'hover:bg-node-vscode-button text-node-vscode-text' : 
              'hover:bg-node-miro-button text-node-miro-text'}`}
            title={isEditing ? "Save" : "Edit"}
          >
            {isEditing ? (
              <CheckIcon className="w-4 h-4" />
            ) : (
              <PencilIcon className="w-4 h-4" />
            )}
          </button>
          <button
            onClick={onDelete}
            className={`p-1.5 rounded-md transition-colors ${theme === 'vscode' ? 
              'hover:bg-red-900/30 text-red-400 hover:text-red-300' : 
              'hover:bg-red-50 text-red-500 hover:text-red-600'}`}
            title="Delete"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
      <div className="flex gap-2">
        <InputField
          type="number"
          value={value}
          onChange={onValueChange}
          className="flex-1"
        />
        <button
          className={`p-2 rounded-md transition-colors ${theme === 'vscode' ? 
            'bg-node-vscode-button hover:bg-node-vscode-button-hover text-node-vscode-text' : 
            'bg-node-miro-button hover:bg-node-miro-button-hover text-node-miro-text'}`}
        >
          <ChevronDownIcon className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
};

const ParameterSection: React.FC<{
  title: string;
  parameters: Record<string, any>;
  onParameterChange: (name: string, value: any) => void;
  onParameterNameChange: (oldName: string, newName: string) => void;
  onParameterDelete: (name: string) => void;
  onParameterAdd: () => void;
  theme: string;
}> = ({ title, parameters, onParameterChange, onParameterNameChange, onParameterDelete, onParameterAdd, theme }) => {
  const [editingParam, setEditingParam] = useState<string | null>(null);

  return (
    <div className={`rounded-lg shadow-lg border ${theme === 'vscode' ? 
      'bg-node-vscode-bg border-node-vscode-border' : 
      'bg-node-miro-bg border-node-miro-border'}`}>
      <div className={`p-4 border-b flex justify-between items-center ${theme === 'vscode' ? 
        'border-node-vscode-border' : 'border-node-miro-border'}`}>
        <h2 className={`text-lg font-medium ${theme === 'vscode' ? 
          'text-node-vscode-text' : 'text-node-miro-text'}`}>
          {title}
        </h2>
        <button
          onClick={onParameterAdd}
          className={`p-1.5 rounded-md transition-colors ${theme === 'vscode' ? 
            'hover:bg-node-vscode-button text-node-vscode-text' : 
            'hover:bg-node-miro-button text-node-miro-text'}`}
          title="Add Parameter"
        >
          <PlusIcon className="w-5 h-5" />
        </button>
      </div>
      <div className="p-4 space-y-4">
        {Object.entries(parameters).map(([name, value]) => (
          <ParameterRow
            key={name}
            name={name}
            value={value}
            onValueChange={(newValue) => onParameterChange(name, newValue)}
            onNameChange={(newName) => {
              onParameterNameChange(name, newName);
              setEditingParam(null);
            }}
            onDelete={() => onParameterDelete(name)}
            isEditing={editingParam === name}
            onEditToggle={() => setEditingParam(editingParam === name ? null : name)}
            theme={theme}
          />
        ))}
      </div>
    </div>
  );
};

const NodeSettingsPage: React.FC = () => {
  const { nodeId } = useParams<{ nodeId: string }>();
  const navigate = useNavigate();
  const nodes = useWorkflowStore((state) => state.nodes);
  const updateNodeData = useWorkflowStore((state) => state.updateNodeData);
  const theme = useWorkflowStore((state) => state.theme);

  const node = nodes.find((n) => n.id === nodeId);
  const [nodeData, setNodeData] = useState(node?.data);
  const [isEditingName, setIsEditingName] = useState(false);
  const [newParameterName, setNewParameterName] = useState('');
  const [showAddParameter, setShowAddParameter] = useState(false);
  const [showSuccessToast, setShowSuccessToast] = useState(false);

  useEffect(() => {
    if (!node) {
      navigate('/');
    }
  }, [node, navigate]);

  if (!node || !nodeData) {
    return null;
  }

  const handleSave = () => {
    if (nodeData) {
      updateNodeData(nodeId!, nodeData);
      // Show success toast
      setShowSuccessToast(true);
      // Hide toast after 3 seconds
      setTimeout(() => {
        setShowSuccessToast(false);
        navigate('/');
      }, 2000);
    }
  };

  const handleInputChange = (key: string, value: any) => {
    setNodeData((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        inputs: {
          ...prev.inputs,
          [key]: Number(value),
        },
      };
    });
  };

  const handleNameChange = (newName: string) => {
    setNodeData((prev) => prev ? { ...prev, name: newName } : prev);
  };

  const handleAddParameter = () => {
    if (!newParameterName.trim()) return;
    
    setNodeData((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        inputs: {
          ...prev.inputs,
          [newParameterName.trim()]: 0,
        },
      };
    });
    setNewParameterName('');
    setShowAddParameter(false);
  };

  const handleRemoveParameter = (paramName: string) => {
    setNodeData((prev) => {
      if (!prev) return prev;
      const newInputs = { ...prev.inputs };
      delete newInputs[paramName];
      return {
        ...prev,
        inputs: newInputs,
      };
    });
  };

  const handleInputNameChange = (oldName: string, newName: string) => {
    if (!newName.trim() || oldName === newName) return;
    
    setNodeData((prev) => {
      if (!prev) return prev;
      const newInputs = { ...prev.inputs };
      const value = newInputs[oldName];
      delete newInputs[oldName];
      newInputs[newName] = value;
      return { ...prev, inputs: newInputs };
    });
  };

  const handleOutputNameChange = (oldName: string, newName: string) => {
    if (!newName.trim() || oldName === newName) return;
    
    setNodeData((prev) => {
      if (!prev) return prev;
      const newOutputs = { ...prev.outputs };
      const value = newOutputs[oldName];
      delete newOutputs[oldName];
      newOutputs[newName] = value;
      return { ...prev, outputs: newOutputs };
    });
  };

  const handleAddInput = () => {
    setNodeData((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        inputs: {
          ...prev.inputs,
          [`input${Object.keys(prev.inputs).length + 1}`]: 0,
        },
      };
    });
  };

  const handleAddOutput = () => {
    setNodeData((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        outputs: {
          ...prev.outputs,
          [`output${Object.keys(prev.outputs).length + 1}`]: 0,
        },
      };
    });
  };

  return (
    <div className={`${theme === 'vscode' ? 'bg-node-vscode-bg' : 'bg-node-miro-bg'}`}>
      {/* Success Toast */}
      {showSuccessToast && (
        <div className="fixed top-4 right-4 flex items-center gap-2 bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg animate-fade-in">
          <CheckCircleIcon className="w-5 h-5" />
          <span>Changes saved successfully!</span>
        </div>
      )}

      {/* Header */}
      <header className={`shadow border-b ${theme === 'vscode' ? 
        'bg-node-vscode-bg border-node-vscode-border' : 
        'bg-node-miro-bg border-node-miro-border'}`}>
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate('/')}
              className={`p-2 rounded-md transition-colors ${theme === 'vscode' ? 
                'hover:bg-node-vscode-button text-node-vscode-text' : 
                'hover:bg-node-miro-button text-node-miro-text'}`}
            >
              <ArrowLeftIcon className="w-5 h-5" />
            </button>
            <div className="flex items-center gap-2">
              {isEditingName ? (
                <InputField
                  value={nodeData.name}
                  onChange={handleNameChange}
                  className="text-xl font-semibold"
                />
              ) : (
                <div className="flex items-center gap-2">
                  <h1 className={`text-xl font-semibold ${theme === 'vscode' ? 
                    'text-node-vscode-text' : 'text-node-miro-text'}`}>
                    {nodeData.name}
                  </h1>
                  <button
                    onClick={() => setIsEditingName(true)}
                    className={`p-1 transition-colors ${theme === 'vscode' ? 
                      'text-node-vscode-text opacity-60 hover:opacity-100' : 
                      'text-node-miro-text opacity-60 hover:opacity-100'}`}
                  >
                    <PencilIcon className="w-4 h-4" />
                  </button>
                </div>
              )}
            </div>
            <span className={`px-2 py-1 text-xs font-medium rounded ${theme === 'vscode' ? 
              'bg-node-vscode-button text-node-vscode-text' : 
              'bg-node-miro-button text-node-miro-text'}`}>
              {nodeData.type}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button className="p-2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300">
              <DocumentDuplicateIcon className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6">
        <div className="grid grid-cols-2 gap-6">
          {/* Left Column - Inputs */}
          <div className="space-y-6">
            <ParameterSection
              title="Input Parameters"
              parameters={nodeData.inputs}
              onParameterChange={handleInputChange}
              onParameterNameChange={handleInputNameChange}
              onParameterDelete={handleRemoveParameter}
              onParameterAdd={handleAddInput}
              theme={theme}
            />
          </div>

          {/* Right Column - Outputs */}
          <div className="space-y-6">
            <ParameterSection
              title="Output Parameters"
              parameters={nodeData.outputs}
              onParameterChange={(name, value) => {
                setNodeData((prev) => prev ? {
                  ...prev,
                  outputs: { ...prev.outputs, [name]: value }
                } : prev);
              }}
              onParameterNameChange={handleOutputNameChange}
              onParameterDelete={(name) => {
                setNodeData((prev) => {
                  if (!prev) return prev;
                  const newOutputs = { ...prev.outputs };
                  delete newOutputs[name];
                  return { ...prev, outputs: newOutputs };
                });
              }}
              onParameterAdd={handleAddOutput}
              theme={theme}
            />

            {/* Node Settings with Description */}
            <div className={`rounded-lg shadow-lg border ${theme === 'vscode' ? 
              'bg-node-vscode-bg border-node-vscode-border' : 
              'bg-node-miro-bg border-node-miro-border'}`}>
              <div className={`p-4 border-b ${theme === 'vscode' ? 
                'border-node-vscode-border' : 'border-node-miro-border'}`}>
                <h2 className={`text-lg font-medium ${theme === 'vscode' ? 
                  'text-node-vscode-text' : 'text-node-miro-text'}`}>
                  Node Settings
                </h2>
              </div>
              <div className="p-4 space-y-6">
                {/* Node Info */}
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      Node Type
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white capitalize">
                      {nodeData.type}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      Node ID
                    </span>
                    <span className="text-sm font-mono text-gray-900 dark:text-white">
                      {nodeId}
                    </span>
                  </div>
                </div>

                {/* Description */}
                <div className="space-y-2">
                  <label className={`text-sm ${theme === 'vscode' ? 
                    'text-node-vscode-text' : 'text-node-miro-text'}`}>
                    Description
                  </label>
                  <TextArea
                    value={nodeData.description || ''}
                    onChange={(value) => setNodeData((prev) => prev ? { ...prev, description: value } : prev)}
                    placeholder="Add a description of what this node does..."
                  />
                </div>

                {/* Additional Settings can be added here */}
                <div className="space-y-2">
                  <label className="text-sm text-gray-600 dark:text-gray-400">
                    Node Status
                  </label>
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-green-500"></div>
                    <span className="text-sm text-gray-900 dark:text-white">Active</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className={`fixed bottom-0 inset-x-0 border-t shadow-lg ${theme === 'vscode' ? 
        'bg-node-vscode-bg border-node-vscode-border' : 
        'bg-node-miro-bg border-node-miro-border'}`}>
        <div className="container mx-auto px-4 py-3 flex justify-between items-center">
          <div className={`text-sm ${theme === 'vscode' ? 
            'text-node-vscode-text opacity-60' : 
            'text-node-miro-text opacity-60'}`}>
            Last saved: {new Date().toLocaleTimeString()}
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => navigate('/')}
              className={`
                px-4 py-2 text-sm font-medium rounded-md transition-colors
                ${theme === 'vscode' ? 
                  'bg-node-vscode-button hover:bg-node-vscode-button-hover text-node-vscode-text' : 
                  'bg-node-miro-button hover:bg-node-miro-button-hover text-node-miro-text'}
              `}
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              className={`
                px-4 py-2 text-sm font-medium text-white rounded-md transition-colors
                ${theme === 'vscode' ? 
                  'bg-node-vscode-selected hover:opacity-90' : 
                  'bg-node-miro-selected hover:opacity-90'}
              `}
            >
              Save Changes
            </button>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default NodeSettingsPage; 