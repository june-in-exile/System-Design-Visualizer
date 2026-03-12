import Sidebar from './components/Sidebar'
import Canvas from './components/Canvas'

function App() {
  return (
    <div
      style={{
        display: 'flex',
        height: '100vh',
        width: '100vw',
        fontFamily: 'system-ui, sans-serif',
      }}
    >
      <Sidebar />
      <Canvas />
    </div>
  )
}

export default App
