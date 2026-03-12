import { useCallback, useRef, useState } from 'react'
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
import { NODE_TYPE_CONFIG } from '../nodes/nodeConfig'
import { analyzeTopology } from '../api/topologyApi'
import type {
  ComponentType,
  SystemTopology,
  AnalyzeResponse,
} from '../types/topology'

const nodeTypes = {
  architecture: ArchitectureNode,
}

let nodeIdCounter = 0
function generateNodeId(): string {
  nodeIdCounter += 1
  return `node-${nodeIdCounter}`
}

function Canvas() {
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null)
  const [analysisResult, setAnalysisResult] = useState<AnalyzeResponse | null>(
    null
  )
  const [analyzing, setAnalyzing] = useState(false)

  const onConnect: OnConnect = useCallback(
    (params) => {
      const newEdge: Edge = {
        id: `edge-${params.source}-${params.target}`,
        source: params.source,
        target: params.target,
        sourceHandle: params.sourceHandle,
        targetHandle: params.targetHandle,
        data: { connectionType: 'sync' },
      }
      setEdges((eds) => [...eds, newEdge])
    },
    [setEdges]
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
    [rfInstance, setNodes]
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
      setAnalysisResult({
        success: false,
        nodeCount: 0,
        edgeCount: 0,
        warnings: [
          error instanceof Error ? error.message : 'Analysis failed',
        ],
      })
    } finally {
      setAnalyzing(false)
    }
  }, [nodes, edges])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
      <div
        style={{
          padding: '8px 16px',
          borderBottom: '1px solid #e5e7eb',
          display: 'flex',
          alignItems: 'center',
          gap: 12,
        }}
      >
        <button
          onClick={handleAnalyze}
          disabled={analyzing || nodes.length === 0}
          style={{
            padding: '6px 16px',
            borderRadius: 6,
            border: 'none',
            backgroundColor: nodes.length === 0 ? '#d1d5db' : '#3b82f6',
            color: '#fff',
            fontSize: 13,
            fontWeight: 600,
            cursor: nodes.length === 0 ? 'not-allowed' : 'pointer',
          }}
        >
          {analyzing ? 'Analyzing...' : 'Analyze'}
        </button>
        {analysisResult && (
          <span style={{ fontSize: 13, color: '#6b7280' }}>
            {analysisResult.success
              ? `${analysisResult.nodeCount} nodes, ${analysisResult.edgeCount} edges`
              : 'Analysis failed'}
            {analysisResult.warnings && analysisResult.warnings.length > 0 && (
              <span style={{ color: '#f59e0b', marginLeft: 8 }}>
                {analysisResult.warnings.length} warning(s)
              </span>
            )}
          </span>
        )}
      </div>
      <div ref={reactFlowWrapper} style={{ flex: 1 }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onInit={setRfInstance}
          onDrop={onDrop}
          onDragOver={onDragOver}
          nodeTypes={nodeTypes}
          fitView
          snapToGrid
          snapGrid={[16, 16]}
        >
          <Controls />
          <Background gap={16} size={1} />
        </ReactFlow>
      </div>
    </div>
  )
}

export default Canvas
