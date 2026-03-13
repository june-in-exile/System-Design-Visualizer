import { memo, useState, useMemo, useRef, useEffect } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { ComponentType, Warning } from '../types/topology'
import { NODE_TYPE_CONFIG, getMergedConfig } from './nodeConfig'
import rough from 'roughjs'
import { stableSeed } from '../utils/rough'

interface ArchitectureNodeData {
  label: string
  componentType: ComponentType
  roles?: ComponentType[]
  warnings?: Warning[]
  properties?: Record<string, unknown>
  [key: string]: unknown
}

function ArchitectureNode({ id, data }: NodeProps) {
  const nodeData = data as unknown as ArchitectureNodeData
  const config = NODE_TYPE_CONFIG[nodeData.componentType]
  const roles = nodeData.roles && nodeData.roles.length > 1 ? nodeData.roles : null
  const mergedConfig = roles ? getMergedConfig(roles) : null
  const hasWarnings = nodeData.warnings && nodeData.warnings.length > 0
  const [showTooltip, setShowTooltip] = useState(false)
  const svgRef = useRef<SVGSVGElement>(null)

  const primaryColor = roles ? mergedConfig!.colors[0] : config.color
  const secondaryColor = roles ? mergedConfig!.colors[1] : config.color

  const width = 160
  const height = 80
  const seed = useMemo(() => stableSeed(id), [id])

  const replicas = (nodeData.properties?.replicas as number) || 1
  const extraLayers = Math.min(replicas - 1, 3)
  const extraOffset = extraLayers * 6

  useEffect(() => {
    if (!svgRef.current) return
    const svg = svgRef.current
    while (svg.firstChild) svg.removeChild(svg.firstChild)

    const rc = rough.svg(svg)

    // Main rectangle with hand-drawn border
    const rect = rc.rectangle(4, 4, width - 8, height - 8, {
      stroke: primaryColor,
      strokeWidth: 2,
      fill: `${primaryColor}18`,
      fillStyle: 'solid',
      roughness: 1.5,
      seed,
    })
    svg.appendChild(rect)

    // Extra layers for replicas
    if (replicas > 1) {
      for (let i = extraLayers; i >= 1; i--) {
        const offset = i * 6
        const bgRect = rc.rectangle(
          4 + offset, 4 + offset,
          width - 8, height - 8,
          {
            stroke: primaryColor,
            strokeWidth: 1.5,
            fill: `${primaryColor}10`,
            fillStyle: 'solid',
            roughness: 1.5,
            seed: seed + i,
          }
        )
        svg.insertBefore(bgRect, svg.firstChild)
      }
    }
  }, [primaryColor, seed, width, height, replicas, extraLayers])

  return (
    <div
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
      style={{
        position: 'relative',
        width: width + extraOffset,
        height: height + extraOffset,
        fontFamily: 'var(--font-hand)',
        ...(hasWarnings && {
          animation: 'warningPulse 2s ease-in-out infinite',
        }),
      }}
    >
      {/* Hand-drawn SVG border */}
      <svg
        ref={svgRef}
        width={width + extraOffset}
        height={height + extraOffset}
        style={{ position: 'absolute', top: 0, left: 0 }}
      />

      {/* Text content overlaid on SVG */}
      <div style={{
        position: 'relative',
        zIndex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: height,
        padding: '8px 12px',
      }}>
        <div style={{ fontSize: 20, marginBottom: 2, display: 'flex', gap: 6 }}>
          {roles
            ? mergedConfig!.icons.map((icon, i) => (
                <span key={i} title={mergedConfig!.descriptions[i]} style={{ cursor: 'help' }}>{icon}</span>
              ))
            : <span title={config.description} style={{ cursor: 'help' }}>{config.icon}</span>
          }
        </div>
        <div style={{
          fontWeight: 600,
          fontSize: 14,
          color: hasWarnings ? '#f59e0b' : primaryColor,
          fontFamily: 'var(--font-hand)',
          display: 'flex',
          alignItems: 'center',
          gap: 6,
        }}>
          {config.label}
          {roles && (
            <span
              style={{
                fontSize: 10,
                padding: '2px 6px',
                borderRadius: 12,
                backgroundColor: 'var(--bg-secondary)',
                color: 'var(--text-secondary)',
                border: `1px solid ${secondaryColor}60`,
                cursor: 'help',
              }}
              title={`Merged with: ${roles.filter((r) => r !== nodeData.componentType).map((r) => NODE_TYPE_CONFIG[r].label).join(', ')}`}
            >
              +{roles.length - 1}
            </span>
          )}
        </div>
        <div style={{ color: 'var(--node-text-secondary)', fontSize: 12, fontFamily: 'var(--font-hand)' }}>
          {nodeData.label}
        </div>
      </div>

      {/* Tooltip for warnings */}
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
            fontFamily: 'var(--font-hand)',
            color: 'var(--text-primary)',
            whiteSpace: 'pre-wrap',
            width: 'max-content',
            maxWidth: 400,
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
              <div style={{ color: 'var(--text-secondary)', fontSize: 11, marginBottom: 4 }}>{w.message}</div>
              <div style={{ color: '#b45309', fontSize: 10.5, fontStyle: 'italic', fontWeight: 500 }}>
                {w.solution}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Handles */}
      <Handle type="target" position={Position.Top} id="top-target" />
      <Handle type="source" position={Position.Top} id="top-source" />
      <Handle type="target" position={Position.Right} id="right-target" />
      <Handle type="source" position={Position.Right} id="right-source" />
      <Handle type="target" position={Position.Bottom} id="bottom-target" />
      <Handle type="source" position={Position.Bottom} id="bottom-source" />
      <Handle type="target" position={Position.Left} id="left-target" />
      <Handle type="source" position={Position.Left} id="left-source" />
    </div>
  )
}

export default memo(ArchitectureNode)
