import { useState, useEffect, useCallback } from 'react'
import Sidebar from './components/Sidebar'
import Canvas from './components/Canvas'
import TabBar from './components/TabBar'
import { useCanvasTabs } from './hooks/useCanvasTabs'

function App() {
  const [isDarkMode, setIsDarkMode] = useState(() => {
    const saved = localStorage.getItem('theme')
    return saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)
  })

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }, [isDarkMode])

  const toggleDarkMode = () => setIsDarkMode(prev => !prev)

  const {
    tabs,
    activeTab,
    activeTabId,
    addTab,
    switchTab,
    closeTab,
    renameTab,
    updateCanvasStateRef,
    getCurrentState,
  } = useCanvasTabs()

  const handleCanvasStateChange = useCallback(
    (nodes: import('@xyflow/react').Node[], edges: import('@xyflow/react').Edge[]) => {
      updateCanvasStateRef(nodes, edges)
    },
    [updateCanvasStateRef]
  )

  return (
    <div
      style={{
        display: 'flex',
        height: '100vh',
        width: '100vw',
        fontFamily: 'system-ui, sans-serif',
        backgroundColor: 'var(--bg-primary)',
        color: 'var(--text-primary)',
      }}
    >
      <Sidebar 
        isDarkMode={isDarkMode} 
        toggleDarkMode={toggleDarkMode} 
        getNodes={() => getCurrentState().nodes}
        getEdges={() => getCurrentState().edges}
      />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <TabBar
          tabs={tabs}
          activeTabId={activeTabId}
          onSwitchTab={switchTab}
          onAddTab={addTab}
          onCloseTab={closeTab}
          onRenameTab={renameTab}
        />
        <Canvas
          key={activeTabId}
          isDarkMode={isDarkMode}
          initialNodes={[...activeTab.nodes]}
          initialEdges={[...activeTab.edges]}
          onStateChange={handleCanvasStateChange}
        />
      </div>
    </div>
  )
}

export default App
