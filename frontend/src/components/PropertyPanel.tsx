import type { Node } from '@xyflow/react'
import type { ComponentType } from '../types/topology'

interface PropertyPanelProps {
  selectedNode: Node | null
  onNodeDataChange: (nodeId: string, newData: Record<string, unknown>) => void
}

export default function PropertyPanel({
  selectedNode,
  onNodeDataChange,
}: PropertyPanelProps) {
  if (!selectedNode) {
    return null
  }

  const data = selectedNode.data as Record<string, unknown>
  const properties = (data.properties as Record<string, unknown>) ?? {}
  const componentType = data.componentType as ComponentType

  const handleLabelChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onNodeDataChange(selectedNode.id, { ...data, label: e.target.value })
  }

  const handlePropertyChange = (key: string, value: unknown) => {
    onNodeDataChange(selectedNode.id, {
      ...data,
      properties: { ...properties, [key]: value },
    })
  }

  return (
    <aside
      style={{
        width: 300,
        padding: 16,
        borderLeft: '1px solid #e5e7eb',
        backgroundColor: '#f9fafb',
        overflowY: 'auto',
      }}
    >
      <h3
        style={{
          margin: '0 0 16px',
          fontSize: 16,
          fontWeight: 600,
          color: '#374151',
        }}
      >
        Properties
      </h3>

      <div style={{ marginBottom: 16 }}>
        <label
          style={{
            display: 'block',
            marginBottom: 4,
            fontSize: 13,
            fontWeight: 500,
            color: '#4b5563',
          }}
        >
          Label
        </label>
        <input
          type="text"
          value={(data.label as string) || ''}
          onChange={handleLabelChange}
          style={{
            width: '100%',
            padding: '6px 8px',
            border: '1px solid #d1d5db',
            borderRadius: 4,
            fontSize: 13,
          }}
        />
      </div>

      {componentType === 'database' && (
        <>
          <div style={{ marginBottom: 16 }}>
            <label
              style={{
                display: 'block',
                marginBottom: 4,
                fontSize: 13,
                fontWeight: 500,
                color: '#4b5563',
              }}
            >
              Database Type
            </label>
            <select
              value={(properties.dbType as string) || 'sql'}
              onChange={(e) => handlePropertyChange('dbType', e.target.value)}
              style={{
                width: '100%',
                padding: '6px 8px',
                border: '1px solid #d1d5db',
                borderRadius: 4,
                fontSize: 13,
                backgroundColor: '#fff',
              }}
            >
              <option value="sql">SQL</option>
              <option value="nosql">NoSQL</option>
              <option value="graph">Graph</option>
              <option value="timeseries">Time Series</option>
            </select>
          </div>
          <div style={{ marginBottom: 16 }}>
            <label
              style={{
                display: 'block',
                marginBottom: 4,
                fontSize: 13,
                fontWeight: 500,
                color: '#4b5563',
              }}
            >
              Read/Write Ratio ({(properties.readWriteRatio as number) || 0})
            </label>
            <input
              type="range"
              min="0"
              max="1"
              step="0.1"
              value={(properties.readWriteRatio as number) || 0}
              onChange={(e) =>
                handlePropertyChange('readWriteRatio', parseFloat(e.target.value))
              }
              style={{ width: '100%' }}
            />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label
              style={{
                display: 'block',
                marginBottom: 4,
                fontSize: 13,
                fontWeight: 500,
                color: '#4b5563',
              }}
            >
              Product
            </label>
            <input
              type="text"
              value={(properties.product as string) || ''}
              onChange={(e) => handlePropertyChange('product', e.target.value)}
              placeholder="e.g. cassandra, dynamodb"
              style={{
                width: '100%',
                padding: '6px 8px',
                border: '1px solid #d1d5db',
                borderRadius: 4,
                fontSize: 13,
              }}
            />
          </div>
        </>
      )}

      {componentType === 'service' && (
        <div style={{ marginBottom: 16 }}>
          <label
            style={{
              display: 'block',
              marginBottom: 4,
              fontSize: 13,
              fontWeight: 500,
              color: '#4b5563',
            }}
          >
            Replicas
          </label>
          <input
            type="number"
            min="1"
            value={(properties.replicas as number) || 1}
            onChange={(e) =>
              handlePropertyChange('replicas', parseInt(e.target.value, 10))
            }
            style={{
              width: '100%',
              padding: '6px 8px',
              border: '1px solid #d1d5db',
              borderRadius: 4,
              fontSize: 13,
            }}
          />
        </div>
      )}

      {componentType === 'load_balancer' && (
        <div style={{ marginBottom: 16 }}>
          <label
            style={{
              display: 'block',
              marginBottom: 4,
              fontSize: 13,
              fontWeight: 500,
              color: '#4b5563',
            }}
          >
            Algorithm
          </label>
          <select
            value={(properties.algorithm as string) || 'round_robin'}
            onChange={(e) => handlePropertyChange('algorithm', e.target.value)}
            style={{
              width: '100%',
              padding: '6px 8px',
              border: '1px solid #d1d5db',
              borderRadius: 4,
              fontSize: 13,
              backgroundColor: '#fff',
            }}
          >
            <option value="round_robin">Round Robin</option>
            <option value="least_connections">Least Connections</option>
            <option value="ip_hash">IP Hash</option>
            <option value="weighted">Weighted</option>
          </select>
        </div>
      )}
    </aside>
  )
}
