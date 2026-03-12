import { memo, useState } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { ComponentType } from '../types/topology'
import { NODE_TYPE_CONFIG } from './nodeConfig'

interface ArchitectureNodeData {
  label: string
  componentType: ComponentType
  warnings?: string[]
  [key: string]: unknown
}

function ArchitectureNode({ data }: NodeProps) {
  const nodeData = data as unknown as ArchitectureNodeData
  const config = NODE_TYPE_CONFIG[nodeData.componentType]
  const warnings = nodeData.warnings ?? []
  const hasWarnings = warnings.length > 0
  const [showTooltip, setShowTooltip] = useState(false)

  return (
    <div
      style={{
        padding: '12px 16px',
        borderRadius: 8,
        border: `2px solid ${hasWarnings ? '#f59e0b' : config.color}`,
        backgroundColor: `${config.color}15`,
        minWidth: 140,
        textAlign: 'center',
        fontSize: 13,
        fontFamily: 'system-ui, sans-serif',
        boxShadow: hasWarnings
          ? '0 0 12px 2px rgba(245, 158, 11, 0.4)'
          : 'none',
        transition: 'border-color 0.2s, box-shadow 0.2s',
        position: 'relative',
      }}
      onMouseEnter={() => {
        if (hasWarnings) setShowTooltip(true)
      }}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <Handle type="target" position={Position.Top} id="top-target" />
      <Handle type="source" position={Position.Top} id="top-source" />

      <Handle type="target" position={Position.Right} id="right-target" />
      <Handle type="source" position={Position.Right} id="right-source" />

      <Handle type="target" position={Position.Bottom} id="bottom-target" />
      <Handle type="source" position={Position.Bottom} id="bottom-source" />

      <Handle type="target" position={Position.Left} id="left-target" />
      <Handle type="source" position={Position.Left} id="left-source" />

      {hasWarnings && (
        <div
          style={{
            position: 'absolute',
            top: -8,
            right: -8,
            width: 20,
            height: 20,
            borderRadius: '50%',
            backgroundColor: '#f59e0b',
            color: '#fff',
            fontSize: 11,
            fontWeight: 700,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
          }}
        >
          {warnings.length}
        </div>
      )}

      <div style={{ fontSize: 22, marginBottom: 4 }}>{config.icon}</div>
      <div style={{ fontWeight: 600, color: config.color, marginBottom: 2 }}>
        {config.label}
      </div>
      <div style={{ color: '#374151', fontSize: 12 }}>{nodeData.label}</div>

      {showTooltip && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: '50%',
            transform: 'translateX(-50%)',
            marginTop: 8,
            backgroundColor: '#1f2937',
            color: '#f9fafb',
            borderRadius: 6,
            padding: '8px 12px',
            fontSize: 12,
            lineHeight: 1.5,
            width: 280,
            zIndex: 1000,
            textAlign: 'left',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
            pointerEvents: 'none',
          }}
        >
          {warnings.map((msg, i) => (
            <div
              key={i}
              style={{
                marginBottom: i < warnings.length - 1 ? 6 : 0,
                paddingBottom: i < warnings.length - 1 ? 6 : 0,
                borderBottom:
                  i < warnings.length - 1
                    ? '1px solid rgba(255,255,255,0.1)'
                    : 'none',
              }}
            >
              {msg}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default memo(ArchitectureNode)
