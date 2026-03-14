import { useState, useRef, useEffect, useCallback } from 'react'
import type { Node, Edge } from '@xyflow/react'
import { copyMermaidToClipboard, downloadMermaidFile } from '../utils/exportMermaid'
import { exportPng } from '../utils/exportImage'
import { exportPdf } from '../utils/exportPdf'
import { toExcalidraw, downloadExcalidraw } from '../utils/exportExcalidraw'

interface SettingsMenuProps {
  theme: 'light' | 'dark' | 'warm' | 'dream'
  setTheme: (theme: 'light' | 'dark' | 'warm' | 'dream') => void
  getNodes: () => Node[]
  getEdges: () => Edge[]
}

type ExportStatus = 'idle' | 'exporting' | 'success' | 'error'
type MenuView = 'main' | 'export' | 'theme'

export default function SettingsMenu({ theme, setTheme, getNodes, getEdges }: SettingsMenuProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [view, setView] = useState<MenuView>('main')
  const [status, setStatus] = useState<ExportStatus>('idle')
  const [statusMessage, setStatusMessage] = useState('')
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as HTMLElement)) {
        setIsOpen(false)
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  const effectiveView = isOpen ? view : 'main'

  const showStatus = useCallback((message: string, type: ExportStatus) => {
    setStatus(type)
    setStatusMessage(message)
    if (type === 'success' || type === 'error') {
      setTimeout(() => {
        setStatus('idle')
        setStatusMessage('')
      }, 2000)
    }
  }, [])

  const handleExport = useCallback(async (action: () => Promise<unknown> | void, successMsg: string, errorMsgPrefix: string) => {
    try {
      showStatus('Processing...', 'exporting')
      await action()
      showStatus(successMsg, 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`${errorMsgPrefix}: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [showStatus])

  const itemStyle = (disabled?: boolean): React.CSSProperties => ({
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
    padding: '10px 12px',
    border: 'none',
    backgroundColor: 'transparent',
    color: disabled ? 'var(--text-secondary)' : 'var(--text-primary)',
    fontSize: 13,
    textAlign: 'left',
    cursor: disabled ? 'not-allowed' : 'pointer',
    borderRadius: 6,
    transition: 'background-color 0.2s',
  })

  const renderMainView = () => {
    const hasNodes = getNodes().length > 0
    const getThemeLabel = () => {
      if (theme === 'light') return 'Light'
      if (theme === 'dark') return 'Dark'
      if (theme === 'warm') return 'Warm'
      return 'Dream'
    }

    return (
    <>
      <button
        onClick={() => setView('theme')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Theme ({getThemeLabel()})</span>
        <span style={{ color: 'var(--text-secondary)' }}>→</span>
      </button>
      <button
        onClick={() => setView('export')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Export</span>
        <span style={{ color: hasNodes ? 'var(--text-secondary)' : 'var(--text-disabled)' }}>→</span>
      </button>
    </>
  )}

  const renderThemeView = () => (
    <>
      <button
        onClick={() => setView('main')}
        style={{ ...itemStyle(), color: 'var(--text-secondary)', marginBottom: 4 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>← Back</span>
      </button>
      <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
      <button
        onClick={() => setTheme('light')}
        style={{ ...itemStyle(), fontWeight: theme === 'light' ? 700 : 400 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Light Mode</span>
        {theme === 'light' && <span>✓</span>}
      </button>
      <button
        onClick={() => setTheme('dark')}
        style={{ ...itemStyle(), fontWeight: theme === 'dark' ? 700 : 400 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Dark Mode</span>
        {theme === 'dark' && <span>✓</span>}
      </button>
      <button
        onClick={() => setTheme('warm')}
        style={{ ...itemStyle(), fontWeight: theme === 'warm' ? 700 : 400 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Warm Mode</span>
        {theme === 'warm' && <span>✓</span>}
      </button>
      <button
        onClick={() => setTheme('dream')}
        style={{ ...itemStyle(), fontWeight: theme === 'dream' ? 700 : 400 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>Dream Mode</span>
        {theme === 'dream' && <span>✓</span>}
      </button>
    </>
  )

  const renderExportView = () => (
    <>
      <button
        onClick={() => setView('main')}
        style={{ ...itemStyle, color: 'var(--text-secondary)', marginBottom: 4 }}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        <span>← Back</span>
      </button>
      <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
      <button
        onClick={() => handleExport(() => copyMermaidToClipboard(getNodes(), getEdges()), 'Copied to clipboard', 'Copy failed')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        Mermaid (Copy)
      </button>
      <button
        onClick={() => handleExport(() => downloadMermaidFile(getNodes(), getEdges()), 'Downloaded .mmd', 'Download failed')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        Mermaid (.mmd)
      </button>
      <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
      <button
        onClick={() => handleExport(() => exportPng(getNodes(), theme), 'PNG exported', 'PNG failed')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        PNG Image
      </button>
      <button
        onClick={() => handleExport(() => exportPdf(getNodes(), theme), 'PDF exported', 'PDF failed')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        PDF Document
      </button>
      <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
      <button
        onClick={() => handleExport(() => {
          const content = toExcalidraw(getNodes(), getEdges())
          downloadExcalidraw(content)
        }, 'Downloaded .excalidraw', 'Export failed')}
        style={itemStyle()}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
      >
        Excalidraw (.excalidraw)
      </button>
    </>
  )

  return (
    <div ref={menuRef} style={{ position: 'relative', marginTop: 'auto' }}>
      <button
        onClick={() => setIsOpen((prev) => !prev)}
        onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--btn-active-bg)' }}
        onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-primary)' }}
        style={{
          width: '100%',
          padding: '8px 12px',
          borderRadius: 6,
          border: '1px solid var(--border-color)',
          backgroundColor: 'var(--bg-primary)',
          color: 'var(--text-primary)',
          fontSize: 13,
          fontWeight: 500,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 8,
          transition: 'background-color 0.2s',
        }}
      >
        Settings
      </button>

      {isOpen && (
        <div
          style={{
            position: 'absolute',
            bottom: '100%',
            left: 0,
            marginBottom: 8,
            minWidth: 200,
            backgroundColor: 'var(--bg-primary)',
            border: '1px solid var(--border-color)',
            borderRadius: 8,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            zIndex: 100,
            padding: 4,
          }}
        >
          {effectiveView === 'main' ? renderMainView() : effectiveView === 'export' ? renderExportView() : renderThemeView()}
        </div>
      )}

      {status !== 'idle' && (
        <div
          style={{
            position: 'absolute',
            bottom: '100%',
            left: 210,
            marginBottom: 8,
            padding: '6px 12px',
            borderRadius: 6,
            fontSize: 12,
            fontWeight: 500,
            whiteSpace: 'nowrap',
            backgroundColor: status === 'error' ? '#ef4444' : status === 'success' ? '#10b981' : 'var(--bg-tertiary)',
            color: status === 'error' || status === 'success' ? '#ffffff' : 'var(--text-primary)',
            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
            zIndex: 101,
          }}
        >
          {statusMessage}
        </div>
      )}
    </div>
  )
}
