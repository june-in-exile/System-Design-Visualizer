import { useState } from 'react'
import type { Warning } from '../types/topology'

interface WarningsPanelProps {
  warnings: Warning[]
  onWarningClick: (nodeIds: string[]) => void
  onWarningHover: (nodeIds: string[]) => void
}

const RULE_COLORS: Record<string, string> = {
  spof: '#ef4444',
  db_selection: '#f59e0b',
  federation: '#8b5cf6',
  cache_consistency: '#3b82f6',
  cap_theorem: '#ec4899',
  schema: '#6b7280',
}

function WarningsPanel({
  warnings,
  onWarningClick,
  onWarningHover,
}: WarningsPanelProps) {
  const [collapsed, setCollapsed] = useState(false)

  if (warnings.length === 0) return null

  return (
    <div
      style={{
        borderTop: '1px solid #e5e7eb',
        backgroundColor: '#fffbeb',
        maxHeight: collapsed ? 36 : 220,
        overflow: 'hidden',
        transition: 'max-height 0.2s ease',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div
        style={{
          padding: '8px 16px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          cursor: 'pointer',
          userSelect: 'none',
          borderBottom: collapsed ? 'none' : '1px solid #fde68a',
        }}
        onClick={() => setCollapsed((v) => !v)}
      >
        <span style={{ fontSize: 13, fontWeight: 600, color: '#92400e' }}>
          {warnings.length} warning{warnings.length > 1 ? 's' : ''}
        </span>
        <span style={{ fontSize: 11, color: '#b45309' }}>
          {collapsed ? '展開 ▲' : '收合 ▼'}
        </span>
      </div>
      {!collapsed && (
        <div style={{ overflowY: 'auto', flex: 1 }}>
          {warnings.map((w, i) => (
            <div
              key={i}
              onClick={() => onWarningClick(w.nodeIds)}
              onMouseEnter={() => onWarningHover(w.nodeIds)}
              onMouseLeave={() => onWarningHover([])}
              style={{
                padding: '8px 16px',
                fontSize: 13,
                lineHeight: 1.5,
                cursor: 'pointer',
                borderBottom: '1px solid #fef3c7',
                display: 'flex',
                alignItems: 'flex-start',
                gap: 8,
              }}
            >
              <span
                style={{
                  display: 'inline-block',
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  backgroundColor: RULE_COLORS[w.rule] ?? '#6b7280',
                  marginTop: 7,
                  flexShrink: 0,
                }}
              />
              <span style={{ color: '#78350f' }}>{w.message}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default WarningsPanel
