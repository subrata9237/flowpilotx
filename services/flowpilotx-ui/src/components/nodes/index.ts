import { BaseNode } from '../BaseNode';
import { StickyNoteNodes } from '../StickyNoteNodes';
import { nodeDefinitions } from '../../data/nodeDefinitions';

// Dynamically create nodeTypes mapping
const nodeTypes: Record<string, any> = {
  stickyNote: StickyNoteNodes,
};

Object.keys(nodeDefinitions).forEach(type => {
  nodeTypes[type] = BaseNode;
});

export { nodeTypes }; 