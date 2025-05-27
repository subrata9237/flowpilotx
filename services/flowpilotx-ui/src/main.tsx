if (process.env.NODE_ENV === 'development') {
  const suppressedWarnings = [
    'ResizeObserver loop completed with undelivered notifications.',
    'ResizeObserver loop limit exceeded'
  ];
  const originalWarn = console.warn;
  console.warn = (...args) => {
    if (typeof args[0] === 'string' && suppressedWarnings.some(w => args[0].includes(w))) {
      return;
    }
    originalWarn(...args);
  };
}

export {}; 