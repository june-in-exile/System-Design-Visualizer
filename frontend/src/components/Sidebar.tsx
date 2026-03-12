import type { ComponentType } from '../types/topology'
import { NODE_TYPE_CONFIG } from '../nodes/nodeConfig'

const COMPONENT_TYPES = Object.keys(NODE_TYPE_CONFIG) as ComponentType[]

function Sidebar() {
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
        borderRight: '1px solid #e5e7eb',
        backgroundColor: '#f9fafb',
        overflowY: 'auto',
      }}
    >
      <h3
        style={{
          margin: '0 0 12px',
          fontSize: 14,
          fontWeight: 600,
          color: '#374151',
        }}
      >
        Components
      </h3>
      {COMPONENT_TYPES.map((type) => {
        const config = NODE_TYPE_CONFIG[type]
        return (
          <div
            key={type}
            draggable
            onDragStart={(e) => onDragStart(e, type)}
            style={{
              padding: '8px 12px',
              marginBottom: 8,
              borderRadius: 6,
              border: `1px solid ${config.color}40`,
              backgroundColor: '#ffffff',
              cursor: 'grab',
              fontSize: 13,
              display: 'flex',
              alignItems: 'center',
              gap: 8,
            }}
          >
            <span>{config.icon}</span>
            <span>{config.label}</span>
          </div>
        )
      })}
    </aside>
  )
}

export default Sidebar
