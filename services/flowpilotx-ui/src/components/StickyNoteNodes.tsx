import React, { useState, useRef } from 'react';
import { NodeProps, NodeResizer } from 'reactflow';
import { NodeData } from '../types/workflow';
import { useWorkflowStore } from '../store/workflowStore';

export const StickyNoteNodes: React.FC<NodeProps<NodeData>> = ({ id, data, selected }) => {
  const [showColors, setShowColors] = useState(false);
  const noteRef = useRef<HTMLDivElement>(null);
  const updateNodeData = useWorkflowStore(state => state.updateNodeData);
  const deleteNode = useWorkflowStore(state => state.deleteNode);

  const COLORS = [
    '255, 255, 204', // n8n yellow
    '255, 230, 204', // light orange
    '204, 255, 229', // light green
    '204, 229, 255', // light blue
    '255, 204, 229', // light pink
    '224, 224, 224', // light gray
    '255, 255, 255', // white
    //Purple
    '224, 204, 255', // Purple
    '255, 204, 229', // Orange
    //RED
    '255, 204, 204', // Red
    //BLUE
    '204, 229, 255', // Blue
    //GREEN
    '204, 255, 204', // Green
    //YELLOW
    
  ];

  const OPACITIES = [0.1,0.2, 0.45, 0.65, 0.85, 1];

  // Helper to extract RGB and opacity from color string
  const parseColor = (color: string) => {
    const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/);
    if (!match) return { rgb: '255, 255, 204', opacity: 1 };
    return {
      rgb: `${match[1]}, ${match[2]}, ${match[3]}`,
      opacity: match[4] !== undefined ? parseFloat(match[4]) : 1,
    };
  };

  const { rgb, opacity } = parseColor(data.color || 'rgba(255, 255, 204, 1)');

  // Change opacity logic
  const handleChangeOpacity = (e: React.MouseEvent) => {
    e.stopPropagation();
    const currentIdx = OPACITIES.findIndex(o => Math.abs(o - opacity) < 0.01);
    const nextIdx = (currentIdx + 1) % OPACITIES.length;
    const newColor = `rgba(${rgb},${OPACITIES[nextIdx]})`;
    updateNodeData(id, { ...data, color: newColor });
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    const nodeElement = document.querySelector(`[data-id="${id}"]`);
    if (nodeElement) {
      nodeElement.classList.add('opacity-0', 'scale-95', 'transition-all', 'duration-200');
      setTimeout(() => deleteNode(id), 200);
    }
  };

  return (
    <div
      ref={noteRef}
      style={{
        background: `rgba(${rgb},${opacity})`,
        zIndex: 1,
        position: 'relative',
        pointerEvents: 'auto',
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        borderRadius: 8,
        padding: 0,
        minHeight: 60,
        minWidth: 120,
        maxWidth: 400,
        maxHeight: 300,
        display: 'flex',
        flexDirection: 'column',
      }}
      className="group transition-all duration-200"
      tabIndex={-1}
    >
      <NodeResizer
      //color transparent
        color="rgba(255, 255, 255, 0)"
        isVisible={selected}
        handleStyle={{ borderRadius: 60, width: 12, height: 12 }}
      />
      {/* Minimal header for color and delete */}
      <div className="flex items-center justify-between px-2 py-1 group/header" style={{ minHeight: 32 }}>
        <div className="flex items-center gap-1">
        <button
          onClick={(e) => {
            e.stopPropagation();
            setShowColors(!showColors);
          }}
            className="w-4 h-4 rounded-full border-2 border-gray-300 hover:border-gray-500 opacity-0 group-hover:opacity-100 group-hover/header:opacity-100 transition-opacity duration-200"
          style={{ background: `rgba(${rgb},${opacity})` }}
          title="Change color"
        />
      {showColors && (
            <div className="flex items-center gap-1 ml-2" onClick={e => e.stopPropagation()}>
          {COLORS.map(c => (
            <button
              key={c}
              onClick={e => {
                e.stopPropagation();
                    const newColor = `rgba(${c},${opacity})`;
                    updateNodeData(id, { ...data, color: newColor });
                setShowColors(false);
              }}
                  className={`w-4 h-4 rounded-full border-2 transition-all duration-200 ${(data.color || 'rgba(255, 255, 204, 1)').startsWith(`rgba(${c}`) ? 'ring-2 ring-yellow-400/80 border-yellow-400/80' : 'border-transparent hover:border-yellow-400/50'}`}
              style={{ background: `rgba(${c},${opacity})` }}
              tabIndex={-1}
              aria-label={`Set color rgba(${c},${opacity})`}
            />
          ))}
          <button
            onClick={handleChangeOpacity}
            className="w-4 h-4 rounded-full bg-white/70 text-xs flex items-center justify-center border border-gray-300 hover:bg-white/90 transition-all duration-200 ml-2"
            title="Change opacity"
            style={{ fontSize: '0.9em' }}
          >
            <span role="img" aria-label="opacity">💧</span>
          </button>
            </div>
          )}
        </div>
        <button
          onClick={handleDelete}
          className="text-xs text-gray-400 hover:text-red-400 transition-colors duration-200 opacity-0 group-hover:opacity-100 group-hover/header:opacity-100 transition-opacity"
          title="Delete Sticky Note"
        >
          ×
        </button>
      </div>
      {/* Editable text area, auto-expanding */}
      <textarea
        value={data.text}
        onChange={e => updateNodeData(id, { ...data, text: e.target.value })}
        className={`bg-transparent border-none outline-none resize-none font-medium text-base px-2 py-1 flex-1 ${data.text ? 'text-gray-900' : 'text-gray-500'} placeholder:text-gray-400`}
        placeholder=""
        style={{
          width: '100%',
          height: '100%',
          minHeight: 32,
          maxHeight: 240,
          overflow: 'auto',
          flex: 1,
        }}
        spellCheck={false}
      />
    </div>
  );
}; 