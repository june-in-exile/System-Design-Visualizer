import type { Node, Edge } from '@xyflow/react'

// ============================================================
// 中性配色（與 nodeConfig.ts 保持一致）
// 只有警告狀態才使用黃色，一般情況下統一使用中性灰
// ============================================================

const NEUTRAL_COLOR = '#4b5563'
const WARNING_COLOR = '#eab308'

function getStrokeColor(_componentType: string, hasWarning = false): string {
  return hasWarning ? WARNING_COLOR : NEUTRAL_COLOR
}

// ============================================================
// 節點尺寸（與 ArchitectureNode.tsx 一致）
// ============================================================

const NODE_WIDTH = 160
const NODE_HEIGHT = 80

// ============================================================
// Handle 位置計算
// ============================================================

interface AnchorPoint {
  x: number
  y: number
}

/**
 * 根據 sourceHandle / targetHandle 字串計算連接點座標
 * handle 格式：'top-source', 'bottom-target', 'left-source', 'right-target' 等
 */
function getAnchorPoint(
  nodeX: number,
  nodeY: number,
  handle: string | null | undefined
): AnchorPoint {
  const h = handle || 'bottom-source'
  const side = h.split('-')[0]

  switch (side) {
    case 'top':
      return { x: nodeX + NODE_WIDTH / 2, y: nodeY }
    case 'bottom':
      return { x: nodeX + NODE_WIDTH / 2, y: nodeY + NODE_HEIGHT }
    case 'left':
      return { x: nodeX, y: nodeY + NODE_HEIGHT / 2 }
    case 'right':
      return { x: nodeX + NODE_WIDTH, y: nodeY + NODE_HEIGHT / 2 }
    default:
      return { x: nodeX + NODE_WIDTH / 2, y: nodeY + NODE_HEIGHT }
  }
}

// ============================================================
// 核心轉換
// ============================================================

export function toExcalidraw(nodes: Node[], edges: Edge[]): string {
  const elements: Record<string, unknown>[] = []

  // ----------------------------------------------------------
  // 節點 → rectangle + text
  // ----------------------------------------------------------
  for (const node of nodes) {
    const data = node.data as Record<string, unknown>
    const label = (data.label as string) || node.id
    const componentType = (data.componentType as string) || 'service'
    const hasWarning = (data.hasWarning as boolean) || false
    const color = getStrokeColor(componentType, hasWarning)

    const rectId = `rect-${node.id}`
    const textId = `text-${node.id}`

    const textWidth = NODE_WIDTH - 20
    const textHeight = 24

    elements.push({
      id: rectId,
      type: 'rectangle',
      x: node.position.x,
      y: node.position.y,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      strokeColor: color,
      strokeWidth: 1,
      backgroundColor: `${color}12`,
      fillStyle: 'solid',
      roughness: 1,
      opacity: 100,
      roundness: { type: 3 },
      boundElements: [{ id: textId, type: 'text' }],
      seed: hashString(node.id),
    })

    elements.push({
      id: textId,
      type: 'text',
      x: node.position.x + (NODE_WIDTH - textWidth) / 2,
      y: node.position.y + (NODE_HEIGHT - textHeight) / 2,
      width: textWidth,
      height: textHeight,
      text: label,
      fontSize: 16,
      fontFamily: 1,
      textAlign: 'center',
      verticalAlign: 'middle',
      strokeColor: color,
      backgroundColor: 'transparent',
      fillStyle: 'solid',
      roughness: 1,
      opacity: 100,
      containerId: rectId,
    })
  }

  // ----------------------------------------------------------
  // 連線 → arrow + label text
  // ----------------------------------------------------------
  for (const edge of edges) {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    if (!sourceNode || !targetNode) continue

    const edgeData = (edge.data as Record<string, unknown>) ?? {}
    const isAsync = edgeData.connectionType === 'async'
    const protocol = (edgeData.protocol as string) || ''
    const connType = (edgeData.connectionType as string) || ''

    const src = getAnchorPoint(
      sourceNode.position.x,
      sourceNode.position.y,
      edge.sourceHandle
    )
    const tgt = getAnchorPoint(
      targetNode.position.x,
      targetNode.position.y,
      edge.targetHandle
    )

    elements.push({
      id: `arrow-${edge.id}`,
      type: 'arrow',
      x: src.x,
      y: src.y,
      width: Math.abs(tgt.x - src.x),
      height: Math.abs(tgt.y - src.y),
      points: [[0, 0], [tgt.x - src.x, tgt.y - src.y]],
      strokeColor: '#555555',
      strokeWidth: 1,
      backgroundColor: 'transparent',
      fillStyle: 'solid',
      roughness: 1,
      opacity: 100,
      strokeStyle: isAsync ? 'dashed' : 'solid',
      startArrowhead: null,
      endArrowhead: 'arrow',
      startBinding: {
        elementId: `rect-${edge.source}`,
        focus: 0,
        gap: 4,
      },
      endBinding: {
        elementId: `rect-${edge.target}`,
        focus: 0,
        gap: 4,
      },
      seed: hashString(edge.id),
    })

    // 連線標籤
    if (protocol || (connType && connType !== 'unspecified')) {
      const parts: string[] = []
      if (protocol) parts.push(protocol.toUpperCase())
      if (connType && connType !== 'unspecified') parts.push(connType)

      const midX = (src.x + tgt.x) / 2
      const midY = (src.y + tgt.y) / 2

      elements.push({
        id: `label-${edge.id}`,
        type: 'text',
        x: midX - 40,
        y: midY - 12,
        width: 80,
        height: 20,
        text: parts.join(' · '),
        fontSize: 12,
        fontFamily: 1,
        textAlign: 'center',
        strokeColor: '#888888',
        backgroundColor: 'transparent',
        fillStyle: 'solid',
        roughness: 0,
        opacity: 70,
      })
    }
  }

  return JSON.stringify({
    type: 'excalidraw',
    version: 2,
    source: 'ArchitectMind',
    elements,
    appState: {
      viewBackgroundColor: '#fafaf8',
      gridSize: 20,
      theme: 'light',
    },
    files: {},
  }, null, 2)
}

// ============================================================
// 工具函數
// ============================================================

function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash + str.charCodeAt(i)) | 0
  }
  return Math.abs(hash)
}

export function downloadExcalidraw(content: string, filename = 'architecture.excalidraw') {
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
