import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { ReactFlowProvider } from 'reactflow';
import { Sidebar } from './components/Sidebar';
import { WorkflowEditor } from './components/WorkflowEditor';
import NodeSettingsPage from './pages/NodeSettingsPage';
import { Layout } from './components/Layout';
import { Home } from './pages/Home';
import { useWorkflowStore } from './store/workflowStore';

const App: React.FC = () => {
  const theme = useWorkflowStore((state) => state.theme);

  return (
    <Router>
      <div className={theme === 'vscode' ? 'dark' : 'light'}>
        <Layout>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route 
              path="/workflow/new" 
              element={
                <DndProvider backend={HTML5Backend}>
                  <ReactFlowProvider>
                    <div className="flex h-screen overflow-hidden">
                      <Sidebar />
                      <WorkflowEditor />
                    </div>
                  </ReactFlowProvider>
                </DndProvider>
              } 
            />
            <Route 
              path="/workflow/:id" 
              element={
                <DndProvider backend={HTML5Backend}>
                  <ReactFlowProvider>
                    <div className="flex h-screen overflow-hidden">
                      <Sidebar />
                      <WorkflowEditor />
                    </div>
                  </ReactFlowProvider>
                </DndProvider>
              } 
            />
            <Route path="/workflows" element={<Navigate to="/" />} />
            <Route path="/workflows/recent" element={<Navigate to="/" />} />
            <Route path="/credentials" element={<Navigate to="/" />} />
            <Route path="/credentials/new" element={<Navigate to="/" />} />
            <Route path="/env" element={<Navigate to="/" />} />
            <Route path="/env/new" element={<Navigate to="/" />} />
            <Route path="/secrets" element={<Navigate to="/" />} />
            <Route path="/secrets/new" element={<Navigate to="/" />} />
            <Route path="/history" element={<Navigate to="/" />} />
            <Route path="/history/recent" element={<Navigate to="/" />} />
            <Route path="/node/:nodeId/settings" element={<NodeSettingsPage />} />
          </Routes>
        </Layout>
      </div>
    </Router>
  );
};

export default App;
