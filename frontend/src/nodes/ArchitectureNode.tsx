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

function ArchitectureNode({ id, data, selected }: NodeProps) {
  const nodeData = data as unknown as ArchitectureNodeData
  const config = NODE_TYPE_CONFIG[nodeData.componentType]
  const roles = nodeData.roles && nodeData.roles.length > 1 ? nodeData.roles : null
  const mergedConfig = roles ? getMergedConfig(roles) : null
  const hasWarnings = nodeData.warnings && nodeData.warnings.length > 0
  const [showTooltip, setShowTooltip] = useState(false)
  const svgRef = useRef<SVGSVGElement>(null)

  const primaryColor = roles ? mergedConfig!.colors[0] : config.color
  const strokeColor = selected ? '#3b82f6' : hasWarnings ? '#f59e0b' : primaryColor

  const width = 160
  const height = 80
  const seed = useMemo(() => stableSeed(id), [id])

  const replicas = (nodeData.properties?.replicas as number) || 1
  const extraLayers = Math.min(replicas - 1, 2)
  const showBadge = replicas > 3
  const extraOffset = extraLayers * 8

  useEffect(() => {
    if (!svgRef.current) return
    const svg = svgRef.current
    while (svg.firstChild) svg.removeChild(svg.firstChild)

    const rc = rough.svg(svg)

    // A) If has warnings, create SVG glow filter
    if (hasWarnings) {
      const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs')
      const filter = document.createElementNS('http://www.w3.org/2000/svg', 'filter')
      filter.setAttribute('id', `glow-${id}`)
      filter.setAttribute('x', '-50%')
      filter.setAttribute('y', '-50%')
      filter.setAttribute('width', '200%')
      filter.setAttribute('height', '200%')
      filter.innerHTML = `
        <feGaussianBlur stdDeviation="3" result="blur" />
        <feFlood flood-color="#f59e0b" flood-opacity="0.35" />
        <feComposite in2="blur" operator="in" result="glow" />
        <feMerge>
          <feMergeNode in="glow" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      `
      defs.appendChild(filter)
      svg.appendChild(defs)
    }

    // B) If selected, create blue glow filter
    if (selected) {
      const defs = document.querySelector('defs') || document.createElementNS('http://www.w3.org/2000/svg', 'defs')
      if (!defs.parentNode) {
        svg.appendChild(defs)
      }
      const filter = document.createElementNS('http://www.w3.org/2000/svg', 'filter')
      filter.setAttribute('id', `selection-${id}`)
      filter.setAttribute('x', '-50%')
      filter.setAttribute('y', '-50%')
      filter.setAttribute('width', '200%')
      filter.setAttribute('height', '200%')
      filter.innerHTML = `
        <feGaussianBlur stdDeviation="4" result="blur" />
        <feFlood flood-color="#3b82f6" flood-opacity="0.5" />
        <feComposite in2="blur" operator="in" result="glow" />
        <feMerge>
          <feMergeNode in="glow" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      `
      defs.appendChild(filter)
      if (!defs.parentNode) svg.appendChild(defs)
    }

    // C) Main rectangle
    const rect = rc.rectangle(4, 4, width - 8, height - 8, {
      stroke: strokeColor,
      strokeWidth: hasWarnings || selected ? 2.5 : 2,
      fill: hasWarnings ? '#f59e0b15' : selected ? '#3b82f615' : `${primaryColor}18`,
      fillStyle: 'solid',
      roughness: 1.5,
      seed,
    })

    if (hasWarnings) {
      rect.setAttribute('filter', `url(#glow-${id})`)
    }
    if (selected) {
      rect.setAttribute('filter', `url(#selection-${id})`)
    }
    svg.appendChild(rect)

    // D) Apply stroke pulse animation to warning paths
    if (hasWarnings) {
      const paths = rect.querySelectorAll('path')
      paths.forEach(path => {
        path.style.animation = 'strokePulse 2s ease-in-out infinite'
      })
    }

    // E) Extra layers for replicas
    if (replicas > 1) {
      for (let i = extraLayers; i >= 1; i--) {
        const offset = i * 8
        const bgRect = rc.rectangle(
          4 + offset, 4 + offset,
          width - 8, height - 8,
          {
            stroke: strokeColor,
            strokeWidth: 1.5,
            fill: hasWarnings ? '#f59e0b08' : selected ? '#3b82f610' : `${primaryColor}10`,
            fillStyle: 'solid',
            roughness: 1.5,
            seed: seed + i,
          }
        )
        svg.insertBefore(bgRect, svg.firstChild)
      }
    }
  }, [id, primaryColor, seed, width, height, replicas, extraLayers, hasWarnings, selected, strokeColor])

  return (
    <div
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
      style={{
        position: 'relative',
        width: width + extraOffset,
        height: height + extraOffset,
        fontFamily: 'var(--font-hand)',
      }}
    >
      {/* Replica Badge */}
      {showBadge && (
        <div style={{
          position: 'absolute',
          top: -8,
          right: -8,
          width: 22,
          height: 22,
          borderRadius: '50%',
          backgroundColor: primaryColor,
          color: '#fff',
          fontSize: 11,
          fontWeight: 700,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 2,
          fontFamily: 'system-ui',
          boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
        }}>
          ×{replicas}
        </div>
      )}

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
        {(roles ? mergedConfig!.icons.some(icon => icon !== '') : config.icon !== '') && (
          <div style={{ fontSize: 20, marginBottom: 2, display: 'flex', gap: 6 }}>
            {roles
              ? mergedConfig!.icons.map((icon, i) => (
                  icon && <span key={i} title={mergedConfig!.descriptions[i]} style={{ cursor: 'help' }}>{icon}</span>
                ))
              : <span title={config.description} style={{ cursor: 'help' }}>{config.icon}</span>
            }
          </div>
        )}
        {roles ? (
          <>
            <div style={{
              fontWeight: 600,
              fontSize: 13,
              color: hasWarnings ? '#f59e0b' : selected ? '#3b82f6' : primaryColor,
              fontFamily: 'var(--font-hand)',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              flexWrap: 'wrap',
              justifyContent: 'center',
            }}>
              {roles.map((role, i) => (
                <span key={role} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  {i > 0 && <span style={{ color: 'var(--text-secondary)', fontSize: 11 }}>+</span>}
                  {NODE_TYPE_CONFIG[role].label}
                </span>
              ))}
            </div>
            <div style={{ color: 'var(--node-text-secondary)', fontSize: 11, fontFamily: 'var(--font-hand)' }}>
              {nodeData.label}
            </div>
          </>
        ) : (
          <>
            <div style={{
              fontWeight: 600,
              fontSize: 14,
              color: hasWarnings ? '#f59e0b' : selected ? '#3b82f6' : primaryColor,
              fontFamily: 'var(--font-hand)',
            }}>
              {config.label}
            </div>
            <div style={{ color: 'var(--node-text-secondary)', fontSize: 12, fontFamily: 'var(--font-hand)' }}>
              {nodeData.label}
            </div>
          </>
        )}
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
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
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
