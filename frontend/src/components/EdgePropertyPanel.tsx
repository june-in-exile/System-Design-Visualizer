import type { Edge } from '@xyflow/react'
import type { EdgeProtocol, ConnectionType } from '../types/topology'

interface EdgePropertyPanelProps {
  selectedEdgeId: string | null
  edges: Edge[]
  onEdgeDataChange: (edgeId: string, newData: Record<string, unknown>) => void
}

const PROTOCOL_OPTIONS: { value: EdgeProtocol; label: string }[] = [
  { value: 'http', label: 'HTTP' },
  { value: 'https', label: 'HTTPS' },
  { value: 'grpc', label: 'gRPC' },
  { value: 'websocket', label: 'WebSocket' },
  { value: 'ssh', label: 'SSH' },
  { value: 'tcp', label: 'TCP' },
  { value: 'udp', label: 'UDP' },
  { value: 'amqp', label: 'AMQP' },
  { value: 'mqtt', label: 'MQTT' },
  { value: 'database', label: 'Database' },
  { value: 'dns', label: 'DNS' },
]

const CONNECTION_TYPE_OPTIONS: { value: ConnectionType; label: string }[] = [
  { value: 'sync', label: 'Synchronous' },
  { value: 'async', label: 'Asynchronous' },
  { value: 'replication', label: 'Replication' },
  { value: 'cdn_origin', label: 'CDN Origin' },
]

export default function EdgePropertyPanel({
  selectedEdgeId,
  edges,
  onEdgeDataChange,
}: EdgePropertyPanelProps) {
  const selectedEdge = edges.find((e) => e.id === selectedEdgeId)

  if (!selectedEdge) {
    return null
  }

  const edgeData = (selectedEdge.data as Record<string, unknown>) ?? {}
  const protocol = (edgeData.protocol as EdgeProtocol) ?? 'http'
  const connectionType = (edgeData.connectionType as ConnectionType) ?? 'sync'
  const label = edgeData.label as string | undefined

  const handleProtocolChange = (newProtocol: EdgeProtocol) => {
    onEdgeDataChange(selectedEdge.id, {
      ...edgeData,
      protocol: newProtocol,
    })
  }

  const handleConnectionTypeChange = (newConnectionType: ConnectionType) => {
    onEdgeDataChange(selectedEdge.id, {
      ...edgeData,
      connectionType: newConnectionType,
    })
  }

  const handleLabelChange = (newLabel: string) => {
    onEdgeDataChange(selectedEdge.id, {
      ...edgeData,
      label: newLabel,
    })
  }

  const sourceNode = edges.find(e => e.id === selectedEdge.id)?.source
  const targetNode = edges.find(e => e.id === selectedEdge.id)?.target

  return (
    <aside
      style={{
        width: 300,
        padding: 16,
        borderLeft: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        overflowY: 'auto',
      }}
    >
      <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 600, color: 'var(--text-primary)' }}>
        Edge Properties
      </h3>

      <div style={{ marginBottom: 16, padding: 8, backgroundColor: 'var(--bg-primary)', borderRadius: 4 }}>
        <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
          <div><strong>From:</strong> {sourceNode}</div>
          <div><strong>To:</strong> {targetNode}</div>
        </div>
      </div>

      <div style={{ marginBottom: 16 }}>
        <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
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
        <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
          Protocol
        </label>
        <select
          value={protocol}
          onChange={(e) => handleProtocolChange(e.target.value as EdgeProtocol)}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
          {PROTOCOL_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>

      <div style={{ marginBottom: 16 }}>
        <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
          Connection Type
        </label>
        <select
          value={connectionType}
          onChange={(e) => handleConnectionTypeChange(e.target.value as ConnectionType)}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
          {CONNECTION_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>
    </aside>
  )
}
