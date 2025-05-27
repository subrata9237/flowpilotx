import React, { useState, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  PlusIcon,
  ClockIcon,
  KeyIcon,
  WrenchIcon,
  DocumentTextIcon,
  FolderIcon,
  PencilIcon,
  TrashIcon,
  PlayIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';
import { useWorkflowStore } from '../store/workflowStore';
import { nanoid } from 'nanoid';
import { Node, Edge } from 'reactflow';

type Tab = 'workflows' | 'credentials' | 'env' | 'secrets' | 'history';
type WorkflowView = 'actions' | 'list' | 'recent';

interface TabItem {
  name: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  action?: () => void;
  link?: string;
}

interface Section {
  title: string;
  description: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  color: string;
  items: TabItem[];
}

interface Sections {
  [key: string]: Section;
}

export const Home: React.FC = () => {
  const theme = useWorkflowStore((state) => state.theme);
  const isVSCode = theme === 'vscode';
  const navigate = useNavigate();
  const setCurrentProjectId = useWorkflowStore((state) => state.setCurrentProjectId);
  const deleteWorkflow = useWorkflowStore((state) => state.deleteWorkflow);
  const workflows = useWorkflowStore((state) => state.workflows);
  const [activeTab, setActiveTab] = useState<Tab>('workflows');
  const [workflowView, setWorkflowView] = useState<WorkflowView>('actions');
  const [searchQuery, setSearchQuery] = useState('');
  const [workflowToDelete, setWorkflowToDelete] = useState<string | null>(null);

  // Filter workflows based on search query and view type
  const filteredWorkflows = useMemo(() => {
    let filtered = workflows;
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(workflow => 
        workflow.name.toLowerCase().includes(query) ||
        (workflow.id && workflow.id.toLowerCase().includes(query))
      );
    }
    if (workflowView === 'recent') {
      filtered = [...filtered]
        .sort((a, b) => b.updatedAt - a.updatedAt)
        .slice(0, 5);
    }
    return filtered;
  }, [workflows, searchQuery, workflowView]);

  // Create and save a new workflow
  const handleCreateWorkflow = () => {
    const newProjectId = nanoid(8);
    setCurrentProjectId(newProjectId);
    navigate(`/workflow/new?projectid=${newProjectId}`);
  };

  const handleViewAllWorkflows = () => {
    setWorkflowView('list');
  };

  const handleViewRecentWorkflows = () => {
    setWorkflowView('recent');
  };

  const handleBackToActions = () => {
    setWorkflowView('actions');
  };

  const handleWorkflowClick = (workflowId: string) => {
    navigate(`/workflow/${workflowId}`);
  };

  // Handle workflow deletion
  const handleDeleteWorkflow = (workflowId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setWorkflowToDelete(workflowId);
  };

  // Confirm workflow deletion
  const confirmDeleteWorkflow = () => {
    if (workflowToDelete) {
      deleteWorkflow(workflowToDelete);
      setWorkflowToDelete(null);
    }
  };

  // Cancel workflow deletion
  const cancelDeleteWorkflow = () => {
    setWorkflowToDelete(null);
  };

  const tabs = [
    { id: 'workflows', name: 'Workflows', icon: DocumentTextIcon },
    { id: 'credentials', name: 'Credentials', icon: KeyIcon },
    { id: 'env', name: 'Environment Variables', icon: WrenchIcon },
    { id: 'secrets', name: 'Secrets', icon: KeyIcon },
    { id: 'history', name: 'History', icon: ClockIcon },
  ];

  const sections: Sections = {
    workflows: {
      title: 'Workflows',
      description: 'Create and manage your automation workflows',
      icon: DocumentTextIcon,
      color: 'bg-blue-500',
      items: [
        { name: 'Create New Workflow', icon: PlusIcon, action: handleCreateWorkflow },
        { name: 'Recent Workflows', icon: ClockIcon, action: handleViewRecentWorkflows },
        { name: 'All Workflows', icon: FolderIcon, action: handleViewAllWorkflows },
      ],
    },
    credentials: {
      title: 'Credentials',
      description: 'Manage your API keys and authentication tokens',
      icon: KeyIcon,
      color: 'bg-green-500',
      items: [
        { name: 'Add New Credential', icon: PlusIcon, link: '/credentials/new' },
        { name: 'View All Credentials', icon: FolderIcon, link: '/credentials' },
      ],
    },
    env: {
      title: 'Environment Variables',
      description: 'Configure environment-specific settings',
      icon: WrenchIcon,
      color: 'bg-purple-500',
      items: [
        { name: 'Add New Variable', icon: PlusIcon, link: '/env/new' },
        { name: 'View All Variables', icon: FolderIcon, link: '/env' },
      ],
    },
    secrets: {
      title: 'Secrets',
      description: 'Manage sensitive configuration data',
      icon: KeyIcon,
      color: 'bg-red-500',
      items: [
        { name: 'Add New Secret', icon: PlusIcon, link: '/secrets/new' },
        { name: 'View All Secrets', icon: FolderIcon, link: '/secrets' },
      ],
    },
    history: {
      title: 'Workflow History',
      description: 'View execution history and logs',
      icon: ClockIcon,
      color: 'bg-yellow-500',
      items: [
        { name: 'Recent Executions', icon: ClockIcon, link: '/history/recent' },
        { name: 'View All History', icon: FolderIcon, link: '/history' },
      ],
    },
  };

  const IconComponent = sections[activeTab].icon;

  const renderWorkflowList = () => (
    <div className="space-y-4">
      <div className="flex justify-between items-center mb-4">
        <h3 className={`text-lg font-medium ${isVSCode ? 'text-white' : 'text-gray-900'}`}>
          {workflowView === 'recent' ? 'Recent Workflows' : 'All Workflows'}
        </h3>
        <button
          onClick={handleBackToActions}
          className={`px-4 py-2 rounded-lg text-sm font-medium
            ${isVSCode 
              ? 'bg-[#2d2d2d] hover:bg-[#3d3d3d] text-white' 
              : 'bg-gray-100 hover:bg-gray-200 text-gray-900'}`}
        >
          Back to Actions
        </button>
      </div>
      {/* Search Bar */}
      <div className="relative mb-6">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <MagnifyingGlassIcon className={`h-5 w-5 ${isVSCode ? 'text-gray-400' : 'text-gray-500'}`} />
        </div>
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder={`Search ${workflowView === 'recent' ? 'recent' : 'all'} workflows...`}
          className={`block w-full pl-10 pr-3 py-2 border rounded-lg text-sm
            ${isVSCode
              ? 'bg-[#2d2d2d] border-[#454545] text-white placeholder-gray-400 focus:border-blue-500'
              : 'bg-white border-gray-300 text-gray-900 placeholder-gray-500 focus:border-blue-500'
            } focus:outline-none focus:ring-1 focus:ring-blue-500`}
        />
      </div>
      {/* Workflow List */}
      <div className="space-y-3">
        {filteredWorkflows.length === 0 ? (
          <div className={`text-center py-8 ${isVSCode ? 'text-gray-400' : 'text-gray-500'}`}>
            No workflows found matching your search.
          </div>
        ) : (
          filteredWorkflows.map((workflow) => (
            <div
              key={workflow.id}
              className={`p-4 rounded-lg border transition-all duration-200 cursor-pointer
                ${isVSCode 
                  ? 'bg-[#1e1e1e] border-[#454545] hover:border-[#666666]' 
                  : 'bg-white border-gray-200 hover:border-gray-300'}`}
              onClick={() => handleWorkflowClick(workflow.id)}
            >
              <div className="flex justify-between items-start">
                <div>
                  <h4 className={`text-lg font-medium mb-1 ${isVSCode ? 'text-white' : 'text-gray-900'}`}>
                    {workflow.name}
                  </h4>
                  <div className="flex items-center space-x-2 mb-1">
                    <span className={`text-xs font-mono ${isVSCode ? 'text-gray-400' : 'text-gray-500'}`}>ID: {workflow.id}</span>
                  </div>
                  {(() => {
                    let desc: React.ReactNode = 'No description';
                    if ('description' in workflow && typeof workflow.description === 'string' && workflow.description) {
                      desc = workflow.description;
                    }
                    return <p className={`text-sm ${isVSCode ? 'text-gray-400' : 'text-gray-600'}`}>{desc}</p>;
                  })()}
                  <div className="mt-2 flex items-center space-x-4">
                    <span className={`text-xs ${isVSCode ? 'text-gray-400' : 'text-gray-500'}`}>
                      Last modified: {workflow.updatedAt ? new Date(workflow.updatedAt).toLocaleString() : 'N/A'}
                    </span>
                  </div>
                </div>
                <div className="flex space-x-2">
                  <button
                    className={`p-2 rounded-lg transition-colors pointer-events-auto"
                      ${isVSCode
                        ? 'hover:bg-[#2d2d2d] text-gray-400 hover:text-white'
                        : 'hover:bg-gray-100 text-gray-600 hover:text-gray-900'}`}
                    onClick={e => { e.stopPropagation(); /* Play logic here */ }}
                  >
                    <PlayIcon className="w-5 h-5" />
                  </button>
                  <button
                    className={`p-2 rounded-lg transition-colors pointer-events-auto"
                      ${isVSCode
                        ? 'hover:bg-[#2d2d2d] text-gray-400 hover:text-white'
                        : 'hover:bg-gray-100 text-gray-600 hover:text-gray-900'}`}
                    onClick={e => { e.stopPropagation(); navigate(`/workflow/${workflow.id}`); }}
                  >
                    <PencilIcon className="w-5 h-5" />
                  </button>
                  <button
                    className={`p-2 rounded-lg transition-colors pointer-events-auto"
                      ${isVSCode
                        ? 'hover:bg-[#2d2d2d] text-gray-400 hover:text-white'
                        : 'hover:bg-gray-100 text-gray-600 hover:text-gray-900'}`}
                    onClick={e => handleDeleteWorkflow(workflow.id, e)}
                  >
                    <TrashIcon className="w-5 h-5" />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );

  return (
    <div className={`min-h-screen ${isVSCode ? 'bg-node-vscode-bg' : 'bg-node-miro-bg'}`}>
      {/* Delete Confirmation Modal */}
      {workflowToDelete && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className={`p-6 rounded-lg shadow-xl max-w-md w-full mx-4
            ${isVSCode ? 'bg-[#1e1e1e] text-white' : 'bg-white text-gray-900'}`}>
            <h3 className="text-lg font-semibold mb-4">Delete Workflow</h3>
            <p className={`mb-6 ${isVSCode ? 'text-gray-300' : 'text-gray-600'}`}>
              Are you sure you want to delete this workflow? This action cannot be undone.
            </p>
            <div className="flex justify-end space-x-4">
              <button
                onClick={cancelDeleteWorkflow}
                className={`px-4 py-2 rounded-lg text-sm font-medium
                  ${isVSCode 
                    ? 'bg-[#2d2d2d] hover:bg-[#3d3d3d] text-white' 
                    : 'bg-gray-100 hover:bg-gray-200 text-gray-900'}`}
              >
                Cancel
              </button>
              <button
                onClick={confirmDeleteWorkflow}
                className="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 hover:bg-red-700 text-white"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="text-center mb-8">
          <h1 className={`text-4xl font-bold ${isVSCode ? 'text-white' : 'text-gray-900'} mb-4`}>
            Welcome to FlowPilotX
          </h1>
          <p className={`text-lg ${isVSCode ? 'text-gray-300' : 'text-gray-600'}`}>
            Your automation workflow platform
          </p>
        </div>
        {/* Tabs */}
        <div className="mb-8">
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex space-x-8" aria-label="Tabs">
              {tabs.map((tab) => {
                const TabIcon = tab.icon;
                return (
                  <button
                    key={tab.id}
                    onClick={() => {
                      setActiveTab(tab.id as Tab);
                      setWorkflowView('actions');
                    }}
                    className={`
                      group inline-flex items-center py-4 px-1 border-b-2 font-medium text-sm
                      ${activeTab === tab.id
                        ? isVSCode
                          ? 'border-blue-500 text-blue-400'
                          : 'border-blue-500 text-blue-600'
                        : isVSCode
                          ? 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-300'
                          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                      }
                    `}
                  >
                    <TabIcon
                      className={`
                        -ml-0.5 mr-2 h-5 w-5
                        ${activeTab === tab.id
                          ? isVSCode
                            ? 'text-blue-400'
                            : 'text-blue-500'
                          : isVSCode
                            ? 'text-gray-400 group-hover:text-gray-300'
                            : 'text-gray-400 group-hover:text-gray-500'
                        }
                      `}
                    />
                    {tab.name}
                  </button>
                );
              })}
            </nav>
          </div>
        </div>
        {/* Active Section Content */}
        <div className={`rounded-lg shadow-lg overflow-hidden
          ${isVSCode ? 'bg-[#1e1e1e] border border-[#454545]' : 'bg-white border border-gray-200'}`}>
          <div className="p-6">
            {activeTab === 'workflows' && (workflowView === 'list' || workflowView === 'recent') ? (
              renderWorkflowList()
            ) : (
              <>
                <div className="flex items-center mb-4">
                  <div className={`p-3 rounded-lg ${sections[activeTab].color} bg-opacity-10`}>
                    <IconComponent className={`w-6 h-6 ${sections[activeTab].color.replace('bg-', 'text-')}`} />
                  </div>
                  <h2 className={`ml-4 text-xl font-semibold ${isVSCode ? 'text-white' : 'text-gray-900'}`}>
                    {sections[activeTab].title}
                  </h2>
                </div>
                <p className={`mb-6 ${isVSCode ? 'text-gray-400' : 'text-gray-600'}`}>
                  {sections[activeTab].description}
                </p>
                <div className="space-y-2">
                  {sections[activeTab].items.map((item) => {
                    const ItemIcon = item.icon;
                    return item.action ? (
                      <button
                        key={item.name}
                        onClick={item.action}
                        className={`flex items-center p-3 rounded-lg transition-all duration-200 w-full
                          ${isVSCode 
                            ? 'hover:bg-[#2d2d2d] text-gray-300 hover:text-white' 
                            : 'hover:bg-gray-50 text-gray-600 hover:text-gray-900'}`}
                      >
                        <ItemIcon className="w-5 h-5 mr-3" />
                        <span>{item.name}</span>
                      </button>
                    ) : (
                      <Link
                        key={item.name}
                        to={item.link || '#'}
                        className={`flex items-center p-3 rounded-lg transition-all duration-200
                          ${isVSCode 
                            ? 'hover:bg-[#2d2d2d] text-gray-300 hover:text-white' 
                            : 'hover:bg-gray-50 text-gray-600 hover:text-gray-900'}`}
                      >
                        <ItemIcon className="w-5 h-5 mr-3" />
                        <span>{item.name}</span>
                      </Link>
                    );
                  })}
                </div>
              </>
            )}
          </div>
        </div>
        {/* Quick Actions */}
        <div className={`mt-8 p-6 rounded-lg shadow-lg
          ${isVSCode ? 'bg-[#1e1e1e] border border-[#454545]' : 'bg-white border border-gray-200'}`}>
          <h2 className={`text-xl font-semibold mb-4 ${isVSCode ? 'text-white' : 'text-gray-900'}`}>
            Quick Actions
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button
              onClick={handleCreateWorkflow}
              className={`flex items-center justify-center p-4 rounded-lg transition-all duration-200 w-full
                ${isVSCode 
                  ? 'bg-blue-600 hover:bg-blue-700 text-white' 
                  : 'bg-blue-500 hover:bg-blue-600 text-white'}`}
            >
              <PlusIcon className="w-5 h-5 mr-2" />
              Create New Workflow
            </button>
            <button
              onClick={handleViewAllWorkflows}
              className={`flex items-center justify-center p-4 rounded-lg transition-all duration-200
                ${isVSCode 
                  ? 'bg-[#2d2d2d] hover:bg-[#3d3d3d] text-white' 
                  : 'bg-gray-100 hover:bg-gray-200 text-gray-900'}`}
            >
              <FolderIcon className="w-5 h-5 mr-2" />
              View All Workflows
            </button>
            <button
              onClick={handleViewRecentWorkflows}
              className={`flex items-center justify-center p-4 rounded-lg transition-all duration-200
                ${isVSCode 
                  ? 'bg-[#2d2d2d] hover:bg-[#3d3d3d] text-white' 
                  : 'bg-gray-100 hover:bg-gray-200 text-gray-900'}`}
            >
              <ClockIcon className="w-5 h-5 mr-2" />
              View Recent Workflows
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}; 