import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { ComponentType } from '../types/topology'
import { NODE_TYPE_CONFIG } from './nodeConfig'

interface ArchitectureNodeData {
  label: string
  componentType: ComponentType
  [key: string]: unknown
}

function ArchitectureNode({ data }: NodeProps) {
  const nodeData = data as unknown as ArchitectureNodeData
  const config = NODE_TYPE_CONFIG[nodeData.componentType]

  return (
    <div
      style={{
        padding: '12px 16px',
        borderRadius: 8,
        border: `2px solid ${config.color}`,
        backgroundColor: `${config.color}15`,
        minWidth: 140,
        textAlign: 'center',
        fontSize: 13,
        fontFamily: 'system-ui, sans-serif',
      }}
    >
      <Handle type="target" position={Position.Top} />
      <div style={{ fontSize: 22, marginBottom: 4 }}>{config.icon}</div>
      <div style={{ fontWeight: 600, color: config.color, marginBottom: 2 }}>
        {config.label}
      </div>
      <div style={{ color: '#374151', fontSize: 12 }}>{nodeData.label}</div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

export default memo(ArchitectureNode)
