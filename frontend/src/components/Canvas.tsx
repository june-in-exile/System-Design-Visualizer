import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  type OnConnect,
  type Node,
  type Edge,
  type ReactFlowInstance,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import ArchitectureNode from '../nodes/ArchitectureNode'
import PropertyPanel from './PropertyPanel'
import { NODE_TYPE_CONFIG } from '../nodes/nodeConfig'
import { analyzeTopology } from '../api/topologyApi'
import type {
  ComponentType,
  SystemTopology,
  AnalyzeResponse,
  Warning,
} from '../types/topology'

const nodeTypes = {
  architecture: ArchitectureNode,
}

let nodeIdCounter = 0
function generateNodeId(): string {
  nodeIdCounter += 1
  return `node-${nodeIdCounter}`
}

interface CanvasProps {
  isDarkMode: boolean;
}

function Canvas({ isDarkMode }: CanvasProps) {
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [analysisResult, setAnalysisResult] = useState<AnalyzeResponse | null>(
    null
  )
  const [analyzing, setAnalyzing] = useState(false)
  const [showWarnings, setShowWarnings] = useState(false)

  const historyRef = useRef<{ nodes: Node[]; edges: Edge[] }[]>([])
  const isUndoRedoRef = useRef(false)
  const prevNodesRef = useRef<Node[]>([])
  const prevEdgesRef = useRef<Edge[]>([])

  const MAX_HISTORY = 10

  const buildWarningsByNode = useCallback((warnings: Warning[]): Map<string, Warning[]> => {
    const map = new Map<string, Warning[]>()
    for (const w of warnings) {
      for (const nodeId of w.nodeIds) {
        const existing = map.get(nodeId) ?? []
        map.set(nodeId, [...existing, w])
      }
    }
    return map
  }, [])

  const warningsByNode = useMemo(
    () => buildWarningsByNode(analysisResult?.warnings ?? []),
    [analysisResult, buildWarningsByNode]
  )

  const fitViewToNode = useCallback((nodeId: string) => {
    if (!rfInstance) return
    const node = nodes.find(n => n.id === nodeId)
    if (!node) return
    
    rfInstance.fitView({
      nodes: [{ id: nodeId, position: node.position }],
      padding: 0.5,
      duration: 300,
    })
    setSelectedNodeId(nodeId)
  }, [rfInstance, nodes])

  useEffect(() => {
    if (isUndoRedoRef.current) return
    
    const currentNodeIds = new Set(nodes.map(n => n.id))
    const prevNodeIds = new Set(prevNodesRef.current.map(n => n.id))
    const currentEdgeIds = new Set(edges.map(e => e.id))
    const prevEdgeIds = new Set(prevEdgesRef.current.map(e => e.id))
    
    const nodeAdded = nodes.length > prevNodesRef.current.length
    const nodeRemoved = prevNodeIds.size > currentNodeIds.size
    const edgeAdded = edges.length > prevEdgesRef.current.length
    const edgeRemoved = prevEdgeIds.size > currentEdgeIds.size
    
    if (nodeAdded || nodeRemoved || edgeAdded || edgeRemoved) {
      historyRef.current.push({ 
        nodes: [...prevNodesRef.current], 
        edges: [...prevEdgesRef.current] 
      })
      if (historyRef.current.length > MAX_HISTORY) {
        historyRef.current.shift()
      }
    }
    
    prevNodesRef.current = [...nodes]
    prevEdgesRef.current = [...edges]
  }, [nodes, edges])

  const pushHistory = useCallback(() => {
    if (isUndoRedoRef.current) return
    historyRef.current.push({ nodes: [...nodes], edges: [...edges] })
    if (historyRef.current.length > MAX_HISTORY) {
      historyRef.current.shift()
    }
  }, [nodes, edges])

  const undo = useCallback(() => {
    if (historyRef.current.length === 0) return
    
    isUndoRedoRef.current = true
    const previousState = historyRef.current.pop()
    if (previousState) {
      setNodes(previousState.nodes)
      setEdges(previousState.edges)
    }
    setTimeout(() => { isUndoRedoRef.current = false }, 0)
  }, [setNodes, setEdges])

  // Close warnings panel if new analysis returns empty or if we re-analyze
  useEffect(() => {
    if (analyzing || !analysisResult?.warnings?.length) {
      setShowWarnings(false)
    }
  }, [analyzing, analysisResult])

  const selectedNode = nodes.find((n) => n.id === selectedNodeId) || null

  const duplicateSelectedNode = useCallback(() => {
    if (!selectedNode) return
    
    pushHistory()
    const data = selectedNode.data as Record<string, unknown>
    const newNode: Node = {
      id: generateNodeId(),
      type: 'architecture',
      position: {
        x: selectedNode.position.x + 50,
        y: selectedNode.position.y + 50,
      },
      data: {
        label: data.label as string,
        componentType: data.componentType as ComponentType,
        properties: { ...(data.properties as Record<string, unknown>) },
      },
    }
    
    setNodes((nds) => [...nds, newNode])
    setSelectedNodeId(newNode.id)
  }, [selectedNode, setNodes])

  const onNodeDataChange = useCallback(
    (nodeId: string, newData: Record<string, unknown>) => {
      pushHistory()
      setNodes((nds) =>
        nds.map((node) =>
          node.id === nodeId ? { ...node, data: newData } : node
        )
      )
    },
    [setNodes, pushHistory]
  )

  const onSelectionChange = useCallback(({ nodes: selectedNodes }: { nodes: Node[] }) => {
    setSelectedNodeId(selectedNodes.length > 0 ? selectedNodes[0].id : null)
  }, [])

  const onConnect: OnConnect = useCallback(
    (params) => {
      if (!params.source || !params.target) return
      pushHistory()
      const newEdge: Edge = {
        id: `edge-${params.source}-${params.target}`,
        source: params.source,
        target: params.target,
        sourceHandle: params.sourceHandle,
        targetHandle: params.targetHandle,
        data: { connectionType: 'sync' },
        style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 },
      }
      setEdges((eds) => [...eds, newEdge])
    },
    [setEdges, isDarkMode, pushHistory]
  )

  const onNodesDelete = useCallback(
    (deletedNodes: Node[]) => {
      const deletedIds = new Set(deletedNodes.map((n) => n.id))
      setEdges((eds) =>
        eds.filter((e) => !deletedIds.has(e.source) && !deletedIds.has(e.target))
      )
      if (selectedNodeId && deletedIds.has(selectedNodeId)) {
        setSelectedNodeId(null)
      }
    },
    [setEdges, selectedNodeId]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const componentType = event.dataTransfer.getData(
        'application/architectmind'
      ) as ComponentType

      if (!componentType || !NODE_TYPE_CONFIG[componentType]) return
      if (!rfInstance || !reactFlowWrapper.current) return

      const bounds = reactFlowWrapper.current.getBoundingClientRect()
      const position = rfInstance.screenToFlowPosition({
        x: event.clientX - bounds.left,
        y: event.clientY - bounds.top,
      })

      pushHistory()
      const config = NODE_TYPE_CONFIG[componentType]
      const newNode: Node = {
        id: generateNodeId(),
        type: 'architecture',
        position,
        data: {
          label: config.label,
          componentType,
          properties: { ...config.defaultProperties },
        },
      }

      setNodes((nds) => [...nds, newNode])
    },
    [rfInstance, setNodes, pushHistory]
  )

  const handleAnalyze = useCallback(async () => {
    if (nodes.length === 0) return

    setAnalyzing(true)
    try {
      const topology: SystemTopology = {
        id: 'current-design',
        name: 'Untitled Design',
        version: 1,
        nodes: nodes.map((n) => {
          const data = n.data as Record<string, unknown>
          return {
            id: n.id,
            componentType: data.componentType as ComponentType,
            label: data.label as string,
            position: n.position,
            properties:
              (data.properties as Record<string, unknown>) ?? {},
          }
        }),
        edges: edges.map((e) => ({
          id: e.id,
          source: e.source,
          target: e.target,
          connectionType:
            ((e.data as Record<string, unknown>)?.connectionType as string ??
            'sync') as SystemTopology['edges'][number]['connectionType'],
        })),
      }

      const result = await analyzeTopology(topology)
      setAnalysisResult(result)
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Analysis failed'
      setAnalysisResult({
        success: false,
        nodeCount: 0,
        edgeCount: 0,
        warnings: [
          { rule: 'error', message, solution: '請檢查後端服務是否正常啟動，並重試。', nodeIds: [] } as Warning,
        ],
      })
    } finally {
      setAnalyzing(false)
    }
  }, [nodes, edges])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'a') {
        e.preventDefault()
        if (nodes.length > 0 && !analyzing) {
          handleAnalyze()
        }
      }
      if ((e.metaKey || e.ctrlKey) && e.key === 'd') {
        e.preventDefault()
        duplicateSelectedNode()
      }
      if ((e.metaKey || e.ctrlKey) && e.key === 'z') {
        e.preventDefault()
        undo()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [nodes.length, analyzing, handleAnalyze, duplicateSelectedNode, undo])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, backgroundColor: 'var(--bg-primary)' }}>
      <div
        style={{
          padding: '8px 16px',
          borderBottom: '1px solid var(--border-color)',
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          backgroundColor: 'var(--bg-primary)'
        }}
      >
        <button
          onClick={handleAnalyze}
          disabled={analyzing || nodes.length === 0}
          style={{
            padding: '6px 16px',
            borderRadius: 6,
            border: 'none',
            backgroundColor: nodes.length === 0 ? 'var(--btn-disabled-bg)' : 'var(--btn-primary-bg)',
            color: 'var(--btn-primary-text)',
            fontSize: 13,
            fontWeight: 600,
            cursor: nodes.length === 0 ? 'not-allowed' : 'pointer',
          }}
        >
          {analyzing ? 'Analyzing...' : 'Analyze'}
        </button>
        <button
          onClick={duplicateSelectedNode}
          disabled={!selectedNode}
          style={{
            padding: '6px 12px',
            borderRadius: 6,
            border: '1px solid var(--border-color)',
            backgroundColor: selectedNode ? 'var(--bg-secondary)' : 'var(--btn-disabled-bg)',
            color: 'var(--text-primary)',
            fontSize: 13,
            fontWeight: 500,
            cursor: selectedNode ? 'pointer' : 'not-allowed',
          }}
          title="Duplicate selected (Cmd+D)"
        >
          Duplicate
        </button>
        <button
          onClick={undo}
          disabled={historyRef.current.length === 0}
          style={{
            padding: '6px 12px',
            borderRadius: 6,
            border: '1px solid var(--border-color)',
            backgroundColor: historyRef.current.length > 0 ? 'var(--bg-secondary)' : 'var(--btn-disabled-bg)',
            color: 'var(--text-primary)',
            fontSize: 13,
            fontWeight: 500,
            cursor: historyRef.current.length > 0 ? 'pointer' : 'not-allowed',
          }}
          title="Undo (Cmd+Z)"
        >
          Undo
        </button>
        {analysisResult && (
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
              {analysisResult.success
                ? `${analysisResult.nodeCount} nodes, ${analysisResult.edgeCount} edges`
                : 'Analysis failed'}
              {analysisResult.warnings && analysisResult.warnings.length > 0 && (
                <span 
                  onClick={() => setShowWarnings((prev) => !prev)}
                  style={{ 
                    color: '#f59e0b', 
                    marginLeft: 8,
                    cursor: 'pointer',
                    textDecoration: 'underline',
                    textUnderlineOffset: 2,
                  }}
                >
                  {analysisResult.warnings.length} warning(s)
                </span>
              )}
            </span>

            {/* Warning Panel */}
            {showWarnings && analysisResult.warnings && analysisResult.warnings.length > 0 && (
              <div
                style={{
                  position: 'absolute',
                  top: '100%',
                  left: 0,
                  marginTop: 12,
                  width: 360,
                  maxHeight: 400,
                  overflowY: 'auto',
                  backgroundColor: 'var(--bg-secondary)',
                  border: '1px solid var(--border-color)',
                  borderRadius: 8,
                  boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
                  zIndex: 1000,
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                <div style={{
                  padding: '12px 16px',
                  borderBottom: '1px solid var(--border-color)',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}>
                  <h3 style={{ margin: 0, fontSize: 14, color: 'var(--text-primary)' }}>Analysis Warnings</h3>
                  <button 
                    onClick={() => setShowWarnings(false)}
                    style={{
                      background: 'none',
                      border: 'none',
                      color: 'var(--text-secondary)',
                      cursor: 'pointer',
                      fontSize: 18,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      padding: 4,
                      lineHeight: 1,
                    }}
                  >
                    ×
                  </button>
                </div>
                <div style={{ padding: 12, display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {analysisResult.warnings.map((w, idx) => (
                    <div 
                      key={idx} 
                      onClick={() => {
                        if (w.nodeIds && w.nodeIds.length > 0) {
                          fitViewToNode(w.nodeIds[0])
                          setShowWarnings(false)
                        }
                      }}
                      style={{
                        padding: 12,
                        backgroundColor: 'var(--bg-primary)',
                        borderLeft: '4px solid #f59e0b',
                        borderRadius: 6,
                        cursor: w.nodeIds && w.nodeIds.length > 0 ? 'pointer' : 'default',
                      }}
                    >
                      <div style={{ 
                        fontSize: 12, 
                        fontWeight: 600, 
                        color: 'var(--text-primary)',
                        marginBottom: 6,
                        textTransform: 'uppercase',
                        letterSpacing: 0.5,
                      }}>
                        {w.rule.replace(/_/g, ' ')}
                      </div>
                      <div style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.4 }}>
                        {w.message}
                      </div>
                      <div style={{ 
                        fontSize: 12, 
                        color: '#b45309', 
                        marginTop: 8, 
                        padding: '6px 8px', 
                        backgroundColor: 'rgba(245, 158, 11, 0.1)', 
                        borderRadius: 4,
                        borderLeft: '3px solid #f59e0b',
                        fontStyle: 'italic'
                      }}>
                        <strong>建議：</strong>{w.solution}
                      </div>
                      {w.nodeIds && w.nodeIds.length > 0 && (
                        <div style={{ 
                          fontSize: 11, 
                          color: '#f59e0b', 
                          marginTop: 8,
                          fontFamily: 'monospace',
                          wordBreak: 'break-all'
                        }}>
                          Nodes: {w.nodeIds.join(', ')}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
      <div ref={reactFlowWrapper} style={{ flex: 1, display: 'flex' }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <ReactFlow
            nodes={nodes.map(node => {
              const nodeWarnings = warningsByNode.get(node.id) ?? []
              const data = node.data as Record<string, unknown>
              return {
                ...node,
                data: {
                  ...data,
                  warnings: nodeWarnings,
                },
              }
            })}
            edges={edges.map(e => ({
              ...e,
              style: { 
                ...e.style, 
                stroke: e.selected 
                  ? '#3b82f6' 
                  : isDarkMode ? '#d1d5db' : '#b1b1b7',
                strokeWidth: e.selected ? 3 : 2,
              }
            }))}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onInit={setRfInstance}
            onDrop={onDrop}
            onDragOver={onDragOver}
            onNodesDelete={onNodesDelete}
            onSelectionChange={onSelectionChange}
            nodeTypes={nodeTypes}
            colorMode={isDarkMode ? 'dark' : 'light'}
            fitView
            snapToGrid
            snapGrid={[16, 16]}
          >
            <Controls />
            <Background gap={16} size={1} color={isDarkMode ? '#4b5563' : '#81818a'} />
          </ReactFlow>
        </div>
        <PropertyPanel
          selectedNode={selectedNode}
          onNodeDataChange={onNodeDataChange}
        />
      </div>
    </div>
  )
}

export default Canvas
