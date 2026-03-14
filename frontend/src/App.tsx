import { useState, useEffect, useCallback } from 'react'
import Sidebar from './components/Sidebar'
import Canvas from './components/Canvas'
import TabBar from './components/TabBar'
import { useCanvasTabs } from './hooks/useCanvasTabs'

function App() {
  const [theme, setTheme] = useState<'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk'>(() => {
    const saved = localStorage.getItem('theme') as 'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk'
    if (saved) return saved
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })

  useEffect(() => {
    document.documentElement.classList.remove('dark', 'warm', 'dream', 'cyberpunk')
    if (theme !== 'light') {
      document.documentElement.classList.add(theme)
    }
    localStorage.setItem('theme', theme)
  }, [theme])

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
        theme={theme} 
        setTheme={setTheme}
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
          theme={theme}
          initialNodes={[...activeTab.nodes]}
          initialEdges={[...activeTab.edges]}
          onStateChange={handleCanvasStateChange}
        />
      </div>
    </div>
  )
}

export default App
