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
    '155, 81, 224',
    '78, 201, 176',
    '97, 175, 239',
    '209, 228, 237',
    '242, 153, 74',
    '224, 108, 117',
  ];
  const OPACITIES = [0.2, 0.45, 0.65, 0.85, 1];

  // Helper to extract RGB and opacity from color string
  const parseColor = (color: string) => {
    const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/);
    if (!match) return { rgb: '242, 153, 74', opacity: 0.45 };
    return {
      rgb: `${match[1]}, ${match[2]}, ${match[3]}`,
      opacity: match[4] !== undefined ? parseFloat(match[4]) : 1,
    };
  };

  const { rgb, opacity } = parseColor(data.color || 'rgba(242, 153, 74, 0.45)');

  // Function to determine text color based on background color
  const getTextColor = (bgColor: string): string => {
    const rgb = bgColor.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
    if (!rgb) return 'text-gray-900/90';
    const r = parseInt(rgb[1]);
    const g = parseInt(rgb[2]);
    const b = parseInt(rgb[3]);
    const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
    return luminance > 0.5 ? 'text-gray-900/90' : 'text-gray-100/90';
  };
  const textColor = getTextColor(`rgba(${rgb},${opacity})`);

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
        borderRadius: 12,
        padding: 8,
        backdropFilter: 'blur(8px)',
      }}
      className="shadow-lg group transition-all duration-200"
      tabIndex={-1}
    >
      <NodeResizer
        color="#888"
        isVisible={selected}
        minWidth={120}
        minHeight={80}
        maxWidth={400}
        maxHeight={300}
        handleStyle={{ borderRadius: 6, width: 12, height: 12 }}
      />
      <div className="flex items-center justify-between mb-2">
        <button
          onClick={(e) => {
            e.stopPropagation();
            setShowColors(!showColors);
          }}
          className="w-4 h-4 rounded-full border-2 border-yellow-400/50 hover:border-yellow-400/80"
          style={{ background: `rgba(${rgb},${opacity})` }}
          title="Change color"
        />
        <button
          onClick={handleDelete}
          className="text-xs text-gray-400 hover:text-red-400 transition-colors duration-200"
          title="Delete Sticky Note"
        >
          ×
        </button>
      </div>
      {showColors && (
        <div className="flex items-center gap-1 mb-2" onClick={e => e.stopPropagation()}>
          {COLORS.map(c => (
            <button
              key={c}
              onClick={e => {
                e.stopPropagation();
                const newColor = `rgba(${c},${opacity})`;
                updateNodeData(id, { ...data, color: newColor });
                setShowColors(false);
              }}
              className={`w-4 h-4 rounded-full border-2 transition-all duration-200 ${(data.color || 'rgba(242, 153, 74, 0.45)').startsWith(`rgba(${c}`) ? 'ring-2 ring-yellow-400/80 border-yellow-400/80' : 'border-transparent hover:border-yellow-400/50'}`}
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
          <span className="ml-1 text-xs text-gray-500 select-none">{Math.round(opacity * 100)}%</span>
        </div>
      )}
      <textarea
        value={data.text}
        onChange={e => updateNodeData(id, { ...data, text: e.target.value })}
        className={`w-full h-full bg-transparent border-none outline-none resize-none font-medium text-base ${textColor} placeholder:text-gray-500/50 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]`}
        placeholder="Write a note..."
        style={{ height: 'calc(100% - 40px)', overflow: 'hidden' }}
      />
    </div>
  );
}; 