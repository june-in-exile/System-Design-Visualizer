import { memo, useState } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { ComponentType, Warning } from '../types/topology'
import { NODE_TYPE_CONFIG } from './nodeConfig'

interface ArchitectureNodeData {
  label: string
  componentType: ComponentType
  warnings?: Warning[]
  [key: string]: unknown
}

function ArchitectureNode({ data }: NodeProps) {
  const nodeData = data as unknown as ArchitectureNodeData
  const config = NODE_TYPE_CONFIG[nodeData.componentType]
  const hasWarnings = nodeData.warnings && nodeData.warnings.length > 0
  const [showTooltip, setShowTooltip] = useState(false)

  const nodeStyle: React.CSSProperties = {
    position: 'relative',
    padding: '12px 16px',
    borderRadius: 8,
    border: `2px solid ${config.color}`,
    backgroundColor: `${config.color}15`,
    minWidth: 140,
    textAlign: 'center',
    fontSize: 13,
    fontFamily: 'system-ui, sans-serif',
    ...(hasWarnings && {
      animation: 'warningPulse 2s ease-in-out infinite',
    }),
  }

  return (
    <div
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
      style={nodeStyle}
    >
      {showTooltip && hasWarnings && (
        <div
          style={{
            position: 'absolute',
            bottom: '100%',
            left: '50%',
            transform: 'translateX(-50%)',
            marginBottom: 8,
            padding: '8px 12px',
            backgroundColor: 'var(--bg-secondary)',
            border: '1px solid var(--border-color)',
            borderRadius: 6,
            fontSize: 12,
            color: 'var(--text-primary)',
            whiteSpace: 'pre-wrap',
            maxWidth: 250,
            zIndex: 1000,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            textAlign: 'left',
          }}
        >
          {nodeData.warnings!.map((w, i) => (
            <div key={i} style={{ marginBottom: i < nodeData.warnings!.length - 1 ? 8 : 0 }}>
              <div style={{ fontWeight: 600, color: '#f59e0b', marginBottom: 2 }}>
                {w.rule.replace(/_/g, ' ')}
              </div>
              <div style={{ color: 'var(--text-secondary)', fontSize: 11 }}>{w.message}</div>
            </div>
          ))}
        </div>
      )}

      <Handle type="target" position={Position.Top} id="top-target" />
      <Handle type="source" position={Position.Top} id="top-source" />
      
      <Handle type="target" position={Position.Right} id="right-target" />
      <Handle type="source" position={Position.Right} id="right-source" />
      
      <Handle type="target" position={Position.Bottom} id="bottom-target" />
      <Handle type="source" position={Position.Bottom} id="bottom-source" />
      
      <Handle type="target" position={Position.Left} id="left-target" />
      <Handle type="source" position={Position.Left} id="left-source" />

      <div style={{ fontSize: 22, marginBottom: 4 }}>{config.icon}</div>
      <div style={{ fontWeight: 600, color: hasWarnings ? '#f59e0b' : config.color, marginBottom: 2 }}>
        {config.label}
      </div>
      <div style={{ color: 'var(--node-text-secondary)', fontSize: 12 }}>{nodeData.label}</div>
    </div>
  )
}

export default memo(ArchitectureNode)
