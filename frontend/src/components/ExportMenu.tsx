import { useState, useRef, useEffect, useCallback } from 'react'
import type { Node, Edge } from '@xyflow/react'
import { copyMermaidToClipboard, downloadMermaidFile } from '../utils/exportMermaid'
import { exportPng } from '../utils/exportImage'
import { exportPdf } from '../utils/exportPdf'
import { toExcalidraw, downloadExcalidraw } from '../utils/exportExcalidraw'

interface ExportMenuProps {
  nodes: Node[]
  edges: Edge[]
  isDarkMode: boolean
}

type ExportStatus = 'idle' | 'exporting' | 'success' | 'error'

export default function ExportMenu({ nodes, edges, isDarkMode }: ExportMenuProps) {
  const [isOpen, setIsOpen] = useState(false)
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

  const handleMermaidCopy = useCallback(async () => {
    try {
      showStatus('Copying...', 'exporting')
      await copyMermaidToClipboard(nodes, edges)
      showStatus('Copied to clipboard', 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`Copy failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [nodes, edges, showStatus])

  const handleMermaidFile = useCallback(() => {
    try {
      showStatus('Downloading...', 'exporting')
      downloadMermaidFile(nodes, edges)
      showStatus('Downloaded .mmd', 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`Download failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [nodes, edges, showStatus])

  const handlePng = useCallback(async () => {
    try {
      showStatus('Generating PNG...', 'exporting')
      await exportPng(nodes, isDarkMode)
      showStatus('PNG exported', 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`PNG failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [nodes, isDarkMode, showStatus])

  const handlePdf = useCallback(async () => {
    try {
      showStatus('Generating PDF...', 'exporting')
      await exportPdf(nodes, isDarkMode)
      showStatus('PDF exported', 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`PDF failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [nodes, isDarkMode, showStatus])

  const handleExcalidraw = useCallback(() => {
    try {
      showStatus('Downloading...', 'exporting')
      const content = toExcalidraw(nodes, edges)
      downloadExcalidraw(content)
      showStatus('Downloaded .excalidraw', 'success')
      setIsOpen(false)
    } catch (err) {
      showStatus(`Export failed: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error')
    }
  }, [nodes, edges, showStatus])

  const disabled = nodes.length === 0

  const buttonStyle: React.CSSProperties = {
    display: 'block',
    width: '100%',
    padding: '8px 14px',
    border: 'none',
    backgroundColor: 'transparent',
    color: 'var(--text-primary)',
    fontSize: 13,
    textAlign: 'left',
    cursor: 'pointer',
    borderRadius: 4,
  }

  return (
    <div ref={menuRef} style={{ position: 'relative' }}>
      <button
        onClick={() => setIsOpen((prev) => !prev)}
        disabled={disabled}
        onMouseEnter={(e) => {
          if (!disabled) e.currentTarget.style.backgroundColor = 'var(--btn-active-bg)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = disabled ? 'var(--btn-disabled-bg)' : 'var(--bg-secondary)'
        }}
        style={{
          padding: '6px 12px',
          borderRadius: 6,
          border: '1px solid var(--border-color)',
          backgroundColor: disabled ? 'var(--btn-disabled-bg)' : 'var(--bg-secondary)',
          color: 'var(--text-primary)',
          fontSize: 13,
          fontWeight: 500,
          cursor: disabled ? 'not-allowed' : 'pointer',
        }}
        title="Export diagram"
      >
        Export
      </button>

      {isOpen && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            marginTop: 4,
            minWidth: 180,
            backgroundColor: 'var(--bg-primary)',
            border: '1px solid var(--border-color)',
            borderRadius: 8,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            zIndex: 50,
            padding: 4,
          }}
        >
          <button
            onClick={handleMermaidCopy}
            style={buttonStyle}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
          >
            Mermaid (Copy)
          </button>
          <button
            onClick={handleMermaidFile}
            style={buttonStyle}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
          >
            Mermaid (.mmd)
          </button>
          <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
          <button
            onClick={handlePng}
            style={buttonStyle}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
          >
            PNG Image
          </button>
          <button
            onClick={handlePdf}
            style={buttonStyle}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
          >
            PDF Document
          </button>
          <div style={{ height: 1, backgroundColor: 'var(--border-color)', margin: '4px 0' }} />
          <button
            onClick={handleExcalidraw}
            style={buttonStyle}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)' }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
          >
            Excalidraw (.excalidraw)
          </button>
        </div>
      )}

      {status !== 'idle' && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: isOpen ? 190 : 0,
            marginTop: 4,
            padding: '6px 12px',
            borderRadius: 6,
            fontSize: 12,
            fontWeight: 500,
            whiteSpace: 'nowrap',
            backgroundColor: status === 'error' ? '#ef4444' : status === 'success' ? '#10b981' : 'var(--bg-tertiary)',
            color: status === 'error' || status === 'success' ? '#ffffff' : 'var(--text-primary)',
            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          }}
        >
          {statusMessage}
        </div>
      )}
    </div>
  )
}
