import { useState, useCallback, useRef, useEffect } from 'react'
import type { CanvasTab } from '../hooks/useCanvasTabs'

interface TabBarProps {
  readonly tabs: readonly CanvasTab[]
  readonly activeTabId: string
  readonly onSwitchTab: (tabId: string) => void
  readonly onAddTab: () => void
  readonly onCloseTab: (tabId: string) => void
  readonly onRenameTab: (tabId: string, newName: string) => void
  readonly isSidebarOpen: boolean
  readonly onToggleSidebar: () => void
}

function TabBar({
  tabs,
  activeTabId,
  onSwitchTab,
  onAddTab,
  onCloseTab,
  onRenameTab,
  isSidebarOpen,
  onToggleSidebar,
}: TabBarProps) {
  const [editingTabId, setEditingTabId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (editingTabId && inputRef.current) {
      inputRef.current.focus()
      inputRef.current.select()
    }
  }, [editingTabId])

  const startEditing = useCallback((tab: CanvasTab) => {
    setEditingTabId(tab.id)
    setEditValue(tab.name)
  }, [])

  const commitEdit = useCallback(() => {
    if (editingTabId && editValue.trim()) {
      onRenameTab(editingTabId, editValue.trim())
    }
    setEditingTabId(null)
  }, [editingTabId, editValue, onRenameTab])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        commitEdit()
      } else if (e.key === 'Escape') {
        setEditingTabId(null)
      }
    },
    [commitEdit]
  )

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        backgroundColor: 'var(--bg-primary)',
        borderBottom: '1px solid var(--border-color)',
        height: 36,
        paddingLeft: 4,
        gap: 0,
        overflowX: 'auto',
        overflowY: 'hidden',
        flexShrink: 0,
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      }}
    >
      <button
        onClick={onToggleSidebar}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: 32,
          height: 28,
          border: 'none',
          background: 'none',
          color: 'var(--text-secondary)',
          cursor: 'pointer',
          borderRadius: 4,
          fontSize: 18,
          lineHeight: 1,
          marginRight: 4,
          flexShrink: 0,
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.backgroundColor = 'var(--bg-secondary)'
          e.currentTarget.style.color = 'var(--text-primary)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = 'transparent'
          e.currentTarget.style.color = 'var(--text-secondary)'
        }}
        title={isSidebarOpen ? "Close sidebar" : "Open sidebar"}
      >
        {isSidebarOpen ? '⇠' : '⇢'}
      </button>

      <div style={{ height: 20, width: 1, backgroundColor: 'var(--border-color)', marginRight: 4 }} />

      {tabs.map((tab) => {
        const isActive = tab.id === activeTabId
        const isEditing = tab.id === editingTabId

        return (
          <div
            key={tab.id}
            onClick={() => onSwitchTab(tab.id)}
            onDoubleClick={() => startEditing(tab)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '0 12px',
              height: '100%',
              fontSize: 12,
              fontWeight: isActive ? 600 : 400,
              color: isActive ? 'var(--text-primary)' : 'var(--text-secondary)',
              backgroundColor: isActive ? 'var(--bg-primary)' : 'transparent',
              borderRight: '1px solid var(--border-color)',
              cursor: 'pointer',
              userSelect: 'none',
              whiteSpace: 'nowrap',
              position: 'relative',
              borderTop: isActive ? '2px solid var(--accent)' : '2px solid transparent',
            }}
          >
            {isEditing ? (
              <input
                ref={inputRef}
                value={editValue}
                onChange={(e) => setEditValue(e.target.value)}
                onBlur={commitEdit}
                onKeyDown={handleKeyDown}
                style={{
                  width: Math.max(60, editValue.length * 7 + 16),
                  height: 22,
                  fontSize: 12,
                  border: '1px solid var(--accent)',
                  borderRadius: 3,
                  padding: '0 4px',
                  outline: 'none',
                  backgroundColor: 'var(--bg-primary)',
                  color: 'var(--text-primary)',
                }}
              />
            ) : (
              <span>{tab.name}</span>
            )}
            {tabs.length > 1 && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onCloseTab(tab.id)
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 16,
                  height: 16,
                  border: 'none',
                  background: 'none',
                  color: 'var(--text-secondary)',
                  cursor: 'pointer',
                  borderRadius: 3,
                  fontSize: 14,
                  lineHeight: 1,
                  padding: 0,
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = 'rgba(255,0,0,0.15)'
                  e.currentTarget.style.color = '#ef4444'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = 'transparent'
                  e.currentTarget.style.color = 'var(--text-secondary)'
                }}
                title="Close tab"
              >
                ×
              </button>
            )}
          </div>
        )
      })}
      <button
        onClick={onAddTab}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: 28,
          height: 28,
          border: 'none',
          background: 'none',
          color: 'var(--text-secondary)',
          cursor: 'pointer',
          borderRadius: 4,
          fontSize: 18,
          lineHeight: 1,
          marginLeft: 4,
          flexShrink: 0,
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.backgroundColor = 'var(--bg-primary)'
          e.currentTarget.style.color = 'var(--text-primary)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = 'transparent'
          e.currentTarget.style.color = 'var(--text-secondary)'
        }}
        title="New canvas"
      >
        +
      </button>
    </div>
  )
}

export default TabBar
