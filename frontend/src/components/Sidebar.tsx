import type { Node, Edge } from '@xyflow/react'
import type { ComponentType } from '../types/topology'
import { NODE_TYPE_CONFIG } from '../nodes/nodeConfig'
import SettingsMenu from './SettingsMenu'

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
        padding: 16,
        borderRight: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <h3
        style={{
          margin: '0 0 12px',
          fontSize: 14,
          fontWeight: 600,
          color: 'var(--text-secondary)',
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
                fontSize: 13,
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
      
      <SettingsMenu 
        theme={theme} 
        setTheme={setTheme} 
        getNodes={getNodes}
        getEdges={getEdges}
      />
    </aside>
  )
}

export default Sidebar
