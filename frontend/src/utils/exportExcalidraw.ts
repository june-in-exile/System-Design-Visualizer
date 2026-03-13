import type { Node, Edge } from '@xyflow/react'

// ============================================================
// 顏色對照表（與 nodeConfig.ts 保持一致）
// ============================================================

const COMPONENT_COLORS: Record<string, string> = {
  client: '#6366f1',
  load_balancer: '#f59e0b',
  api_gateway: '#10b981',
  service: '#3b82f6',
  database: '#ef4444',
  cache: '#8b5cf6',
  message_queue: '#ec4899',
  cdn: '#14b8a6',
  dns: '#64748b',
  storage: '#f97316',
  reverse_proxy: '#0ea5e9',
}

function getStrokeColor(componentType: string): string {
  return COMPONENT_COLORS[componentType] || '#1e1e1e'
}

// ============================================================
// 節點尺寸（與 ArchitectureNode.tsx 一致）
// ============================================================

const NODE_WIDTH = 180
const NODE_HEIGHT = 80

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
    const color = getStrokeColor(componentType)

    const rectId = `rect-${node.id}`
    const textId = `text-${node.id}`

    // 矩形
    elements.push({
      id: rectId,
      type: 'rectangle',
      x: node.position.x,
      y: node.position.y,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      strokeColor: color,
      strokeWidth: 2,
      backgroundColor: `${color}22`,
      fillStyle: 'solid',
      roughness: 1,
      opacity: 100,
      roundness: { type: 3 },
      boundElements: [{ id: textId, type: 'text' }],
      seed: hashString(node.id),
    })

    // 文字標籤（綁定在矩形內）
    elements.push({
      id: textId,
      type: 'text',
      x: node.position.x + NODE_WIDTH / 2,
      y: node.position.y + NODE_HEIGHT / 2,
      width: NODE_WIDTH - 20,
      height: 24,
      text: label,
      fontSize: 16,
      fontFamily: 1,           // 1 = Virgil (手寫), Excalidraw 預設
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

    // 起點：source 底部中央 → 終點：target 頂部中央
    const sx = sourceNode.position.x + NODE_WIDTH / 2
    const sy = sourceNode.position.y + NODE_HEIGHT
    const tx = targetNode.position.x + NODE_WIDTH / 2
    const ty = targetNode.position.y

    // Arrow
    elements.push({
      id: `arrow-${edge.id}`,
      type: 'arrow',
      x: sx,
      y: sy,
      width: Math.abs(tx - sx),
      height: Math.abs(ty - sy),
      points: [[0, 0], [tx - sx, ty - sy]],
      strokeColor: '#1e1e1e',
      strokeWidth: 2,
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

      elements.push({
        id: `label-${edge.id}`,
        type: 'text',
        x: (sx + tx) / 2 - 40,
        y: (sy + ty) / 2 - 12,
        width: 80,
        height: 20,
        text: parts.join(' · '),
        fontSize: 12,
        fontFamily: 1,
        textAlign: 'center',
        strokeColor: '#666666',
        backgroundColor: 'transparent',
        fillStyle: 'solid',
        roughness: 0,
        opacity: 80,
      })
    }
  }

  // ----------------------------------------------------------
  // 組裝 Excalidraw JSON
  // ----------------------------------------------------------
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

/** 穩定的 hash（與 utils/rough.ts 的 stableSeed 同邏輯） */
function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash + str.charCodeAt(i)) | 0
  }
  return Math.abs(hash)
}

/** 觸發瀏覽器下載 */
export function downloadExcalidraw(content: string, filename = 'architecture.excalidraw') {
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
