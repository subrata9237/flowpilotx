import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { ReactFlowProvider } from 'reactflow';
import { Sidebar } from './components/Sidebar';
import { WorkflowEditor } from './components/WorkflowEditor';
import NodeSettingsPage from './pages/NodeSettingsPage';
import { Layout } from './components/Layout';

const App: React.FC = () => {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={
            <DndProvider backend={HTML5Backend}>
              <ReactFlowProvider>
                <div className="flex h-screen overflow-hidden">
                  <Sidebar />
                  <WorkflowEditor />
                </div>
              </ReactFlowProvider>
            </DndProvider>
          } />
          <Route path="/node/:nodeId/settings" element={<NodeSettingsPage />} />
        </Routes>
      </Layout>
    </Router>
  );
};

export default App;
