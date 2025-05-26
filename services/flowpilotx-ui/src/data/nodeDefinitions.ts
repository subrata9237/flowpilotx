import { NodeDefinition } from '../types/workflow';
import { PlusIcon, XMarkIcon } from '@heroicons/react/24/outline';

export const nodeDefinitions: Record<string, NodeDefinition> = {
  'add': {
    type: 'add',
    name: 'Add Numbers',
    icon: PlusIcon,
    description: 'Add two numbers together',
    color: '#9B51E0',
    category: 'Math',
    settings: {
      inputs: {
        a: {
          type: 'number',
          label: 'Input A',
          default: 1,
          description: 'First number to add'
        },
        b: {
          type: 'number',
          label: 'Input B',
          default: 0,
          description: 'Second number to add'
        }
      },
      outputs: {
        result: {
          type: 'number',
          label: 'Result',
          description: 'Sum of inputs'
        }
      },
      config: {
        description: {
          type: 'text',
          label: 'Description',
          default: '',
          optional: true
        },
        isActive: {
          type: 'boolean',
          label: 'Active',
          default: true
        }
      }
    }
  },
  'multiply': {
    type: 'multiply',
    name: 'Multiply Numbers',
    icon: XMarkIcon,
    description: 'Multiply two numbers together',
    color: '#F2994A',
    category: 'Math',
    settings: {
      inputs: {
        a: {
          type: 'number',
          label: 'Input A',
          default: 0,
          description: 'First number to multiply'
        },
        b: {
          type: 'number',
          label: 'Input B',
          default: 0,
          description: 'Second number to multiply'
        }
      },
      outputs: {
        result: {
          type: 'number',
          label: 'Result',
          description: 'Product of inputs'
        }
      },
      config: {
        description: {
          type: 'text',
          label: 'Description',
          default: '',
          optional: true
        },
        isActive: {
          type: 'boolean',
          label: 'Active',
          default: true
        }
      }
    }
  }
};