import type { Node, Edge } from '@xyflow/react'
import type { ComponentType } from '../types/topology'
import { NODE_TYPE_CONFIG } from '../nodes/nodeConfig'
import SettingsMenu from './SettingsMenu'
import logo from '../assets/logo.svg'

const COMPONENT_TYPES = Object.keys(NODE_TYPE_CONFIG) as ComponentType[]

interface SidebarProps {
  theme: 'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk';
  setTheme: (theme: 'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk') => void;
  getNodes: () => Node[];
  getEdges: () => Edge[];
}

function Sidebar({ theme, setTheme, getNodes, getEdges }: SidebarProps) {
  const onDragStart = (
    event: React.DragEvent,
    componentType: ComponentType
  ) => {
    event.dataTransfer.setData('application/architectmind', componentType)
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <aside
      style={{
        width: 200,
        padding: 0,
        borderRight: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Brand Header at Top-Left */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '20px 16px',
          borderBottom: '1px solid var(--border-color)',
          marginBottom: 16,
          backgroundColor: 'var(--bg-primary)',
        }}
      >
        <img 
          src={logo} 
          alt="ArchitectMind Logo" 
          style={{ 
            width: 36, 
            height: 36, 
            borderRadius: 8,
            objectFit: 'contain'
          }} 
        />
        <h2
          style={{
            margin: 0,
            fontSize: 22,
            fontWeight: 700,
            letterSpacing: '-0.01em',
            color: 'var(--text-primary)',
          }}
        >
          Architect<span style={{ color: 'var(--accent)' }}>Mind</span>
        </h2>
      </div>

      <div style={{ padding: '0 16px', flex: 1, display: 'flex', flexDirection: 'column' }}>
        <h3
          style={{
            margin: '0 0 12px',
            fontSize: 14,
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontWeight: 700,
            color: 'var(--text-secondary)',
            opacity: 0.8,
          }}
        >
          Components
        </h3>
        <div style={{ flex: 1 }}>
          {COMPONENT_TYPES.map((type) => {
            const config = NODE_TYPE_CONFIG[type]
            return (
              <div
                key={type}
                draggable
                onDragStart={(e) => onDragStart(e, type)}
                title={config.description}
                style={{
                  padding: '8px 12px',
                  marginBottom: 8,
                  borderRadius: 6,
                  border: `1px solid ${config.color}40`,
                  backgroundColor: 'var(--bg-primary)',
                  cursor: 'grab',
                  fontSize: 16,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  color: 'var(--text-primary)',
                }}
              >
                {config.icon && <span>{config.icon}</span>}
                <span>{config.label}</span>
              </div>
            )
          })}
        </div>
      </div>
      
      <div style={{ padding: 16 }}>
        <SettingsMenu 
          theme={theme} 
          setTheme={setTheme} 
          getNodes={getNodes}
          getEdges={getEdges}
        />
      </div>
    </aside>
  )
}

export default Sidebar
