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
import ComponentPropertyPanel from './ComponentPropertyPanel'
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
  initialNodes?: Node[];
  initialEdges?: Edge[];
  onStateChange?: (nodes: Node[], edges: Edge[]) => void;
}

const PROTOCOL_LABELS: Record<string, string> = {
  http: 'HTTP',
  https: 'HTTPS',
  grpc: 'gRPC',
  websocket: 'WebSocket',
  ssh: 'SSH',
  tcp: 'TCP',
  udp: 'UDP',
  amqp: 'AMQP',
  mqtt: 'MQTT',
  database: 'Database',
  dns: 'DNS',
}

const CONNECTION_TYPE_LABELS: Record<string, string> = {
  sync: 'Sync',
  async: 'Async',
  replication: 'Replication',
  cdn_origin: 'CDN Origin',
  unspecified: '',
}

function Canvas({ isDarkMode, initialNodes = [], initialEdges = [], onStateChange }: CanvasProps) {
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(initialEdges)
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [selectedNodeIds, setSelectedNodeIds] = useState<string[]>([])
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null)
  const [analysisResult, setAnalysisResult] = useState<AnalyzeResponse | null>(
    null
  )
  const [analyzing, setAnalyzing] = useState(false)
  const [showWarnings, setShowWarnings] = useState(false)
  const [dismissedWarnings, setDismissedWarnings] = useState<Set<number>>(new Set())

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

  const activeWarnings = useMemo(() => {
    if (!analysisResult?.warnings) return []
    return analysisResult.warnings.filter((_, idx) => !dismissedWarnings.has(idx))
  }, [analysisResult?.warnings, dismissedWarnings])

  const warningsByNode = useMemo(
    () => buildWarningsByNode(activeWarnings),
    [activeWarnings, buildWarningsByNode]
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

    if (onStateChange) {
      onStateChange(nodes, edges)
    }
  }, [nodes, edges, onStateChange])

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

  const canMerge = selectedNodeIds.length === 2
  const selectedNodeForSplit = selectedNode
    ? ((selectedNode.data as Record<string, unknown>).roles as ComponentType[] | undefined)
    : undefined
  const canSplit = selectedNodeForSplit && selectedNodeForSplit.length > 1

  const mergeSelectedNodes = useCallback(() => {
    if (selectedNodeIds.length !== 2) return
    const [primaryId, secondaryId] = selectedNodeIds
    const primary = nodes.find(n => n.id === primaryId)
    const secondary = nodes.find(n => n.id === secondaryId)
    if (!primary || !secondary) return

    pushHistory()

    const primaryData = primary.data as Record<string, unknown>
    const secondaryData = secondary.data as Record<string, unknown>
    const primaryType = primaryData.componentType as ComponentType
    const secondaryType = secondaryData.componentType as ComponentType
    const primaryRoles = (primaryData.roles as ComponentType[] | undefined) ?? [primaryType]
    const secondaryRoles = (secondaryData.roles as ComponentType[] | undefined) ?? [secondaryType]

    // Union of roles (deduplicated)
    const mergedRoles = [...new Set([...primaryRoles, ...secondaryRoles])]

    // Merge properties: union, shared keys use primary's value except replicas (take max)
    const primaryProps = (primaryData.properties as Record<string, unknown>) ?? {}
    const secondaryProps = (secondaryData.properties as Record<string, unknown>) ?? {}
    const mergedProps = { ...secondaryProps, ...primaryProps }
    const primaryReplicas = (primaryProps.replicas as number) ?? 1
    const secondaryReplicas = (secondaryProps.replicas as number) ?? 1
    if (primaryReplicas > 1 || secondaryReplicas > 1) {
      mergedProps.replicas = Math.max(primaryReplicas, secondaryReplicas)
    }

    const mergedData = {
      ...primaryData,
      label: `${primaryData.label}`,
      roles: mergedRoles,
      properties: mergedProps,
      mergedFrom: {
        primary: { id: primaryId, data: primaryData, position: primary.position },
        secondary: { id: secondaryId, data: secondaryData, position: secondary.position },
      },
    }

    // Update primary node with merged data
    setNodes(nds => nds
      .filter(n => n.id !== secondaryId)
      .map(n => n.id === primaryId ? { ...n, data: mergedData } : n)
    )

    // Re-route edges from secondary to primary, remove duplicate edges between them
    setEdges(eds => {
      const rerouted = eds
        .filter(e => !(
          (e.source === primaryId && e.target === secondaryId) ||
          (e.source === secondaryId && e.target === primaryId)
        ))
        .map(e => {
          const newEdge = { ...e }
          if (e.source === secondaryId) {
            newEdge.source = primaryId
            newEdge.id = `edge-${primaryId}-${e.target}-merged`
          }
          if (e.target === secondaryId) {
            newEdge.target = primaryId
            newEdge.id = `edge-${e.source}-${primaryId}-merged`
          }
          return newEdge
        })

      // Deduplicate edges (same source+target)
      const seen = new Set<string>()
      return rerouted.filter(e => {
        const key = `${e.source}-${e.target}`
        if (seen.has(key)) return false
        seen.add(key)
        return true
      })
    })

    setSelectedNodeIds([primaryId])
    setSelectedNodeId(primaryId)
  }, [selectedNodeIds, nodes, pushHistory, setNodes, setEdges])

  const splitSelectedNode = useCallback(() => {
    if (!selectedNode) return
    const data = selectedNode.data as Record<string, unknown>
    const mergedFrom = data.mergedFrom as {
      primary: { id: string; data: Record<string, unknown>; position: { x: number; y: number } }
      secondary: { id: string; data: Record<string, unknown>; position: { x: number; y: number } }
    } | undefined

    if (!mergedFrom) return

    pushHistory()

    // Restore original nodes
    const restoredPrimary: Node = {
      id: mergedFrom.primary.id,
      type: 'architecture',
      position: mergedFrom.primary.position,
      data: mergedFrom.primary.data,
    }
    const restoredSecondary: Node = {
      id: mergedFrom.secondary.id,
      type: 'architecture',
      position: mergedFrom.secondary.position,
      data: mergedFrom.secondary.data,
    }

    setNodes(nds => [
      ...nds.filter(n => n.id !== selectedNode.id),
      restoredPrimary,
      restoredSecondary,
    ])

    // Restore edges: redirect merged edges back
    setEdges(eds => eds.map(e => {
      if (e.source === selectedNode.id && e.id.endsWith('-merged')) {
        return { ...e, source: mergedFrom.primary.id, id: e.id.replace('-merged', '') }
      }
      if (e.target === selectedNode.id && e.id.endsWith('-merged')) {
        return { ...e, target: mergedFrom.primary.id, id: e.id.replace('-merged', '') }
      }
      return e
    }))

    setSelectedNodeId(null)
    setSelectedNodeIds([])
  }, [selectedNode, pushHistory, setNodes, setEdges])

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
      const selIds = new Set(selectedNodes.map(n => n.id))
      setEdges(eds => eds.map(edge => ({
        ...edge,
        selected: selIds.has(edge.source) || selIds.has(edge.target)
      })))
    }
    setSelectedNodeIds(selectedNodes.map(n => n.id))
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
        data: { connectionType: 'unspecified', protocol: '', direction: 'uni' },
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
    setDismissedWarnings(new Set()) // Reset dismissed warnings for new analysis
    try {
      const topology: SystemTopology = {
        id: 'current-design',
        name: 'Untitled Design',
        version: 1,
        nodes: nodes.map((n) => {
          const data = n.data as Record<string, unknown>
          const roles = data.roles as ComponentType[] | undefined
          return {
            id: n.id,
            componentType: data.componentType as ComponentType,
            label: data.label as string,
            position: n.position,
            properties:
              (data.properties as Record<string, unknown>) ?? {},
            ...(roles && roles.length > 1 ? { roles } : {}),
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
            '') as SystemTopology['edges'][number]['protocol'],
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
      if (e.ctrlKey && e.key === 'a') {
        e.preventDefault()
        if (nodes.length > 0 && !analyzing) {
          handleAnalyze()
        }
      }
      if (e.ctrlKey && e.key === 'd') {
        e.preventDefault()
        duplicateSelectedNode()
      }
      if (e.ctrlKey && e.key === 'z') {
        e.preventDefault()
        undo()
      }
      if (e.ctrlKey && e.key === 'm') {
        e.preventDefault()
        if (canMerge) mergeSelectedNodes()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [nodes.length, analyzing, handleAnalyze, duplicateSelectedNode, undo, canMerge, mergeSelectedNodes])

  const handleDemo = useCallback(() => {
    pushHistory()
    
    const demoNodes: Node[] = [
      { id: 'demo-client', type: 'architecture', position: { x: 400, y: 50 }, data: { label: 'Client', componentType: 'client', properties: {} } },
      { id: 'demo-dns', type: 'architecture', position: { x: 700, y: 50 }, data: { label: 'DNS', componentType: 'dns', properties: {} } },
      { id: 'demo-cdn', type: 'architecture', position: { x: 650, y: 250 }, data: { label: 'CDN', componentType: 'cdn', properties: {} } },
      { id: 'demo-lb', type: 'architecture', position: { x: 400, y: 250 }, data: { label: 'Load Balancer', componentType: 'load_balancer', properties: { algorithm: 'round_robin', layer: 'l7' } } },
      { id: 'demo-reverse-proxy', type: 'architecture', position: { x: 400, y: 450 }, data: { label: 'Reverse Proxy', componentType: 'reverse_proxy', properties: { product: 'nginx', sslTermination: true } } },
      { id: 'demo-apigw', type: 'architecture', position: { x: 400, y: 650 }, data: { label: 'API Gateway', componentType: 'api_gateway', properties: { rateLimit: 1000, authEnabled: true } } },
      { id: 'demo-service', type: 'architecture', position: { x: 400, y: 850 }, data: { label: 'Service', componentType: 'service', properties: { replicas: 3 } } },
      { id: 'demo-mq', type: 'architecture', position: { x: 700, y: 850 }, data: { label: 'Message Queue', componentType: 'message_queue', properties: { queueType: 'kafka' } } },
      { id: 'demo-db-master', type: 'architecture', position: { x: 250, y: 1050 }, data: { label: 'Database', componentType: 'database', properties: { dbType: 'sql' } } },
      { id: 'demo-db-slave', type: 'architecture', position: { x: 450, y: 1050 }, data: { label: 'Database', componentType: 'database', properties: { dbType: 'sql' } } },
      { id: 'demo-cache', type: 'architecture', position: { x: 650, y: 1050 }, data: { label: 'Cache', componentType: 'cache', properties: { cacheType: 'distributed' } } },
      { id: 'demo-storage', type: 'architecture', position: { x: 850, y: 650 }, data: { label: 'Storage', componentType: 'storage', properties: {} } }
    ]

    const demoEdges: Edge[] = [
      { id: 'e-client-dns', source: 'demo-client', target: 'demo-dns', sourceHandle: 'right-source', targetHandle: 'left-target', data: { connectionType: 'sync', protocol: 'dns' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-client-cdn', source: 'demo-client', target: 'demo-cdn', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'https' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-client-lb', source: 'demo-client', target: 'demo-lb', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'https' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-cdn-storage', source: 'demo-cdn', target: 'demo-storage', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'https' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-lb-proxy', source: 'demo-lb', target: 'demo-reverse-proxy', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'http' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-proxy-apigw', source: 'demo-reverse-proxy', target: 'demo-apigw', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'http' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-apigw-service', source: 'demo-apigw', target: 'demo-service', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'http' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-mq', source: 'demo-service', target: 'demo-mq', sourceHandle: 'right-source', targetHandle: 'left-target', data: { connectionType: 'async', protocol: 'amqp' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-dbm', source: 'demo-service', target: 'demo-db-master', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'database' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-dbs', source: 'demo-service', target: 'demo-db-slave', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'database' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true },
      { id: 'e-service-cache', source: 'demo-service', target: 'demo-cache', sourceHandle: 'bottom-source', targetHandle: 'top-target', data: { connectionType: 'sync', protocol: 'resp' }, style: { stroke: isDarkMode ? '#d1d5db' : '#b1b1b7', strokeWidth: 2 }, type: 'default', animated: true }
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
          title="Analyze (Ctrl+A)"
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
          title="Duplicate selected (Ctrl+D)"
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
          title="Undo (Ctrl+Z)"
        >
          Undo
        </button>
        {canMerge && (
          <button
            onClick={mergeSelectedNodes}
            style={{
              padding: '6px 12px',
              borderRadius: 6,
              border: '1px solid var(--border-color)',
              backgroundColor: 'var(--bg-secondary)',
              color: 'var(--text-primary)',
              fontSize: 13,
              fontWeight: 500,
              cursor: 'pointer',
            }}
            title="Merge selected nodes (Ctrl+M)"
          >
            Merge
          </button>
        )}
        {canSplit && (
          <button
            onClick={splitSelectedNode}
            style={{
              padding: '6px 12px',
              borderRadius: 6,
              border: '1px solid var(--border-color)',
              backgroundColor: 'var(--bg-secondary)',
              color: 'var(--text-primary)',
              fontSize: 13,
              fontWeight: 500,
              cursor: 'pointer',
            }}
            title="Split merged node back into individual components"
          >
            Split
          </button>
        )}
        {analysisResult && (
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
              {analysisResult.success
                ? `${analysisResult.nodeCount} nodes, ${analysisResult.edgeCount} edges`
                : 'Analysis failed'}
              {activeWarnings.length > 0 && (
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
                  {activeWarnings.length} warning(s)
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
              const connectionType = edgeData.connectionType as string | undefined
              const edgeLabel = edgeData.label as string | undefined
              
              let autoLabel = ''
              const protocolLabel = protocol && protocol !== 'unspecified' ? (PROTOCOL_LABELS[protocol] || protocol) : ''
              const connTypeLabel = connectionType && connectionType !== 'unspecified' ? (CONNECTION_TYPE_LABELS[connectionType] || connectionType) : ''
              
              if (protocolLabel && connTypeLabel) {
                autoLabel = `${protocolLabel} (${connTypeLabel})`
              } else if (protocolLabel) {
                autoLabel = protocolLabel
              } else if (connTypeLabel) {
                autoLabel = connTypeLabel
              }

              const displayLabel = edgeLabel || autoLabel || undefined
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
          {showWarnings && activeWarnings.length > 0 && (
            <div
              style={{
                height: panelHeight,
                minHeight: 100,
                backgroundColor: 'var(--bg-primary)',
                borderTop: '1px solid var(--border-color)',
                display: 'flex',
                flexDirection: 'column',
                zIndex: 10,
                position: 'relative',
                boxShadow: '0 -4px 12px rgba(0, 0, 0, 0.05)',
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
                backgroundColor: 'var(--bg-secondary)',
                padding: '0 16px',
                borderBottom: '1px solid var(--border-color)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                minHeight: '36px',
              }}>
                <div style={{ display: 'flex', height: '100%' }}>
                  <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    borderBottom: '1px solid var(--text-primary)',
                    padding: '0 8px',
                    marginBottom: '-1px',
                    gap: 8,
                  }}>
                    <h3 style={{ margin: 0, fontSize: 11, textTransform: 'uppercase', color: 'var(--text-primary)', letterSpacing: 0.5 }}>
                      PROBLEMS
                    </h3>
                    <span style={{ 
                      backgroundColor: isDarkMode ? '#1e293b' : '#f1f5f9',
                      color: 'var(--text-primary)', 
                      padding: '2px 6px', 
                      borderRadius: 12, 
                      fontSize: 10,
                      fontWeight: 600,
                    }}>{activeWarnings.length}</span>
                  </div>
                </div>
                <button 
                  onClick={() => setShowWarnings(false)}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: 'var(--text-secondary)',
                    cursor: 'pointer',
                    fontSize: 16,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    padding: '4px',
                    lineHeight: 1,
                    borderRadius: 4,
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.backgroundColor = 'var(--text-primary)'
                    e.currentTarget.style.color = 'var(--bg-primary)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.backgroundColor = 'transparent'
                    e.currentTarget.style.color = 'var(--text-secondary)'
                  }}
                  title="Close Panel"
                >
                  ×
                </button>
              </div>
              <div style={{ padding: '8px 0', display: 'flex', flexDirection: 'column', flex: 1, overflowY: 'auto' }}>
                {analysisResult?.warnings?.map((w, idx) => {
                  if (dismissedWarnings.has(idx)) return null
                  return (
                    <div 
                      key={idx} 
                      onClick={() => {
                        if (w.nodeIds && w.nodeIds.length > 0) {
                          fitViewToNode(w.nodeIds[0])
                        }
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.backgroundColor = isDarkMode ? '#1e293b' : '#f8fafc'
                        const closeBtn = e.currentTarget.querySelector('.dismiss-btn') as HTMLElement
                        if (closeBtn) closeBtn.style.opacity = '1'
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.backgroundColor = 'transparent'
                        const closeBtn = e.currentTarget.querySelector('.dismiss-btn') as HTMLElement
                        if (closeBtn) closeBtn.style.opacity = '0'
                      }}
                      style={{
                        padding: '8px 16px 8px 36px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 2,
                        cursor: w.nodeIds && w.nodeIds.length > 0 ? 'pointer' : 'default',
                        position: 'relative',
                        transition: 'background-color 0.1s ease',
                      }}
                    >
                      <svg 
                        viewBox="0 0 16 16" 
                        fill="#f59e0b" 
                        style={{ position: 'absolute', left: 14, top: 10, width: 14, height: 14 }}
                      >
                        <path fillRule="evenodd" clipRule="evenodd" d="M7.56 1h.88l6.5 13v.88H1.05V14l6.5-13zm.45 1.74l-5.6 11.2h11.2L8.01 2.74zM9 12.01H7v-1.9h2v1.9zm0-2.92H7V5.01h2v4.08z" />
                      </svg>
                      
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <div style={{ fontSize: 13, color: 'var(--text-primary)', lineHeight: 1.5, display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 6 }}>
                          <span>{w.message}</span>
                          <span style={{ fontSize: 11, color: 'var(--text-secondary)', opacity: 0.7, fontFamily: 'sans-serif' }}>
                            [{w.rule.replace(/_/g, ' ')}]
                          </span>
                        </div>
                        <button
                          className="dismiss-btn"
                          onClick={(e) => {
                            e.stopPropagation()
                            setDismissedWarnings(prev => {
                              const next = new Set(prev)
                              next.add(idx)
                              return next
                            })
                          }}
                          style={{
                            background: 'none',
                            border: 'none',
                            color: 'var(--text-secondary)',
                            cursor: 'pointer',
                            fontSize: 16,
                            opacity: 0,
                            transition: 'opacity 0.2s',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            padding: '0 4px',
                            marginLeft: 8,
                          }}
                          title="Dismiss warning"
                        >
                          ×
                        </button>
                      </div>

                      {w.solution && (
                        <div style={{ 
                          fontSize: 12, 
                          color: 'var(--text-secondary)', 
                          display: 'flex',
                          alignItems: 'flex-start',
                          gap: 6,
                          marginTop: 4
                        }}>
                          <svg viewBox="0 0 16 16" fill="currentColor" style={{ width: 12, height: 12, marginTop: 2, flexShrink: 0, opacity: 0.7 }}>
                            <path fillRule="evenodd" clipRule="evenodd" d="M8 1.5a4.5 4.5 0 00-4.5 4.5c0 1.25.5 2.4 1.34 3.23.4.4.66.95.66 1.52A1.25 1.25 0 006.75 12h2.5a1.25 1.25 0 001.25-1.25c0-.57.26-1.12.66-1.52.84-.83 1.34-1.98 1.34-3.23A4.5 4.5 0 008 1.5zm-3.5 4.5a3.5 3.5 0 117 0c0 .99-.4 1.88-1.04 2.53-.54.54-.86 1.26-.86 2.02v.2H6.4v-.2c0-.76-.32-1.48-.86-2.02A3.5 3.5 0 014.5 6zM6 13.5v.5a.5.5 0 00.5.5h3a.5.5 0 00.5-.5v-.5H6z" />
                          </svg>
                          <span style={{ fontStyle: 'italic', opacity: 0.9 }}>{w.solution}</span>
                        </div>
                      )}
                      
                      {w.nodeIds && w.nodeIds.length > 0 && (
                        <div style={{ fontSize: 11, color: '#3b82f6', opacity: 0.9, fontFamily: 'monospace', marginTop: 4 }}>
                          Nodes: {w.nodeIds.join(', ')}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
        <ComponentPropertyPanel
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
