import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  SelectionMode,
  MarkerType,
  type OnConnect,
  type Node,
  type Edge,
  type ReactFlowInstance,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import ArchitectureNode from '../nodes/ArchitectureNode'
import PropertyPanel from './PropertyPanel'
import EdgePropertyPanel from './EdgePropertyPanel'
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
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null)
  const [analysisResult, setAnalysisResult] = useState<AnalyzeResponse | null>(
    null
  )
  const [analyzing, setAnalyzing] = useState(false)
  const [showWarnings, setShowWarnings] = useState(false)

  const [panelHeight, setPanelHeight] = useState(250)
  const isDraggingRef = useRef(false)
  const panelHeightRef = useRef(250)
  
  useEffect(() => {
    panelHeightRef.current = panelHeight
  }, [panelHeight])

  const onDragStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    isDraggingRef.current = true
    const startY = e.clientY
    const startHeight = panelHeightRef.current

    const onMouseMove = (moveEvent: MouseEvent) => {
      if (!isDraggingRef.current) return
      const delta = startY - moveEvent.clientY
      const newHeight = Math.max(100, Math.min(window.innerHeight * 0.8, startHeight + delta))
      setPanelHeight(newHeight)
    }

    const onMouseUp = () => {
      isDraggingRef.current = false
      window.removeEventListener('mousemove', onMouseMove)
      window.removeEventListener('mouseup', onMouseUp)
      document.body.style.cursor = 'default'
    }

    document.body.style.cursor = 'row-resize'
    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  }, [])

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

  // Auto-show warnings panel when analysis returns warnings
  useEffect(() => {
    if (!analyzing && analysisResult?.warnings && analysisResult.warnings.length > 0) {
      setShowWarnings(true)
    }
  }, [analyzing, analysisResult])

  const selectedNode = nodes.find((n) => n.id === selectedNodeId) || null

  const duplicateSelectedNode = useCallback(() => {
    if (!selectedNode) return
    
    pushHistory()
    const data = selectedNode.data as Record<string, unknown>
    const newNodeId = generateNodeId()
    const newNode: Node = {
      id: newNodeId,
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
    
    const connectedEdges = edges.filter(
      (e) => e.source === selectedNode.id || e.target === selectedNode.id
    )
    
    const newEdges: Edge[] = connectedEdges.map((edge) => ({
      id: `edge-${newNodeId}-${edge.source === selectedNode.id ? edge.target : edge.source}`,
      source: edge.source === selectedNode.id ? newNodeId : edge.source,
      target: edge.target === selectedNode.id ? newNodeId : edge.target,
      sourceHandle: edge.sourceHandle,
      targetHandle: edge.targetHandle,
      data: edge.data,
      style: edge.style,
      type: edge.type,
      animated: edge.animated,
    }))
    
    setNodes((nds) => [...nds, newNode])
    setEdges((eds) => [...eds, ...newEdges])
    setSelectedNodeId(newNodeId)
  }, [selectedNode, edges, setNodes, setEdges])

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

  const onEdgeDataChange = useCallback(
    (edgeId: string, newData: Record<string, unknown>) => {
      pushHistory()
      setEdges((eds) =>
        eds.map((edge) =>
          edge.id === edgeId ? { ...edge, data: newData } : edge
        )
      )
    },
    [setEdges, pushHistory]
  )

  const onEdgeAnimatedChange = useCallback(
    (edgeId: string, animated: boolean) => {
      pushHistory()
      setEdges((eds) =>
        eds.map((edge) =>
          edge.id === edgeId ? { ...edge, animated } : edge
        )
      )
    },
    [setEdges, pushHistory]
  )

  const onEdgeDirectionChange = useCallback(
    (edgeId: string, direction: 'uni' | 'bi' | 'none') => {
      pushHistory()
      setEdges((eds) =>
        eds.map((edge) => {
          if (edge.id !== edgeId) return edge
          
          const color = isDarkMode ? '#d1d5db' : '#b1b1b7'
          let markerEnd: Edge['markerEnd'] = undefined
          let markerStart: Edge['markerStart'] = undefined
          
          if (direction === 'uni') {
            markerEnd = { type: MarkerType.ArrowClosed, color }
          } else if (direction === 'bi') {
            markerEnd = { type: MarkerType.ArrowClosed, color }
            markerStart = { type: MarkerType.ArrowClosed, color }
          }
          
          return {
            ...edge,
            data: { ...edge.data, direction },
            markerEnd,
            markerStart,
          }
        })
      )
    },
    [setEdges, pushHistory, isDarkMode]
  )

  const onEdgeReverse = useCallback(
    (edgeId: string) => {
      pushHistory()
      setEdges((eds) =>
        eds.map((edge) => {
          if (edge.id !== edgeId) return edge
          const oldSourceHandle = edge.sourceHandle ?? ''
          const oldTargetHandle = edge.targetHandle ?? ''
          return {
            ...edge,
            source: edge.target,
            target: edge.source,
            sourceHandle: oldTargetHandle.replace('-target', '-source'),
            targetHandle: oldSourceHandle.replace('-source', '-target'),
          }
        })
      )
    },
    [setEdges, pushHistory]
  )

  const onSelectionChange = useCallback(({ nodes: selectedNodes, edges: selectedEdges }: { nodes: Node[]; edges: Edge[] }) => {
    if (selectedNodes.length > 0) {
      const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
      setEdges(eds => eds.map(edge => ({
        ...edge,
        selected: selectedNodeIds.has(edge.source) || selectedNodeIds.has(edge.target)
      })))
    }
    setSelectedNodeId(selectedNodes.length > 0 ? selectedNodes[0].id : null)
    setSelectedEdgeId(selectedEdges.length > 0 ? selectedEdges[0].id : null)
  }, [setEdges])

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
        data: { connectionType: 'sync', protocol: 'http', direction: 'uni' },
        style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 },
        markerEnd: { type: MarkerType.ArrowClosed, color: isDarkMode ? '#d1d5db' : '#b1b1b7' },
        animated: true,
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
          protocol:
            ((e.data as Record<string, unknown>)?.protocol as string ??
            'http') as SystemTopology['edges'][number]['protocol'],
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

  const handleDemo = useCallback(() => {
    pushHistory()
    
    const demoNodes: Node[] = [
      { id: 'demo-client', type: 'architecture', position: { x: 400, y: 50 }, data: { label: 'Client', componentType: 'client', properties: {} } },
      { id: 'demo-dns', type: 'architecture', position: { x: 700, y: 50 }, data: { label: 'DNS', componentType: 'dns', properties: {} } },
      { id: 'demo-cdn', type: 'architecture', position: { x: 650, y: 250 }, data: { label: 'CDN', componentType: 'cdn', properties: {} } },
      { id: 'demo-lb', type: 'architecture', position: { x: 400, y: 250 }, data: { label: 'Load Balancer', componentType: 'load_balancer', properties: { algorithm: 'round_robin', layer: 'l7' } } },
      { id: 'demo-apigw', type: 'architecture', position: { x: 400, y: 450 }, data: { label: 'API Gateway', componentType: 'api_gateway', properties: { rateLimit: 1000, authEnabled: true } } },
      { id: 'demo-service', type: 'architecture', position: { x: 400, y: 650 }, data: { label: 'Service', componentType: 'service', properties: { replicas: 3 } } },
      { id: 'demo-mq', type: 'architecture', position: { x: 700, y: 650 }, data: { label: 'Message Queue', componentType: 'message_queue', properties: { queueType: 'kafka' } } },
      { id: 'demo-db-master', type: 'architecture', position: { x: 250, y: 850 }, data: { label: 'Database', componentType: 'database', properties: { dbType: 'sql' } } },
      { id: 'demo-db-slave', type: 'architecture', position: { x: 450, y: 850 }, data: { label: 'Database', componentType: 'database', properties: { dbType: 'sql' } } },
      { id: 'demo-cache', type: 'architecture', position: { x: 650, y: 850 }, data: { label: 'Cache', componentType: 'cache', properties: { cacheType: 'distributed' } } },
      { id: 'demo-storage', type: 'architecture', position: { x: 850, y: 450 }, data: { label: 'Storage', componentType: 'storage', properties: {} } }
    ]

    const demoEdges: Edge[] = [
      { id: 'e-client-dns', source: 'demo-client', target: 'demo-dns', sourceHandle: 'right-source', targetHandle: 'left-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-client-cdn', source: 'demo-client', target: 'demo-cdn', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-client-lb', source: 'demo-client', target: 'demo-lb', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-cdn-storage', source: 'demo-cdn', target: 'demo-storage', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-lb-apigw', source: 'demo-lb', target: 'demo-apigw', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-apigw-service', source: 'demo-apigw', target: 'demo-service', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-mq', source: 'demo-service', target: 'demo-mq', sourceHandle: 'right-source', targetHandle: 'left-target', data: { connectionType: 'async' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-dbm', source: 'demo-service', target: 'demo-db-master', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-dbs', source: 'demo-service', target: 'demo-db-slave', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-cache', source: 'demo-service', target: 'demo-cache', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true }
    ]

    setNodes(demoNodes)
    setEdges(demoEdges)

    setTimeout(() => {
      if (rfInstance) {
        rfInstance.fitView({ padding: 0.2, duration: 500 })
      }
    }, 50)
  }, [pushHistory, setNodes, setEdges, isDarkMode, rfInstance])

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


          </div>
        )}
      </div>
      <div ref={reactFlowWrapper} style={{ flex: 1, display: 'flex' }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', position: 'relative' }}>
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
            edges={edges.map(e => {
              const edgeData = (e.data as Record<string, unknown>) ?? {}
              const protocol = edgeData.protocol as string | undefined
              const edgeLabel = edgeData.label as string | undefined
              const displayLabel = edgeLabel || (protocol ? protocol.toUpperCase() : undefined)
              const color = e.selected ? '#3b82f6' : isDarkMode ? '#d1d5db' : '#b1b1b7'
              const direction = edgeData.direction as string | undefined
              const arrowMarker = { type: MarkerType.ArrowClosed, color }
              const markerEnd = direction === 'none' ? undefined : arrowMarker
              const markerStart = direction === 'bi' ? arrowMarker : undefined
              return {
                ...e,
                selectable: true,
                label: displayLabel,
                labelStyle: { fill: color, fontSize: 11, fontWeight: 600 },
                labelBgStyle: { fill: isDarkMode ? '#1f2937' : '#ffffff', fillOpacity: 0.85 },
                labelBgPadding: [6, 3] as [number, number],
                labelBgBorderRadius: 3,
                markerEnd,
                markerStart,
                style: {
                  ...e.style,
                  stroke: color,
                  strokeWidth: e.selected ? 3 : 2,
                },
              }
            })}
            selectionMode={SelectionMode.Partial}
            selectionOnDrag
            panOnScroll
            multiSelectionKeyCode="Shift"
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
            defaultEdgeOptions={{
              selectable: true,
              type: 'default',
            }}
            proOptions={{ hideAttribution: true }}
          >
            <Controls />
            <Background gap={16} size={1} color={isDarkMode ? '#4b5563' : '#81818a'} />
            
            <div style={{ position: 'absolute', bottom: 16, left: 70, zIndex: 10 }}>
              <button
                onClick={handleDemo}
                style={{
                  padding: '8px 16px',
                  borderRadius: 6,
                  border: '1px solid var(--border-color)',
                  backgroundColor: 'var(--bg-secondary)',
                  color: 'var(--text-primary)',
                  fontSize: 13,
                  fontWeight: 600,
                  cursor: 'pointer',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
                }}
              >
                Demo
              </button>
            </div>
          </ReactFlow>
          </div>

          {/* Warning Panel terminal style */}
          {showWarnings && analysisResult?.warnings && analysisResult.warnings.length > 0 && (
            <div
              style={{
                height: panelHeight,
                minHeight: 100,
                backgroundColor: 'var(--bg-secondary)',
                borderTop: '1px solid var(--border-color)',
                display: 'flex',
                flexDirection: 'column',
                zIndex: 10,
                position: 'relative',
              }}
            >
              <div
                onMouseDown={onDragStart}
                style={{
                  position: 'absolute',
                  top: -4,
                  left: 0,
                  right: 0,
                  height: 8,
                  cursor: 'row-resize',
                  zIndex: 20,
                  backgroundColor: 'transparent',
                }}
              />
              <div style={{
                backgroundColor: 'var(--bg-primary)',
                padding: '8px 16px',
                borderBottom: '1px solid var(--border-color)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
                <div style={{ display: 'flex', gap: 16 }}>
                  <h3 style={{ margin: 0, fontSize: 12, textTransform: 'uppercase', color: 'var(--text-primary)', letterSpacing: 0.5, borderBottom: '1px solid #3b82f6', paddingBottom: 6 }}>
                    PROBLEMS <span style={{ backgroundColor: 'rgba(245, 158, 11, 0.2)', color: '#f59e0b', padding: '2px 6px', borderRadius: 10, marginLeft: 6, fontSize: 11 }}>{analysisResult.warnings.length}</span>
                  </h3>
                </div>
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
                  title="Close Panel"
                >
                  ×
                </button>
              </div>
              <div style={{ padding: '0 12px 12px 12px', display: 'flex', flexDirection: 'column', gap: 12, flex: 1, overflowY: 'auto' }}>
                <div style={{ marginTop: 12 }} />
                {analysisResult.warnings.map((w, idx) => (
                  <div 
                    key={idx} 
                    onClick={() => {
                      if (w.nodeIds && w.nodeIds.length > 0) {
                        fitViewToNode(w.nodeIds[0])
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
        <PropertyPanel
          selectedNode={selectedNode}
          onNodeDataChange={onNodeDataChange}
        />
        <EdgePropertyPanel
          selectedEdgeId={selectedEdgeId}
          edges={edges}
          onEdgeDataChange={onEdgeDataChange}
          onEdgeAnimatedChange={onEdgeAnimatedChange}
          onEdgeDirectionChange={onEdgeDirectionChange}
          onEdgeReverse={onEdgeReverse}
        />
      </div>
    </div>
  )
}

export default Canvas
