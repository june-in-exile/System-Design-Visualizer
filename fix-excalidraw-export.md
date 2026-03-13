# ArchitectMind — 修復 Excalidraw 匯出問題

## 問題清單

| # | 問題 | 原因 |
|---|---|---|
| 1 | 箭頭起終點位置錯誤（水平連線偏移嚴重） | 一律用「底部中央→頂部中央」，忽略了 `sourceHandle` / `targetHandle` |
| 2 | 箭頭線條太粗太黑 | `strokeWidth: 2` + `strokeColor: '#1e1e1e'` 在 Excalidraw 裡太重 |
| 3 | 節點尺寸不一致 | ArchitectureNode 用 `160×80`，export 用 `180×80` |

---

## 修復：替換整個 `exportExcalidraw.ts`

```typescript
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

const NODE_WIDTH = 160   // ← 修正：原本是 180，改為 160
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
  const side = h.split('-')[0] // 'top' | 'bottom' | 'left' | 'right'

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
    const color = getStrokeColor(componentType)

    const rectId = `rect-${node.id}`
    const textId = `text-${node.id}`

    elements.push({
      id: rectId,
      type: 'rectangle',
      x: node.position.x,
      y: node.position.y,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      strokeColor: color,
      strokeWidth: 1,           // ← 修正：2 → 1
      backgroundColor: `${color}22`,
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
      x: node.position.x + NODE_WIDTH / 2,
      y: node.position.y + NODE_HEIGHT / 2,
      width: NODE_WIDTH - 20,
      height: 24,
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

    // ← 修正：根據 sourceHandle / targetHandle 計算起終點
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
      strokeColor: '#555555',   // ← 修正：#1e1e1e → #555 柔和灰
      strokeWidth: 1,           // ← 修正：2 → 1
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

      // ← 修正：label 位置用實際起終點中點
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
        strokeColor: '#888888',   // ← 修正：#666 → #888 更柔和
        backgroundColor: 'transparent',
        fillStyle: 'solid',
        roughness: 0,
        opacity: 70,              // ← 修正：80 → 70 更淡
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
```

---

## 修改摘要

| 項目 | 舊值 | 新值 | 原因 |
|---|---|---|---|
| `NODE_WIDTH` | 180 | **160** | 與 ArchitectureNode.tsx 一致 |
| 矩形 `strokeWidth` | 2 | **1** | Excalidraw 預設就是 1，看起來更自然 |
| 箭頭 `strokeWidth` | 2 | **1** | 線條太粗，改細 |
| 箭頭 `strokeColor` | `#1e1e1e` | **`#555555`** | 純黑太重，改柔和灰 |
| 標籤 `strokeColor` | `#666666` | **`#888888`** | 更淡，不搶視線 |
| 標籤 `opacity` | 80 | **70** | 更淡 |
| 箭頭起終點 | 一律底部→頂部 | **讀取 `sourceHandle` / `targetHandle`** | 修正水平連線位置 |
| 標籤位置 | 用固定的 sx/tx 中點 | **用實際 anchor 中點** | 跟著箭頭走 |

---

## 新增函數：`getAnchorPoint`

根據 React Flow 的 handle ID（如 `right-source`、`left-target`）計算連接點座標：

| Handle | 座標 |
|---|---|
| `top-*` | 節點頂部中央 `(x + w/2, y)` |
| `bottom-*` | 節點底部中央 `(x + w/2, y + h)` |
| `left-*` | 節點左側中央 `(x, y + h/2)` |
| `right-*` | 節點右側中央 `(x + w, y + h/2)` |
| 預設（無 handle） | 底部中央 |

---

## 測試驗證

載入各 Preset 後匯出，檢查：

| 測試案例 | 重點檢查 |
|---|---|
| **Demo** | Client → DNS（水平箭頭，right → left） |
| **Twitter** | Client → DNS（right → left）、Client → CDN（斜向） |
| **YouTube** | CDN → Storage（底部 → 頂部）、metadata → CDN（top → bottom） |
| **Google** | 多條水平 + 垂直混合連線 |

每個 Preset 都要確認：
- ✅ 箭頭從正確的 handle 方向出發
- ✅ 線條細且柔和
- ✅ 標籤在箭頭中間位置
- ✅ 虛線（async）清晰但不突兀
