import type { Node } from '@xyflow/react'
import type { ComponentType } from '../types/topology'

interface PropertyPanelProps {
  selectedNode: Node | null
  onNodeDataChange: (nodeId: string, newData: Record<string, unknown>) => void
}

const SQL_PRODUCTS = [
  { value: 'postgresql', label: 'PostgreSQL' },
  { value: 'mysql', label: 'MySQL' },
  { value: 'mariadb', label: 'MariaDB' },
  { value: 'sql_server', label: 'SQL Server' },
  { value: 'oracle', label: 'Oracle' },
]

const NOSQL_GROUPS = [
  {
    label: 'Document',
    products: [
      { value: 'mongodb', label: 'MongoDB' },
      { value: 'couchdb', label: 'CouchDB (AP)' },
    ],
  },
  {
    label: 'Key-Value / Wide Column',
    products: [
      { value: 'redis', label: 'Redis' },
      { value: 'dynamodb', label: 'DynamoDB (AP)' },
      { value: 'cassandra', label: 'Cassandra (AP)' },
      { value: 'riak', label: 'Riak (AP)' },
    ],
  },
  {
    label: 'Graph',
    products: [
      { value: 'neo4j', label: 'Neo4j' },
      { value: 'arangodb', label: 'ArangoDB' },
    ],
  },
  {
    label: 'Time Series',
    products: [
      { value: 'influxdb', label: 'InfluxDB' },
      { value: 'prometheus', label: 'Prometheus' },
    ],
  },
]

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

  const handleDBTypeChange = (newType: string) => {
    // 切換類型時，自動選取該類型下的第一個產品以維持一致性
    const firstProduct = newType === 'sql' 
      ? SQL_PRODUCTS[0].value 
      : NOSQL_GROUPS[0].products[0].value

    onNodeDataChange(selectedNode.id, {
      ...data,
      properties: { ...properties, dbType: newType, product: firstProduct },
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
      <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 600, color: 'var(--text-primary)' }}>
        Properties
      </h3>

      {/* Label - Common for all nodes */}
      <div style={{ marginBottom: 16 }}>
        <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
          Label
        </label>
        <input
          type="text"
          value={(data.label as string) || ''}
          onChange={handleLabelChange}
          style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        />
      </div>

      {componentType === 'database' && (
        <>
          {/* Database Category */}
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
              Database Category
            </label>
            <select
              value={(properties.dbType as string) || 'sql'}
              onChange={(e) => handleDBTypeChange(e.target.value)}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              <option value="sql">SQL (Relational)</option>
              <option value="nosql">NoSQL (Non-Relational)</option>
            </select>
          </div>

          {/* Product Selection - Filtered by DB Type */}
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
              Product
            </label>
            <select
              value={(properties.product as string) || ''}
              onChange={(e) => handlePropertyChange('product', e.target.value)}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {properties.dbType === 'sql' ? (
                SQL_PRODUCTS.map((p) => (
                  <option key={p.value} value={p.value}>{p.label}</option>
                ))
              ) : (
                NOSQL_GROUPS.map((group) => (
                  <optgroup key={group.label} label={group.label}>
                    {group.products.map((p) => (
                      <option key={p.value} value={p.value}>{p.label}</option>
                    ))}
                  </optgroup>
                ))
              )}
            </select>
          </div>

          {/* Read/Write Ratio - Only shown for SQL (Rule 2) */}
          {properties.dbType === 'sql' && (
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
                Read/Write Ratio ({(properties.readWriteRatio as number) || 0})
              </label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={(properties.readWriteRatio as number) || 0}
                onChange={(e) => handlePropertyChange('readWriteRatio', parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
              <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
                Lower values mean higher write pressure.
              </p>
            </div>
          )}
        </>
      )}

      {componentType === 'service' && (
        <div style={{ marginBottom: 16 }}>
          <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
            Replicas
          </label>
          <input
            type="number"
            min="1"
            value={(properties.replicas as number) || 1}
            onChange={(e) => handlePropertyChange('replicas', parseInt(e.target.value, 10))}
            style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
          />
        </div>
      )}

      {componentType === 'load_balancer' && (
        <div style={{ marginBottom: 16 }}>
          <label style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>
            Algorithm
          </label>
          <select
            value={(properties.algorithm as string) || 'round_robin'}
            onChange={(e) => handlePropertyChange('algorithm', e.target.value)}
            style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
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
