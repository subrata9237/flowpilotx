import { Edge, MarkerType } from 'reactflow';
 
export const defaultEdgeOptions: Partial<Edge> = {
  type: 'default',
  animated: false,
  style: { stroke: '#b1b1b7', strokeWidth: 2 },
  markerEnd: { type: MarkerType.ArrowClosed, color: '#b1b1b7' },
}; 