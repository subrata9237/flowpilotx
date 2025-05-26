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
  MarkerType,
  Viewport,
} from 'reactflow';
import { useWorkflowStore } from '../store/workflowStore';
import { BaseNode } from './BaseNode';
import { NodeData } from '../types/workflow';
import { nodeDefinitions } from '../data/nodeDefinitions';
import { ToolbarHeader } from './ToolbarHeader';
import 'reactflow/dist/style.css';
import { StickyNoteNodes } from './StickyNoteNodes';

// Create nodeTypes dynamically from nodeDefinitions
const nodeTypes = {
  ...Object.fromEntries(Object.keys(nodeDefinitions).map(type => [type, BaseNode])),
  sticky: StickyNoteNodes,
};

const defaultEdgeOptions = {
  animated: false,
  type: 'straight',
  style: {
    stroke: '#555',
    strokeWidth: 2
  },
  markerEnd: {
    type: MarkerType.ArrowClosed,
    color: '#555',
  },
  className: 'react-flow__edge-path-selector'
};

export const WorkflowEditor: React.FC = () => {
  const { 
    nodes, 
    edges, 
    addNode, 
    addEdge, 
    updateNodePosition, 
    deleteNode, 
    theme, 
    setEdges,
    viewport,
    setViewport,
    showStickyNotes,
    setNodes,
  } = useWorkflowStore();
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);
  const [selectedEdges, setSelectedEdges] = useState<string[]>([]);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [showDeleteTooltip, setShowDeleteTooltip] = useState(false);

  // Track delete key press
  const deletePressed = useKeyPress(['Delete', 'Backspace']);

  // Filter nodes based on showStickyNotes state
  const visibleNodes = nodes.filter(node => 
    node.type !== 'sticky' || showStickyNotes
  );

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
    onChange: ({ nodes, edges }) => {
      const selectedNodeIds = nodes.map(node => node.id);
      const selectedEdgeIds = edges.map(edge => edge.id);
      setSelectedNodes(selectedNodeIds);
      setSelectedEdges(selectedEdgeIds);
      setShowDeleteTooltip(selectedNodeIds.length > 0 || selectedEdgeIds.length > 0);
    },
  });

  // Handle deletion with animation
  const handleDeletion = useCallback(() => {
    // Handle node deletion
    selectedNodes.forEach(nodeId => {
      const nodeElement = document.querySelector(`[data-id="${nodeId}"]`);
      if (nodeElement) {
        nodeElement.classList.add('scale-95', 'opacity-50');
      }
    });

    // Handle edge deletion
    selectedEdges.forEach(edgeId => {
      const edgeElement = document.querySelector(`[data-id="${edgeId}"]`);
      if (edgeElement) {
        edgeElement.classList.add('opacity-0');
      }
    });

    setTimeout(() => {
      // Delete nodes
      selectedNodes.forEach(nodeId => deleteNode(nodeId));
      // Delete edges
      if (selectedEdges.length > 0) {
        setEdges(edges.filter(edge => !selectedEdges.includes(edge.id)));
      }
      setShowDeleteTooltip(false);
    }, 150);
  }, [selectedNodes, selectedEdges, deleteNode, setEdges, edges]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.key === 'Delete' || event.key === 'Backspace') && 
          (selectedNodes.length > 0 || selectedEdges.length > 0)) {
        handleDeletion();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodes, selectedEdges, handleDeletion]);

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
      const nodeDefinition = nodeDefinitions[nodeData.type];
      
      // Initialize config with default values
      const config: Record<string, any> = {};
      const inputs: Record<string, any> = {};
      const outputs: Record<string, any> = {};

      if (nodeDefinition) {
        // Initialize inputs from node definition
        Object.entries(nodeDefinition.settings.inputs).forEach(([key, setting]) => {
          inputs[key] = setting.default;
        });

        // Initialize outputs from node definition
        Object.entries(nodeDefinition.settings.outputs).forEach(([key, setting]) => {
          outputs[key] = setting.default || 0;
        });

        // Initialize config
        Object.entries(nodeDefinition.settings.config).forEach(([key, setting]) => {
          config[key] = setting.default;
        });
      }

      const newNode = {
        id: `${nodeData.type}-${Date.now()}`,
        type: nodeData.type,
        position,
        data: {
          name: nodeData.name,
          type: nodeData.type,
          description: nodeData.description,
          inputs: inputs,
          outputs: outputs,
          config: config
        } as NodeData,
      };

      addNode(newNode);
      
      if (!viewport) {
        reactFlowInstance.setViewport({
          x: 0,
          y: 0,
          zoom: 1.0
        });
      }
    } catch (error) {
      console.error('Failed to handle node drop:', error);
    }
  }, [reactFlowInstance, addNode, viewport]);

  // Add this function for validation
  const isValidConnection = useCallback((connection: Connection): boolean => {
    if (!connection.source || !connection.target) return false;
    
    // Check for duplicate connection
    return !edges.some(edge => 
      edge.source === connection.source && 
      edge.target === connection.target
    );
  }, [edges]);

  // Modify onConnect to handle the actual connection
  const onConnect = useCallback((connection: Connection) => {
    if (connection.source && connection.target) {
      const edge: Edge = {
        id: `${connection.source}-${connection.target}`,
        source: connection.source,
        target: connection.target,
        sourceHandle: connection.sourceHandle,
        targetHandle: connection.targetHandle,
        type: 'straight'
      };
      addEdge(edge);
    }
  }, [addEdge]);

  const onNodeDragStop: NodeDragHandler = useCallback(
    (event, node) => {
      updateNodePosition(node.id, node.position);
    },
    [updateNodePosition]
  );

  const onNodesChange: OnNodesChange = useCallback((changes) => {
    changes.forEach((change) => {
      if (change.type === 'position' && change.position && change.id) {
        // Update node position
        updateNodePosition(change.id, change.position);
        
        // Update connected edges
        const connectedEdges = edges.filter(
          edge => edge.source === change.id || edge.target === change.id
        );
        
        if (connectedEdges.length > 0) {
          const updatedEdges = edges.map(edge => {
            if (edge.source === change.id || edge.target === change.id) {
              return {
                ...edge,
                // This ensures smooth edge updates
                type: 'bezier',
                animated: false
              };
            }
            return edge;
          });
          setEdges(updatedEdges);
        }
      }
      if (change.type === 'dimensions' && change.id && change.dimensions) {
        const { width, height } = change.dimensions;
        if (width !== undefined && height !== undefined) {
          setNodes((nodes) =>
            nodes.map((node) =>
              node.id === change.id
                ? { ...node, width, height }
                : node
            )
          );
        }
      }
    });
  }, [updateNodePosition, edges, setEdges, setNodes]);

  const onEdgesChange: OnEdgesChange = useCallback((changes) => {
    changes.forEach((change) => {
      if (change.type === 'add' && change.item) {
        addEdge(change.item as Edge);
      }
    });
  }, [addEdge]);

  // Save viewport on change
  const onMoveEnd = useCallback((event: any, viewport: Viewport) => {
    setViewport(viewport);
  }, [setViewport]);

  const handleFitView = useCallback(() => {
    if (reactFlowInstance) {
      reactFlowInstance.fitView({
        padding: 0.2,
        maxZoom: 1.2,
        duration: 200
      });
    }
  }, [reactFlowInstance]);

  return (
    <div className="w-full h-screen flex flex-col">
      <ToolbarHeader />
      <div className="flex-1 h-full relative" ref={reactFlowWrapper} onDrop={onDrop} onDragOver={onDragOver}>
        <ReactFlow
          style={{ position: 'absolute', inset: 0, zIndex: 10 }}
          nodes={visibleNodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          defaultEdgeOptions={defaultEdgeOptions}
          onConnect={onConnect}
          onNodeDragStop={onNodeDragStop}
          onInit={setReactFlowInstance}
          onMoveEnd={onMoveEnd}
          defaultViewport={viewport || { x: 0, y: 0, zoom: 1.0 }}
          className={theme === 'vscode' ? 'bg-node-vscode-bg' : 'bg-node-miro-bg'}
          minZoom={0.1}
          maxZoom={4}
          snapToGrid
          snapGrid={[16, 16]}
          fitView={false}
          fitViewOptions={{ 
            padding: 0.2,
            maxZoom: 1.2,
            duration: 200
          }}
          elementsSelectable={true}
          selectNodesOnDrag={false}
          proOptions={{ hideAttribution: true }}
          isValidConnection={isValidConnection}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={16}
            size={1}
            color={theme === 'vscode' ? '#404040' : '#E6E6E6'}
          />
          <Controls 
            showFitView={true}
            fitViewOptions={{ 
              padding: 0.2,
              maxZoom: 1.2,
              duration: 200
            }}
            className={`
              ${theme === 'vscode' 
                ? 'bg-node-vscode-bg border-node-vscode-border' 
                : 'bg-node-miro-bg border-node-miro-border'
              }
              border rounded-lg shadow-lg
            `}
          />
          <MiniMap
            nodeColor={theme === 'vscode' ? '#3C3C3C' : '#F5F5F5'}
            maskColor={theme === 'vscode' ? 'rgba(45, 45, 45, 0.9)' : 'rgba(255, 255, 255, 0.9)'}
            className={`
              ${theme === 'vscode' 
                ? 'bg-node-vscode-bg border-node-vscode-border hover:border-node-vscode-selected' 
                : 'bg-node-miro-bg border-node-miro-border hover:border-node-miro-selected'
              }
              border-2 rounded-lg shadow-lg transition-all duration-200
            `}
            style={{
              backgroundColor: theme === 'vscode' ? '#252526' : '#FFFFFF',
              border: `2px solid ${theme === 'vscode' ? '#454545' : '#E0E0E0'}`,
            }}
          />
          {showDeleteTooltip && (
            <Panel position="top-right" className="m-2.5">
              <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm
                ${deletePressed ? 'bg-red-600 text-white' : 'bg-gray-800 text-gray-200'}
                transition-colors duration-150`}
              >
                <span className="font-medium">
                  {selectedNodes.length > 0 && `${selectedNodes.length} node${selectedNodes.length > 1 ? 's' : ''}`}
                  {selectedNodes.length > 0 && selectedEdges.length > 0 && ' and '}
                  {selectedEdges.length > 0 && `${selectedEdges.length} connection${selectedEdges.length > 1 ? 's' : ''}`}
                  {' selected'}
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
    </div>
  );
}; 