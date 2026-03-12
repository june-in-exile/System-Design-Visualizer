import { useState } from 'react'
import type { Warning } from '../types/topology'

interface WarningsPanelProps {
  warnings: Warning[]
  onWarningClick: (nodeIds: string[]) => void
  onWarningHover: (nodeIds: string[]) => void
  onDismiss?: (index: number) => void
}

const RULE_COLORS: Record<string, string> = {
  spof: '#ef4444',
  lb_spof: '#f87171',
  db_selection: '#f59e0b',
  read_write_separation: '#d97706',
  federation: '#8b5cf6',
  cache_consistency: '#3b82f6',
  cache_eviction: '#60a5fa',
  cap_theorem: '#ec4899',
  cdn_usage: '#10b981',
  async_decoupling: '#84cc16',
  schema: '#6b7280',
}

function WarningsPanel({
  warnings,
  onWarningClick,
  onWarningHover,
  onDismiss,
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
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead style={{ position: 'sticky', top: 0, backgroundColor: '#fef3c7', zIndex: 1 }}>
              <tr style={{ textAlign: 'left', fontSize: 11, color: '#92400e', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                <th style={{ padding: '8px 16px', fontWeight: 600 }}>警告訊息</th>
                <th style={{ padding: '8px 16px', fontWeight: 600 }}>建議解決方式</th>
              </tr>
            </thead>
            <tbody>
              {warnings.map((w, i) => (
                <tr
                  key={i}
                  onClick={() => onWarningClick(w.nodeIds)}
                  onMouseEnter={() => onWarningHover(w.nodeIds)}
                  onMouseLeave={() => onWarningHover([])}
                  style={{
                    fontSize: 13,
                    lineHeight: 1.5,
                    cursor: 'pointer',
                    borderBottom: '1px solid #fef3c7',
                    verticalAlign: 'top',
                  }}
                >
                  <td style={{ padding: '8px 16px', position: 'relative' }}>
                    {onDismiss && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          onDismiss(i)
                        }}
                        style={{
                          position: 'absolute',
                          top: 4,
                          left: 4,
                          background: 'none',
                          border: 'none',
                          color: '#b45309',
                          cursor: 'pointer',
                          fontSize: 14,
                          padding: 0,
                          lineHeight: 1,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          zIndex: 5,
                        }}
                        title="忽略警告"
                      >
                        ×
                      </button>
                    )}
                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8, paddingLeft: onDismiss ? 12 : 0 }}>
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
                  </td>
                  <td style={{ padding: '8px 16px', color: '#b45309', fontStyle: 'italic' }}>
                    {w.solution}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

export default WarningsPanel
