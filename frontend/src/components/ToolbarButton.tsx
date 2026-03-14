import { useState, useRef, useCallback, type ReactNode, type CSSProperties } from 'react'

interface ToolbarButtonProps {
  readonly label: ReactNode
  readonly shortcut?: string
  readonly onClick: () => void
  readonly disabled?: boolean
  readonly title?: string
  readonly style?: CSSProperties
}

export default function ToolbarButton({
  label,
  shortcut,
  onClick,
  disabled = false,
  title,
  style,
}: ToolbarButtonProps) {
  const [hovered, setHovered] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [showTooltip, setShowTooltip] = useState(false)

  const handleMouseEnter = useCallback(() => {
    setHovered(true)
    if (shortcut) {
      timerRef.current = setTimeout(() => setShowTooltip(true), 400)
    }
  }, [shortcut])

  const handleMouseLeave = useCallback(() => {
    setHovered(false)
    setShowTooltip(false)
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
  }, [])

  return (
    <div style={{ position: 'relative', display: 'inline-block' }}>
      <button
        onClick={onClick}
        disabled={disabled}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        style={{
          padding: '6px 12px',
          borderRadius: 6,
          border: '1px solid var(--border-color)',
          backgroundColor: disabled
            ? 'var(--btn-disabled-bg)'
            : hovered
              ? 'var(--btn-active-bg)'
              : 'var(--bg-secondary)',
          color: disabled ? 'var(--text-secondary)' : 'var(--text-primary)',
          fontSize: 16,
          fontWeight: 600,
          cursor: disabled ? 'not-allowed' : 'pointer',
          ...style,
        }}
        title={title}
      >
        {label}
      </button>
      {showTooltip && shortcut && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: '50%',
            transform: 'translateX(-50%)',
            marginTop: 6,
            padding: '4px 8px',
            borderRadius: 4,
            backgroundColor: 'var(--bg-primary)',
            border: '1px solid var(--border-color)',
            boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
            fontSize: 11,
            fontWeight: 500,
            color: 'var(--text-secondary)',
            whiteSpace: 'nowrap',
            zIndex: 50,
            pointerEvents: 'none',
          }}
        >
          {shortcut}
        </div>
      )}
    </div>
  )
}
