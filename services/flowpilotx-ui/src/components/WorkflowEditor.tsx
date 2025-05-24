import React, { useCallback, useEffect, useState, useRef } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  Connection,
  NodeDragHandler,
  useOnSelectionChange,
  SelectionMode,
  Panel,
  BackgroundVariant,
  ReactFlowInstance,
  useKeyPress,
  OnNodesChange,
  OnEdgesChange,
  Node,
  Edge,
} from 'reactflow';
import { useWorkflowStore } from '../store/workflowStore';
import { BaseNode } from './BaseNode';
import { ThemeToggle } from './ThemeToggle';
import { NodeData } from '../types/workflow';
import 'reactflow/dist/style.css';

const nodeTypes = {
  add: BaseNode,
  multiply: BaseNode,
};

const defaultEdgeOptions = {
  animated: true,
  style: {
    stroke: '#555',
    strokeWidth: 2,
  },
};

export const WorkflowEditor: React.FC = () => {
  const { nodes, edges, addNode, addEdge, updateNodePosition, deleteNode, theme } = useWorkflowStore();
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [showDeleteTooltip, setShowDeleteTooltip] = useState(false);

  // Track delete key press
  const deletePressed = useKeyPress(['Delete', 'Backspace']);

  // Handle select all
  useEffect(() => {
    const handleSelectAll = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && (e.key === 'a' || e.key === 'A')) {
        e.preventDefault();
        setSelectedNodes(nodes.map(node => node.id));
      }
    };

    window.addEventListener('keydown', handleSelectAll);
    return () => window.removeEventListener('keydown', handleSelectAll);
  }, [nodes]);

  useOnSelectionChange({
    onChange: ({ nodes }) => {
      const selected = nodes.map(node => node.id);
      setSelectedNodes(selected);
      setShowDeleteTooltip(selected.length > 0);
    },
  });

  // Handle node deletion with animation
  const handleNodeDeletion = useCallback((nodesToDelete: string[]) => {
    nodesToDelete.forEach(nodeId => {
      const nodeElement = document.querySelector(`[data-id="${nodeId}"]`);
      if (nodeElement) {
        nodeElement.classList.add('scale-95', 'opacity-50');
      }
    });

    setTimeout(() => {
      nodesToDelete.forEach(nodeId => deleteNode(nodeId));
      setShowDeleteTooltip(false);
    }, 150);
  }, [deleteNode]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.key === 'Delete' || event.key === 'Backspace') && selectedNodes.length > 0) {
        handleNodeDeletion(selectedNodes);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodes, handleNodeDeletion]);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback((event: React.DragEvent) => {
    event.preventDefault();

    if (!reactFlowWrapper.current || !reactFlowInstance) return;

    const reactFlowBounds = reactFlowWrapper.current.getBoundingClientRect();
    const position = reactFlowInstance.project({
      x: event.clientX - reactFlowBounds.left,
      y: event.clientY - reactFlowBounds.top,
    });

    try {
      const jsonData = event.dataTransfer.getData('application/json');
      if (!jsonData) return;

      const nodeData = JSON.parse(jsonData);
      const newNode = {
        id: `${nodeData.type}-${Date.now()}`,
        type: nodeData.type,
        position,
        data: {
          name: nodeData.name,
          type: nodeData.type,
          description: nodeData.description,
          inputs: nodeData.defaults?.inputs || { a: 0, b: 0 },
          outputs: nodeData.defaults?.outputs || { result: 0 },
        } as NodeData,
      };

      addNode(newNode);
    } catch (error) {
      console.error('Failed to handle node drop:', error);
    }
  }, [reactFlowInstance, addNode]);

  const onConnect = useCallback(
    (connection: Connection) => {
      const newEdge = {
        id: `e${connection.source}-${connection.target}`,
        source: connection.source || '',
        target: connection.target || '',
        sourceHandle: connection.sourceHandle,
        targetHandle: connection.targetHandle,
        type: 'smoothstep',
      };
      addEdge(newEdge);
    },
    [addEdge]
  );

  const onNodeDragStop: NodeDragHandler = useCallback(
    (event, node) => {
      updateNodePosition(node.id, node.position);
    },
    [updateNodePosition]
  );

  const onNodesChange: OnNodesChange = useCallback((changes) => {
    changes.forEach((change) => {
      if (change.type === 'position' && change.position && change.id) {
        updateNodePosition(change.id, change.position);
      }
    });
  }, [updateNodePosition]);

  const onEdgesChange: OnEdgesChange = useCallback((changes) => {
    changes.forEach((change) => {
      if (change.type === 'add' && change.item) {
        addEdge(change.item as Edge);
      }
    });
  }, [addEdge]);

  return (
    <div 
      ref={reactFlowWrapper}
      className="flex-1 h-full"
      onDrop={onDrop}
      onDragOver={onDragOver}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        onConnect={onConnect}
        onNodeDragStop={onNodeDragStop}
        onInit={setReactFlowInstance}
        className={theme === 'vscode' ? 'bg-node-vscode-bg' : 'bg-node-miro-bg'}
        minZoom={0.1}
        maxZoom={4}
        snapToGrid
        snapGrid={[16, 16]}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={16}
          size={1}
          color={theme === 'vscode' ? '#404040' : '#E6E6E6'}
        />
        <Controls />
        <MiniMap
          nodeColor={theme === 'vscode' ? '#2D2D2D' : '#FFFFFF'}
          maskColor={theme === 'vscode' ? 'rgba(45, 45, 45, 0.8)' : 'rgba(255, 255, 255, 0.8)'}
        />
        <ThemeToggle />
        {showDeleteTooltip && (
          <Panel position="top-right" className="m-2.5">
            <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm
              ${deletePressed ? 'bg-red-600 text-white' : 'bg-gray-800 text-gray-200'}
              transition-colors duration-150`}
            >
              <span className="font-medium">
                {selectedNodes.length} node{selectedNodes.length > 1 ? 's' : ''} selected
              </span>
              <span className={deletePressed ? 'text-white/80' : 'text-gray-400'}>
                Press Delete to remove
              </span>
            </div>
          </Panel>
        )}
        {nodes.length === 0 && (
          <Panel position="bottom-center" className="text-center">
            <div className="px-4 py-3 bg-gray-800/80 backdrop-blur rounded-lg text-gray-200">
              <p className="text-lg font-medium mb-2">Start Building Your Workflow</p>
              <p className="text-sm text-gray-400">Drag nodes from the sidebar to begin</p>
            </div>
          </Panel>
        )}
      </ReactFlow>
    </div>
  );
}; 