import { useState } from 'react'
import type { Node } from '@xyflow/react'
import type { ComponentType } from '../types/topology'

const COMMON_PRODUCTS = [
  'MySQL',
  'PostgreSQL',
  'MongoDB',
  'Cassandra',
  'DynamoDB',
  'Redis',
  'Elasticsearch',
  'Neo4j',
  'InfluxDB',
]

interface PropertyPanelProps {
  selectedNode: Node | null
  onNodeDataChange: (nodeId: string, newData: Record<string, unknown>) => void
}

export default function PropertyPanel({
  selectedNode,
  onNodeDataChange,
}: PropertyPanelProps) {
  const [showOtherInput, setShowOtherInput] = useState(false)

  if (!selectedNode) {
    return null
  }

  const data = selectedNode.data as Record<string, unknown>
  const properties = (data.properties as Record<string, unknown>) ?? {}
  const componentType = data.componentType as ComponentType

  const currentProduct = (properties.product as string) || ''
  const isOtherProduct = showOtherInput || (currentProduct !== '' && !COMMON_PRODUCTS.includes(currentProduct))

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
        borderLeft: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        overflowY: 'auto',
      }}
    >
      <h3
        style={{
          margin: '0 0 16px',
          fontSize: 16,
          fontWeight: 600,
          color: 'var(--text-primary)',
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
            color: 'var(--text-secondary)',
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
            border: '1px solid var(--border-color)',
            borderRadius: 4,
            fontSize: 13,
            backgroundColor: 'var(--bg-primary)',
            color: 'var(--text-primary)',
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
                color: 'var(--text-secondary)',
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
                border: '1px solid var(--border-color)',
                borderRadius: 4,
                fontSize: 13,
                backgroundColor: 'var(--bg-primary)',
                color: 'var(--text-primary)',
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
                color: 'var(--text-secondary)',
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
                color: 'var(--text-secondary)',
              }}
            >
              Product
            </label>
            <select
              value={isOtherProduct ? '___other___' : currentProduct}
              onChange={(e) => {
                const val = e.target.value
                if (val === '___other___') {
                  setShowOtherInput(true)
                  if (COMMON_PRODUCTS.includes(currentProduct)) {
                    handlePropertyChange('product', '')
                  }
                } else {
                  setShowOtherInput(false)
                  handlePropertyChange('product', val)
                }
              }}
              style={{
                width: '100%',
                padding: '6px 8px',
                border: '1px solid var(--border-color)',
                borderRadius: 4,
                fontSize: 13,
                backgroundColor: 'var(--bg-primary)',
                color: 'var(--text-primary)',
                marginBottom: isOtherProduct ? 8 : 0,
              }}
            >
              <option value="">Select a product...</option>
              {COMMON_PRODUCTS.map(product => (
                <option key={product} value={product}>{product}</option>
              ))}
              <option value="___other___">其它 / Other</option>
            </select>
            
            {isOtherProduct && (
              <input
                type="text"
                value={currentProduct}
                onChange={(e) => handlePropertyChange('product', e.target.value)}
                placeholder="Enter custom product name"
                style={{
                  width: '100%',
                  padding: '6px 8px',
                  border: '1px solid var(--border-color)',
                  borderRadius: 4,
                  fontSize: 13,
                  backgroundColor: 'var(--bg-primary)',
                  color: 'var(--text-primary)',
                }}
              />
            )}
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
              color: 'var(--text-secondary)',
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
              border: '1px solid var(--border-color)',
              borderRadius: 4,
              fontSize: 13,
              backgroundColor: 'var(--bg-primary)',
              color: 'var(--text-primary)',
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
              color: 'var(--text-secondary)',
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
              border: '1px solid var(--border-color)',
              borderRadius: 4,
              fontSize: 13,
              backgroundColor: 'var(--bg-primary)',
              color: 'var(--text-primary)',
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
