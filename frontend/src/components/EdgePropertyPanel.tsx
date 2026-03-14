import type { Edge } from '@xyflow/react'
import type { EdgeProtocol, ConnectionType, EdgeDirection } from '../types/topology'

interface EdgePropertyPanelProps {
  selectedEdgeId: string | null
  edges: Edge[]
  onEdgeDataChange: (edgeId: string, newData: Record<string, unknown>) => void
  onEdgeAnimatedChange: (edgeId: string, animated: boolean) => void
  onEdgeDirectionChange: (edgeId: string, direction: EdgeDirection) => void
  onEdgeReverse: (edgeId: string) => void
}

const PROTOCOL_OPTIONS: { value: EdgeProtocol; label: string; description: string }[] = [
  { value: '', label: '(Unspecified)', description: 'No specific protocol defined' },
  { value: 'http', label: 'HTTP', description: 'Hypertext Transfer Protocol (Unsecured)' },
  { value: 'https', label: 'HTTPS', description: 'Hypertext Transfer Protocol Secure (TLS/SSL)' },
  { value: 'grpc', label: 'gRPC', description: 'High-performance, open source universal RPC framework' },
  { value: 'websocket', label: 'WebSocket', description: 'Full-duplex communication over a single TCP connection' },
  { value: 'ssh', label: 'SSH', description: 'Secure Shell protocol for secure remote login' },
  { value: 'tcp', label: 'TCP', description: 'Transmission Control Protocol (Connection-oriented)' },
  { value: 'udp', label: 'UDP', description: 'User Datagram Protocol (Connectionless)' },
  { value: 'amqp', label: 'AMQP', description: 'Advanced Message Queuing Protocol' },
  { value: 'mqtt', label: 'MQTT', description: 'Message Queuing Telemetry Transport (IoT)' },
  { value: 'database', label: 'Database (Generic)', description: 'Direct database connection protocol (JDBC/ODBC/Native)' },
  { value: 'resp', label: 'RESP', description: 'REdis Serialization Protocol (used by Redis)' },
  { value: 'binary', label: 'Binary Protocol', description: 'Custom binary protocol for high-performance communication' },
  { value: 'uds', label: 'UDS', description: 'Unix Domain Socket (Inter-process communication)' },
  { value: 'dns', label: 'DNS', description: 'Domain Name System query/response' },
]

const CONNECTION_TYPE_OPTIONS: { value: ConnectionType; label: string; description: string }[] = [
  { value: 'unspecified', label: '(Unspecified)', description: 'No specific connection type defined' },
  { value: 'sync', label: 'Synchronous', description: 'Request-response, client waits for server to process' },
  { value: 'async', label: 'Asynchronous', description: 'Fire-and-forget or callback-based, client doesn\'t wait' },
  { value: 'replication', label: 'Replication', description: 'Data synchronization between primary and replicas' },
  { value: 'cdn_origin', label: 'CDN Origin', description: 'Fetching content from origin server to edge cache' },
]

const DIRECTION_OPTIONS: { value: EdgeDirection; label: string; description: string }[] = [
  { value: 'uni', label: 'One-way (→)', description: 'Unidirectional flow from source to target' },
  { value: 'bi', label: 'Two-way (↔)', description: 'Bidirectional communication between nodes' },
  { value: 'none', label: 'No arrows (―)', description: 'Connection without specific directionality' },
]

const getTooltip = (label: string, description?: string) => {
  const lowerLabel = label.toLowerCase()
  if (!description || lowerLabel.includes('unspecified') || lowerLabel.includes('default') || lowerLabel.includes('auto-detect')) {
    return undefined
  }
  return description
}

export default function EdgePropertyPanel({
  selectedEdgeId,
  edges,
  onEdgeDataChange,
  onEdgeAnimatedChange,
  onEdgeDirectionChange,
  onEdgeReverse,
}: EdgePropertyPanelProps) {
  const selectedEdge = edges.find((e) => e.id === selectedEdgeId)

  if (!selectedEdge) {
    return null
  }

  const edgeData = (selectedEdge.data as Record<string, unknown>) ?? {}
  const rawProtocol = edgeData.protocol as string | undefined
  const protocol = (rawProtocol === 'unspecified' ? '' : (rawProtocol ?? '')) as EdgeProtocol
  const rawConnectionType = edgeData.connectionType as string | undefined
  const connectionType = (!rawConnectionType ? 'unspecified' : rawConnectionType) as ConnectionType
  const label = edgeData.label as string | undefined
  const direction = (edgeData.direction as EdgeDirection) ?? 'uni'
  const isAnimated = selectedEdge.animated ?? false

  const handleProtocolChange = (newProtocol: EdgeProtocol) => {
    onEdgeDataChange(selectedEdge.id, {
      protocol: newProtocol,
    })
  }

  const handleConnectionTypeChange = (newConnectionType: ConnectionType) => {
    onEdgeDataChange(selectedEdge.id, {
      connectionType: newConnectionType,
    })
  }

  const handleLabelChange = (newLabel: string) => {
    onEdgeDataChange(selectedEdge.id, {
      label: newLabel,
    })
  }

  return (
    <aside
      style={{
        width: '100%',
        padding: 16,
        overflowY: 'auto',
        flex: 1,
      }}
    >
      <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 600, color: 'var(--text-primary)' }}>
        Edge Properties
      </h3>

      <div style={{ marginBottom: 16 }}>
        <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
          Direction
        </label>
        <select
          value={direction}
          onChange={(e) => onEdgeDirectionChange(selectedEdge.id, e.target.value as EdgeDirection)}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
          {DIRECTION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
          ))}
        </select>
      </div>

      {direction === 'uni' && (
        <div style={{ marginBottom: 16 }}>
          <button
            onClick={() => onEdgeReverse(selectedEdge.id)}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid var(--border-color)',
              borderRadius: 4,
              fontSize: 13,
              backgroundColor: 'var(--bg-primary)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
            }}
          >
            ⇄ Reverse Direction
          </button>
        </div>
      )}

      <div style={{ marginBottom: 16 }}>
        <label 
          style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
          title="The name of this connection, displayed on the edge."
        >
          Label
        </label>
        <input
          type="text"
          value={label ?? ''}
          onChange={(e) => handleLabelChange(e.target.value)}
          placeholder="Optional label"
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        />
      </div>

      <div style={{ marginBottom: 16 }}>
        <label 
          style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
          title="The communication protocol used for this connection."
        >
          Protocol
        </label>
        <select
          value={protocol}
          onChange={(e) => handleProtocolChange(e.target.value as EdgeProtocol)}
          title={getTooltip(
            PROTOCOL_OPTIONS.find(opt => opt.value === protocol)?.label || '',
            PROTOCOL_OPTIONS.find(opt => opt.value === protocol)?.description
          )}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
          {PROTOCOL_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
          ))}
        </select>
      </div>

      <div style={{ marginBottom: 16 }}>
        <label 
          style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
          title="The logical nature of the connection (e.g., blocking vs non-blocking)."
        >
          Connection Type
        </label>
        <select
          value={connectionType}
          onChange={(e) => handleConnectionTypeChange(e.target.value as ConnectionType)}
          title={getTooltip(
            CONNECTION_TYPE_OPTIONS.find(opt => opt.value === connectionType)?.label || '',
            CONNECTION_TYPE_OPTIONS.find(opt => opt.value === connectionType)?.description
          )}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
          {CONNECTION_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
          ))}
        </select>
      </div>

      <div style={{ marginBottom: 16 }}>
        <label 
          style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}
          title="Enable a visual animation to indicate data flow direction."
        >
          <input
            type="checkbox"
            checked={isAnimated}
            onChange={(e) => onEdgeAnimatedChange(selectedEdge.id, e.target.checked)}
            style={{ width: 16, height: 16, cursor: 'pointer' }}
          />
          <span style={{ fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
            Animated (Flow Effect)
          </span>
        </label>
      </div>
    </aside>
  )
}
