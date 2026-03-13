import { memo } from 'react'
import {
  BaseEdge,
  getBezierPath,
  EdgeLabelRenderer,
  type EdgeProps,
} from '@xyflow/react'

function HandDrawnEdge({
  id,
  sourceX, sourceY,
  targetX, targetY,
  sourcePosition,
  targetPosition,
  data,
  style,
  markerEnd,
  markerStart,
}: EdgeProps) {
  const edgeData = (data as Record<string, unknown>) ?? {}
  const isAsync = edgeData.connectionType === 'async'
  const protocol = (edgeData.protocol as string) || ''
  const connType = (edgeData.connectionType as string) || ''

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX, sourceY,
    targetX, targetY,
    sourcePosition,
    targetPosition,
  })

  const parts: string[] = []
  if (protocol) parts.push(protocol.toUpperCase())
  if (connType && connType !== 'unspecified') parts.push(connType)
  const label = parts.join(' · ')

  return (
    <>
      <g filter="url(#roughen)">
        <BaseEdge
          id={id}
          path={edgePath}
          style={{
            ...style,
            strokeWidth: 2,
            strokeDasharray: isAsync ? '8,6' : undefined,
          }}
          markerEnd={markerEnd}
          markerStart={markerStart}
        />
        <rect 
          x={Math.min(sourceX, targetX) - 20}
          y={Math.min(sourceY, targetY) - 20}
          width={Math.abs(sourceX - targetX) + 40}
          height={Math.abs(sourceY - targetY) + 40}
          fill="transparent" 
          stroke="transparent" 
          pointerEvents="none" 
        />
      </g>

      {label && (
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              fontFamily: 'var(--font-hand)',
              fontSize: 11,
              fontWeight: 500,
              color: 'var(--text-secondary)',
              background: 'var(--bg-primary)',
              padding: '2px 8px',
              borderRadius: 4,
              pointerEvents: 'none',
            }}
          >
            {label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  )
}

export default memo(HandDrawnEdge)
