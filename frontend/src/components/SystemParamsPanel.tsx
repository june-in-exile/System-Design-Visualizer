import { useState, useRef, useEffect, useCallback } from 'react'
import type { SystemParams } from '../types/topology'

interface SystemParamsPanelProps {
  params: SystemParams
  onChange: (params: SystemParams) => void
}

const AVAILABILITY_OPTIONS = ['99%', '99.9%', '99.99%', '99.999%']
const LATENCY_OPTIONS = ['p99 < 500ms', 'p99 < 200ms', 'p99 < 100ms', 'p95 < 50ms']

export default function SystemParamsPanel({ params, onChange }: SystemParamsPanelProps) {
  const [open, setOpen] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const handleChange = useCallback((key: keyof SystemParams, value: unknown) => {
    onChange({ ...params, [key]: value })
  }, [params, onChange])

  const handleNumberChange = useCallback((key: keyof SystemParams, raw: string) => {
    if (raw === '') {
      const next = { ...params }
      delete next[key]
      onChange(next)
    } else {
      const num = Number(raw)
      if (!isNaN(num)) onChange({ ...params, [key]: num })
    }
  }, [params, onChange])

  const estimatedPeakQPS = params.dau ? Math.ceil(params.dau / 86400 * 10) : null

  const hasParams = Object.values(params).some(v => v !== undefined && v !== '' && v !== 0)

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '6px 8px',
    border: '1px solid var(--border-color)',
    borderRadius: 4,
    fontSize: 13,
    backgroundColor: 'var(--bg-primary)',
    color: 'var(--text-primary)',
  }

  const labelStyle: React.CSSProperties = {
    display: 'block',
    marginBottom: 4,
    fontSize: 12,
    fontWeight: 500,
    color: 'var(--text-secondary)',
  }

  return (
    <div ref={panelRef} style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen(!open)}
        style={{
          padding: '6px 12px',
          borderRadius: 6,
          border: '1px solid var(--border-color)',
          backgroundColor: open ? 'var(--btn-active-bg)' : 'var(--bg-secondary)',
          color: 'var(--text-primary)',
          fontSize: 13,
          fontWeight: 500,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: 4,
        }}
        title="Set system parameters (DAU, QPS, etc.)"
      >
        Params
        {hasParams && (
          <span style={{
            width: 6,
            height: 6,
            borderRadius: '50%',
            backgroundColor: 'var(--accent)',
            display: 'inline-block',
          }} />
        )}
      </button>

      {open && (
        <div style={{
          position: 'absolute',
          top: '100%',
          left: 0,
          marginTop: 8,
          width: 320,
          padding: 16,
          borderRadius: 8,
          border: '1px solid var(--border-color)',
          backgroundColor: 'var(--bg-secondary)',
          boxShadow: '0 8px 24px rgba(0,0,0,0.15)',
          zIndex: 100,
          maxHeight: '70vh',
          overflowY: 'auto',
        }}>
          <h4 style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>
            System Parameters
          </h4>
          <p style={{ margin: '0 0 16px', fontSize: 11, color: 'var(--text-secondary)', lineHeight: 1.4 }}>
            設定系統容量參數，Analyze 時會根據這些參數產生更精準的建議。所有欄位皆為選填。
          </p>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>DAU (日活躍用戶)</label>
            <input
              type="text"
              inputMode="numeric"
              placeholder="e.g., 1000000"
              value={params.dau ?? ''}
              onChange={(e) => handleNumberChange('dau', e.target.value)}
              style={inputStyle}
            />
            {estimatedPeakQPS && (
              <div style={{ fontSize: 11, color: 'var(--accent)', marginTop: 4 }}>
                💡 預估 Peak QPS ≈ {estimatedPeakQPS.toLocaleString()}（DAU/86400 × 10）
              </div>
            )}
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Peak QPS (尖峰每秒請求)</label>
            <input
              type="text"
              inputMode="numeric"
              placeholder={estimatedPeakQPS ? `建議 ≥ ${estimatedPeakQPS}` : 'e.g., 10000'}
              value={params.peakQPS ?? ''}
              onChange={(e) => handleNumberChange('peakQPS', e.target.value)}
              style={inputStyle}
            />
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Avg QPS (平均每秒請求)</label>
            <input
              type="text"
              inputMode="numeric"
              placeholder="e.g., 2000"
              value={params.avgQPS ?? ''}
              onChange={(e) => handleNumberChange('avgQPS', e.target.value)}
              style={inputStyle}
            />
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Storage (GB)</label>
            <input
              type="text"
              inputMode="numeric"
              placeholder="e.g., 500"
              value={params.storageGB ?? ''}
              onChange={(e) => handleNumberChange('storageGB', e.target.value)}
              style={inputStyle}
            />
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Daily Growth (GB/天)</label>
            <input
              type="text"
              inputMode="numeric"
              placeholder="e.g., 10"
              value={params.dailyGrowthGB ?? ''}
              onChange={(e) => handleNumberChange('dailyGrowthGB', e.target.value)}
              style={inputStyle}
            />
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>
              Read/Write Ratio：{params.readWriteRatio !== undefined ? `${Math.round(params.readWriteRatio * 100)}% 讀` : '未設定'}
            </label>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              value={params.readWriteRatio ?? 0.8}
              onChange={(e) => handleChange('readWriteRatio', parseFloat(e.target.value))}
              style={{ width: '100%' }}
            />
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10, color: 'var(--text-secondary)' }}>
              <span>寫多</span>
              <span>讀多</span>
            </div>
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Latency Target</label>
            <select
              value={params.latencyTarget ?? ''}
              onChange={(e) => handleChange('latencyTarget', e.target.value || undefined)}
              style={inputStyle}
            >
              <option value="">(未設定)</option>
              {LATENCY_OPTIONS.map(opt => (
                <option key={opt} value={opt}>{opt}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 14 }}>
            <label style={labelStyle}>Availability Target</label>
            <select
              value={params.availability ?? ''}
              onChange={(e) => handleChange('availability', e.target.value || undefined)}
              style={inputStyle}
            >
              <option value="">(未設定)</option>
              {AVAILABILITY_OPTIONS.map(opt => (
                <option key={opt} value={opt}>{opt}</option>
              ))}
            </select>
          </div>

          {hasParams && (
            <button
              onClick={() => onChange({})}
              style={{
                width: '100%',
                padding: '6px',
                borderRadius: 4,
                border: '1px solid var(--border-color)',
                backgroundColor: 'transparent',
                color: 'var(--text-secondary)',
                fontSize: 12,
                cursor: 'pointer',
                marginTop: 4,
              }}
            >
              清除所有參數
            </button>
          )}
        </div>
      )}
    </div>
  )
}
