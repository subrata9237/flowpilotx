import { NodeDefinition } from '../types/workflow';
import { PlusIcon, XMarkIcon, MinusIcon, DivideIcon } from '@heroicons/react/24/outline';

export const nodeDefinitions: Record<string, NodeDefinition> = {
  'add': {
    type: 'add',
    name: 'Add Numbers',
    icon: PlusIcon,
    description: 'Add two numbers together',
    color: '#9B51E0',
    category: 'Math',
    calculate: (inputs: { [key: string]: any }) => {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: a + b };
    },
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
          default: false
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
    calculate: (inputs: { [key: string]: any }) => {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: a * b };
    },
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
  },
  'subtract': {
    type: 'subtract',
    name: 'Subtract Numbers',
    icon: MinusIcon,
    description: 'Subtract second number from first number',
    color: '#EB5757',
    category: 'Math',
    calculate: (inputs: { [key: string]: any }) => {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: a - b };
    },
    settings: {
      inputs: {
        a: {
          type: 'number',
          label: 'Input A',
          default: 0,
          description: 'Number to subtract from'
        },
        b: {
          type: 'number',
          label: 'Input B',
          default: 0,
          description: 'Number to subtract'
        }
      },
      outputs: {
        result: {
          type: 'number',
          label: 'Result',
          description: 'Difference of inputs'
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
  'divide': {
    type: 'divide',
    name: 'Divide Numbers',
    icon: DivideIcon,
    description: 'Divide first number by second number',
    color: '#4F46E5',
    category: 'Math',
    calculate: (inputs: { [key: string]: any }) => {
      const a = Number(inputs.a) || 0;
      const b = Number(inputs.b) || 0;
      return { result: b !== 0 ? a / b : 0 };
    },
    settings: {
      inputs: {
        a: {
          type: 'number',
          label: 'Input A',
          default: 0,
          description: 'Number to divide'
        },
        b: {
          type: 'number',
          label: 'Input B',
          default: 0,
          description: 'Number to divide by'
        }
      },
      outputs: {
        result: {
          type: 'number',
          label: 'Result',
          description: 'Quotient of inputs'
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
