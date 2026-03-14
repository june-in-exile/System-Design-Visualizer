import { useState, useCallback, useRef } from 'react'
import type { Node, Edge } from '@xyflow/react'
import type { SystemParams } from '../types/topology'

export interface CanvasTab {
  readonly id: string
  readonly name: string
  readonly nodes: readonly Node[]
  readonly edges: readonly Edge[]
  readonly params: SystemParams
}

let tabCounter = 0
function generateTabId(): string {
  tabCounter += 1
  return `tab-${Date.now()}-${tabCounter}`
}

function createEmptyTab(name: string): CanvasTab {
  return {
    id: generateTabId(),
    name: name,
    nodes: [],
    edges: [],
    params: {},
  }
}

export function useCanvasTabs() {
  const [tabs, setTabs] = useState<CanvasTab[]>(() => {
    return [createEmptyTab('Untitled 1')]
  })
  const [activeTabId, setActiveTabId] = useState<string>(() => tabs[0].id)
  const canvasStateRef = useRef<{ nodes: Node[]; edges: Edge[]; params: SystemParams } | null>(null)

  const activeTab = tabs.find((t) => t.id === activeTabId) ?? tabs[0]

  const saveCurrentCanvasState = useCallback(() => {
    const state = canvasStateRef.current
    if (!state) return
    setTabs((prev) =>
      prev.map((tab) =>
        tab.id === activeTabId
          ? { ...tab, nodes: [...state.nodes], edges: [...state.edges], params: { ...state.params } }
          : tab
      )
    )
  }, [activeTabId])

  const addTab = useCallback(() => {
    saveCurrentCanvasState()
    const nextName = `Untitled ${tabs.length + 1}`
    const newTab = createEmptyTab(nextName)
    setTabs((prev) => [...prev, newTab])
    setActiveTabId(newTab.id)
  }, [saveCurrentCanvasState, tabs.length])

  const switchTab = useCallback(
    (tabId: string) => {
      if (tabId === activeTabId) return
      saveCurrentCanvasState()
      setActiveTabId(tabId)
    },
    [activeTabId, saveCurrentCanvasState]
  )

  const closeTab = useCallback(
    (tabId: string) => {
      setTabs((prev) => {
        if (prev.length <= 1) return prev
        const filtered = prev.filter((t) => t.id !== tabId)
        if (tabId === activeTabId) {
          const closedIndex = prev.findIndex((t) => t.id === tabId)
          const newActive = filtered[Math.min(closedIndex, filtered.length - 1)]
          setActiveTabId(newActive.id)
        }
        return filtered
      })
    },
    [activeTabId]
  )

  const renameTab = useCallback((tabId: string, newName: string) => {
    setTabs((prev) =>
      prev.map((tab) =>
        tab.id === tabId ? { ...tab, name: newName } : tab
      )
    )
  }, [])

  const updateCanvasStateRef = useCallback(
    (nodes: Node[], edges: Edge[], params?: SystemParams) => {
      canvasStateRef.current = { nodes, edges, params: params ?? {} }
    },
    []
  )

  return {
    tabs,
    activeTab,
    activeTabId,
    addTab,
    switchTab,
    closeTab,
    renameTab,
    updateCanvasStateRef,
  }
}
