import React from 'react';
import { useWorkflowStore } from '../store/workflowStore';
import { HandRaisedIcon, CursorArrowRaysIcon, RectangleStackIcon, PencilSquareIcon, EyeIcon, EyeSlashIcon, HomeIcon, ArrowDownTrayIcon } from '@heroicons/react/24/outline';
import { ThemeToggleX } from './Theme';
import { useNavigate } from 'react-router-dom';

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

interface ToolbarHeaderProps {
  handleSave?: () => void;
  workflowName: string;
  setWorkflowName: (name: string) => void;
  editingName: boolean;
  setEditingName: (editing: boolean) => void;
  onHomeClick?: () => void;
  hasUnsavedChanges?: boolean;
}

export const ToolbarHeader: React.FC<ToolbarHeaderProps> = ({
  handleSave,
  workflowName,
  setWorkflowName,
  editingName,
  setEditingName,
  onHomeClick,
  hasUnsavedChanges = false,
}) => {
  const editorMode = useWorkflowStore((state) => state.editorMode);
  const setEditorMode = useWorkflowStore((state) => state.setEditorMode);
  const theme = useWorkflowStore((state) => state.theme);
  const addNode = useWorkflowStore((state) => state.addNode);
  const showStickyNotes = useWorkflowStore((state) => state.showStickyNotes);
  const toggleStickyNotes = useWorkflowStore((state) => state.toggleStickyNotes);
  const iconColor = theme === 'vscode' ? '#9B51E0' : '#F2994A';
  const navigate = useNavigate();

  const handleAddStickyNote = () => {
    const newNode = {
      id: `sticky-${Date.now()}`,
      type: 'stickyNote',
      position: { x: 100, y: 100 },
      width: 180, // default width
      height: 100, // default height
      resizable: true, // enable NodeResizer
      data: {
        name: 'Sticky Note',
        type: 'sticky',
        text: '',
        color: 'rgba(253, 230, 138, 0.45)', // Default yellow color with opacity
        inputs: {},
        outputs: {},
      }
    };
    addNode(newNode);
  };

  return (
    <header
      className={`relative w-full h-16 border-b flex items-center justify-center px-2 sm:px-4 ${theme === 'vscode'
        ? 'bg-node-vscode-bg/50 border-node-vscode-border' 
        : 'bg-node-miro-bg/50 border-node-miro-border'
        } backdrop-blur-sm`}
    >
      {/* Far left: Logo Icon and Title */}
      <div className="flex items-center h-full pr-4">
        <AutomatedFlowIcon color={iconColor} />
        <h1
          className={`text-xl font-bold truncate max-w-[12rem] sm:max-w-xs ml-2 ${theme === 'vscode' ? 'text-node-vscode-text' : 'text-node-miro-text'
            }`}
          title="FlowPilotX"
        >
            FlowPilotX
          </h1>

      </div>
      <div className="ml-5 flex items-center">
          {editingName ? (
            <input
              type="text"
              value={workflowName}
              onChange={e => setWorkflowName(e.target.value)}
              onBlur={() => setEditingName(false)}
              onKeyDown={e => {
                if (e.key === 'Enter') setEditingName(false);
              }}
              autoFocus
              className="px-2 py-0.5 rounded border border-gray-300 bg-transparent text-white text-xs"
              style={{ minWidth: 80, maxWidth: 160 }}
            />
          ) : (
            <>
              <span className="text-white font-medium text-xs truncate max-w-[8rem]">{workflowName}</span>
              <button
                onClick={() => setEditingName(true)}
                className="ml-1 p-1 rounded hover:bg-gray-700"
                title="Edit workflow name"
              >
                <PencilSquareIcon className="w-3.5 h-3.5 text-white" />
              </button>
            </>
          )}
        </div>
      <div className="w-full max-w-7xl mx-auto grid grid-cols-3 items-center h-full">
        <div></div>
        
        <div className="flex items-center justify-center gap-2">
          <div
            className={`flex items-center gap-2 rounded-lg p-1 shadow-sm ${theme === 'vscode'
              ? 'bg-node-vscode-bg/50' 
              : 'bg-node-miro-bg/50'
              }`}
          >
            {handleSave && (
              <button
                onClick={handleSave}
                className={`flex items-center justify-center p-2 rounded hover:bg-blue-100 dark:hover:bg-blue-900 transition ${
                  hasUnsavedChanges 
                    ? 'text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20' 
                    : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                title={hasUnsavedChanges ? 'Save changes' : 'All changes saved'}
              >
                <ArrowDownTrayIcon className="w-5 h-5 text-blue-600" />
              </button>
            )}
            <button
              onClick={onHomeClick || (() => navigate('/'))}
              className="flex items-center gap-2 text-gray-700 dark:text-gray-200 hover:text-blue-600 dark:hover:text-blue-400 transition"
              title="Go to Home"
            >
              <HomeIcon className="w-5 h-5" />
              <span className="font-semibold hidden sm:inline"></span>
            </button>
            <button
              onClick={() => setEditorMode('click')}
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 ${editorMode === 'click'
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
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 ${editorMode === 'pan'
                  ? 'bg-blue-500 text-white'
                  : theme === 'vscode'
                    ? 'text-node-vscode-text hover:bg-node-vscode-button'
                    : 'text-node-miro-text hover:bg-node-miro-button'
              }`}
              title="Pan Mode - Move nodes and canvas"
            >
              <HandRaisedIcon className="w-5 h-5" />
            </button>
            <button
              onClick={handleAddStickyNote}
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-yellow-400 ${theme === 'vscode'
                  ? 'text-yellow-400 hover:bg-yellow-900/20'
                  : 'text-yellow-600 hover:bg-yellow-100'
                }`}
              title="Create Sticky Note"
            >
              <PencilSquareIcon className="w-5 h-5" />
            </button>
            <button
              onClick={toggleStickyNotes}
              className={`px-3 py-2 rounded-md transition-colors font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-yellow-400 ${theme === 'vscode'
                  ? showStickyNotes
                    ? 'text-yellow-400 hover:bg-yellow-900/20'
                    : 'text-gray-400 hover:bg-gray-900/20'
                  : showStickyNotes
                    ? 'text-yellow-600 hover:bg-yellow-100'
                    : 'text-gray-600 hover:bg-gray-100'
                }`}
              title={showStickyNotes ? "Hide Sticky Notes" : "Show Sticky Notes"}
            >
              {showStickyNotes ? (
                <EyeIcon className="w-5 h-5" />
              ) : (
                <EyeSlashIcon className="w-5 h-5" />
              )}
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