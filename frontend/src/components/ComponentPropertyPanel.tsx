import type { Node } from '@xyflow/react'
import type { ComponentType } from '../types/topology'

interface PropertyPanelProps {
  selectedNode: Node | null
  onNodeDataChange: (nodeId: string, newData: Record<string, unknown>) => void
}

const SQL_PRODUCTS = [
  { value: 'postgresql', label: 'PostgreSQL', description: 'Open-source relational database with strong ACID compliance' },
  { value: 'mysql', label: 'MySQL', description: 'Popular open-source relational database, widely used in web applications' },
  { value: 'mariadb', label: 'MariaDB', description: 'MySQL fork with enhanced performance and features' },
  { value: 'sql_server', label: 'SQL Server', description: 'Microsoft relational database with enterprise features' },
  { value: 'oracle', label: 'Oracle', description: 'Enterprise-grade relational database with advanced features' },
]

const NOSQL_GROUPS = [
  {
    label: 'Document',
    products: [
      { value: 'mongodb', label: 'MongoDB', description: 'Flexible document-oriented database with JSON-like storage' },
      { value: 'couchdb', label: 'CouchDB (AP)', description: 'Eventual consistency document database' },
    ],
  },
  {
    label: 'Key-Value / Wide Column',
    products: [
      { value: 'redis', label: 'Redis', description: 'In-memory data structure store, used as cache and message broker' },
      { value: 'dynamodb', label: 'DynamoDB (AP)', description: 'AWS fully managed NoSQL database with high scalability' },
      { value: 'cassandra', label: 'Cassandra (AP)', description: 'Distributed wide-column store optimized for writes' },
      { value: 'riak', label: 'Riak (AP)', description: 'Distributed key-value store focusing on availability' },
    ],
  },
  {
    label: 'Graph',
    products: [
      { value: 'neo4j', label: 'Neo4j', description: 'Native graph database optimized for relationship queries' },
      { value: 'arangodb', label: 'ArangoDB', description: 'Multi-model database supporting graphs, documents, and key-value' },
    ],
  },
  {
    label: 'Time Series',
    products: [
      { value: 'influxdb', label: 'InfluxDB', description: 'Time-series database optimized for monitoring and IoT data' },
      { value: 'prometheus', label: 'Prometheus', description: 'Monitoring system with built-in time-series database' },
    ],
  },
]

const CACHE_PRODUCTS = [
  { value: 'redis', label: 'Redis', description: 'In-memory data structure store with optional persistence' },
  { value: 'memcached', label: 'Memcached', description: 'Simple distributed memory object caching system' },
  { value: 'etcd', label: 'etcd', description: 'Distributed key-value store for shared configuration' },
  { value: 'hazelcast', label: 'Hazelcast', description: 'Distributed in-memory computing platform' },
  { value: 'tarantool', label: 'Tarantool', description: 'In-memory database with Lua application server' },
]

const EVICTION_POLICIES = [
  { value: 'lru', label: 'LRU (Least Recently Used)', description: 'Evicts least recently accessed items first' },
  { value: 'lfu', label: 'LFU (Least Frequently Used)', description: 'Evicts least frequently accessed items first' },
  { value: 'fifo', label: 'FIFO (First In First Out)', description: 'Evicts oldest items first, regardless of access pattern' },
  { value: 'tiny_lfu', label: 'TinyLFU', description: 'Frequency-based eviction with approximate LFU using tiny sketch' },
]

const CACHE_TYPES = [
  { value: 'in_memory', label: 'In-Memory', description: 'Data stored in RAM only, fastest but limited by memory' },
  { value: 'distributed', label: 'Distributed', description: 'Data partitioned across multiple nodes, scales horizontally' },
]

const DB_CATEGORIES = [
  { value: 'sql', label: 'SQL (Relational)', description: 'Structured tables with fixed schema, supports JOINs and ACID' },
  { value: 'nosql', label: 'NoSQL (Non-Relational)', description: 'Flexible schema, optimized for scale and specific use cases' },
]

const CONSISTENCY_LEVELS = [
  { value: 'eventual', label: 'Eventual Consistency', description: 'Data will be consistent eventually, prioritizes availability (AP)' },
  { value: 'strong', label: 'Strong Consistency', description: 'Guarantees immediate consistency, may impact availability during partitions (CP)' },
]

const STORAGE_ACCESS_LEVELS = [
  { value: 'public', label: 'Public', description: 'Publicly accessible, suitable for static websites and CDN origins' },
  { value: 'private', label: 'Private', description: 'Private access only, suitable for internal data and backups' },
]

const STORAGE_CLASSES = [
  { value: 'standard', label: 'Standard', description: 'Frequently accessed data with low latency requirements' },
  { value: 'infrequent_access', label: 'Infrequent Access', description: 'Long-lived, less frequently accessed data at lower cost' },
  { value: 'glacier', label: 'Glacier (Archive)', description: 'Cold storage for archival data with retrieval delays' },
]

const MQ_CATEGORIES = [
  { value: 'broker', label: 'Broker-based', description: 'Central broker handles message routing, storage, and delivery' },
  { value: 'log_streaming', label: 'Log Streaming', description: 'Append-only distributed log with consumer-managed offsets' },
  { value: 'in_memory', label: 'In-Memory', description: 'Lightweight, ultra-fast messaging using memory-based stores' },
  { value: 'managed', label: 'Cloud Managed', description: 'Fully managed cloud services with auto-scaling' },
]

const MQ_PRODUCT_GROUPS: Record<string, Array<{ value: string; label: string; description: string }>> = {
  broker: [
    { value: 'rabbitmq', label: 'RabbitMQ', description: 'Feature-rich AMQP broker with complex routing (Exchange/Binding)' },
    { value: 'activemq', label: 'ActiveMQ', description: 'Java-based broker supporting JMS, STOMP, and multiple protocols' },
  ],
  log_streaming: [
    { value: 'kafka', label: 'Apache Kafka', description: 'High-throughput distributed log with replay and stream processing' },
    { value: 'pulsar', label: 'Apache Pulsar', description: 'Cloud-native streaming with separated storage and compute' },
  ],
  in_memory: [
    { value: 'redis_mq', label: 'Redis (Pub/Sub / Streams)', description: 'Ultra-fast in-memory messaging, simple deployment' },
  ],
  managed: [
    { value: 'aws_sqs', label: 'AWS SQS', description: 'Fully managed queue service deeply integrated with AWS ecosystem' },
    { value: 'aws_sns', label: 'AWS SNS', description: 'Managed pub/sub notification service for fan-out messaging' },
    { value: 'google_pubsub', label: 'Google Pub/Sub', description: 'Fully managed real-time messaging with auto-scaling' },
    { value: 'azure_service_bus', label: 'Azure Service Bus', description: 'Enterprise messaging for .NET ecosystem with advanced features' },
  ],
}

const MQ_QUEUE_TYPES = [
  { value: 'point_to_point', label: 'Point-to-Point', description: 'Each message consumed by exactly one consumer' },
  { value: 'pub_sub', label: 'Pub/Sub', description: 'Messages broadcast to all subscribed consumers' },
]

const MQ_DELIVERY_GUARANTEES = [
  { value: 'at_most_once', label: 'At Most Once', description: 'Fire-and-forget, fastest but may lose messages' },
  { value: 'at_least_once', label: 'At Least Once', description: 'Guaranteed delivery, consumers must handle duplicates' },
  { value: 'exactly_once', label: 'Exactly Once', description: 'Strongest guarantee, higher latency and complexity' },
]

const ALGORITHMS = [
  { value: 'round_robin', label: 'Round Robin', description: 'Distributes requests evenly across all servers in sequence' },
  { value: 'least_connections', label: 'Least Connections', description: 'Routes to server with fewest active connections' },
  { value: 'ip_hash', label: 'IP Hash', description: 'Uses client IP to consistently route to same server' },
  { value: 'weighted', label: 'Weighted', description: 'Distributes based on server capacity weights' },
]

const getTooltip = (label: string, description?: string) => {
  const lowerLabel = label.toLowerCase()
  if (!description || lowerLabel.includes('unspecified') || lowerLabel.includes('default') || lowerLabel.includes('auto-detect')) {
    return undefined
  }
  return description
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

  const handleMQCategoryChange = (newCategory: string) => {
    const products = MQ_PRODUCT_GROUPS[newCategory] ?? []
    const firstProduct = products.length > 0 ? products[0].value : ''

    onNodeDataChange(selectedNode.id, {
      ...data,
      properties: { ...properties, category: newCategory, product: firstProduct },
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
        Component Properties
      </h3>

      {/* Label - Common for all nodes */}
      <div style={{ marginBottom: 16 }}>
        <label 
          style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
          title="The display name of this component on the canvas."
        >
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
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Choose between relational (SQL) or non-relational (NoSQL) database types."
            >
              Database Category
            </label>
            <select
              value={(properties.dbType as string) || 'sql'}
              onChange={(e) => handleDBTypeChange(e.target.value)}
              title={getTooltip(
                DB_CATEGORIES.find(opt => opt.value === ((properties.dbType as string) || 'sql'))?.label || '',
                DB_CATEGORIES.find(opt => opt.value === ((properties.dbType as string) || 'sql'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {DB_CATEGORIES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          {/* Product Selection - Filtered by DB Type */}
          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The specific database product or engine to use."
            >
              Product
            </label>
            <select
              value={(properties.product as string) || ''}
              onChange={(e) => handlePropertyChange('product', e.target.value)}
              title={properties.dbType === 'sql' 
                ? getTooltip(
                    SQL_PRODUCTS.find(p => p.value === (properties.product as string))?.label || '',
                    SQL_PRODUCTS.find(p => p.value === (properties.product as string))?.description
                  )
                : getTooltip(
                    NOSQL_GROUPS.flatMap(g => g.products).find(p => p.value === (properties.product as string))?.label || '',
                    NOSQL_GROUPS.flatMap(g => g.products).find(p => p.value === (properties.product as string))?.description
                  )
              }
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {properties.dbType === 'sql' ? (
                SQL_PRODUCTS.map((p) => (
                  <option key={p.value} value={p.value} title={p.description}>{p.label}</option>
                ))
              ) : (
                NOSQL_GROUPS.map((group) => (
                  <optgroup key={group.label} label={group.label}>
                    {group.products.map((p) => (
                      <option key={p.value} value={p.value} title={p.description}>{p.label}</option>
                    ))}
                  </optgroup>
                ))
              )}
            </select>
          </div>

          {/* Consistency Level - Added for CAP Theorem Rule 5 */}
          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Desired data consistency model (CAP theorem trade-offs)."
            >
              Consistency Requirement
            </label>
            <select
              value={(properties.consistencyLevel as string) || ''}
              onChange={(e) => handlePropertyChange('consistencyLevel', e.target.value)}
              title={getTooltip(
                CONSISTENCY_LEVELS.find(opt => opt.value === (properties.consistencyLevel as string))?.label || 'Default',
                CONSISTENCY_LEVELS.find(opt => opt.value === (properties.consistencyLevel as string))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              <option value="">Default (Auto-detect)</option>
              {CONSISTENCY_LEVELS.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
              Explicitly choosing a level acknowledges the CAP trade-offs.
            </p>
          </div>

          {/* Read/Write Ratio - Only shown for SQL (Rule 2) */}
          {properties.dbType === 'sql' && (
            <div style={{ marginBottom: 16 }}>
              <label 
                style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
                title="The estimated balance between read and write operations. High read ratios favor caching and replicas."
              >
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

      {['service', 'database', 'load_balancer', 'cache'].includes(componentType) && (
        <div style={{ marginBottom: 16 }}>
          <label 
            style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
            title="Number of running instances for high availability and load distribution."
          >
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
          <label 
            style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
            title="The policy used to distribute incoming traffic across healthy backend instances."
          >
            Algorithm
          </label>
          <select
            value={(properties.algorithm as string) || 'round_robin'}
            onChange={(e) => handlePropertyChange('algorithm', e.target.value)}
            title={getTooltip(
              ALGORITHMS.find(opt => opt.value === ((properties.algorithm as string) || 'round_robin'))?.label || '',
              ALGORITHMS.find(opt => opt.value === ((properties.algorithm as string) || 'round_robin'))?.description
            )}
            style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
          >
            {ALGORITHMS.map((opt) => (
              <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
            ))}
          </select>
        </div>
      )}

      {componentType === 'storage' && (
        <>
          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Visibility and permission level for the stored objects."
            >
              Access Level
            </label>
            <select
              value={(properties.accessLevel as string) || 'private'}
              onChange={(e) => handlePropertyChange('accessLevel', e.target.value)}
              title={getTooltip(
                STORAGE_ACCESS_LEVELS.find(opt => opt.value === ((properties.accessLevel as string) || 'private'))?.label || '',
                STORAGE_ACCESS_LEVELS.find(opt => opt.value === ((properties.accessLevel as string) || 'private'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {STORAGE_ACCESS_LEVELS.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Data storage tier, balancing access frequency, latency, and cost."
            >
              Storage Class
            </label>
            <select
              value={(properties.storageClass as string) || 'standard'}
              onChange={(e) => handlePropertyChange('storageClass', e.target.value)}
              title={getTooltip(
                STORAGE_CLASSES.find(opt => opt.value === ((properties.storageClass as string) || 'standard'))?.label || '',
                STORAGE_CLASSES.find(opt => opt.value === ((properties.storageClass as string) || 'standard'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {STORAGE_CLASSES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)', cursor: 'pointer' }}
              title="Preserve multiple versions of an object to protect against accidental deletions or overwrites."
            >
              <input
                type="checkbox"
                checked={(properties.versioning as boolean) || false}
                onChange={(e) => handlePropertyChange('versioning', e.target.checked)}
              />
              Enable Versioning
            </label>
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
              Keep multiple versions of objects for recovery and audit.
            </p>
          </div>
        </>
      )}

      {componentType === 'message_queue' && (
        <>
          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The architectural pattern of the messaging system."
            >
              Category
            </label>
            <select
              value={(properties.category as string) || 'broker'}
              onChange={(e) => handleMQCategoryChange(e.target.value)}
              title={getTooltip(
                MQ_CATEGORIES.find(opt => opt.value === ((properties.category as string) || 'broker'))?.label || '',
                MQ_CATEGORIES.find(opt => opt.value === ((properties.category as string) || 'broker'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {MQ_CATEGORIES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The specific messaging product or managed service."
            >
              Product
            </label>
            <select
              value={(properties.product as string) || ''}
              onChange={(e) => handlePropertyChange('product', e.target.value)}
              title={getTooltip(
                (MQ_PRODUCT_GROUPS[(properties.category as string) || 'broker'] ?? []).find(p => p.value === (properties.product as string))?.label || '',
                (MQ_PRODUCT_GROUPS[(properties.category as string) || 'broker'] ?? []).find(p => p.value === (properties.product as string))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {(MQ_PRODUCT_GROUPS[(properties.category as string) || 'broker'] ?? []).map((p) => (
                <option key={p.value} value={p.value} title={p.description}>{p.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Defines how messages are distributed to consumers (one-to-one vs one-to-many)."
            >
              Queue Type
            </label>
            <select
              value={(properties.queueType as string) || 'pub_sub'}
              onChange={(e) => handlePropertyChange('queueType', e.target.value)}
              title={getTooltip(
                MQ_QUEUE_TYPES.find(opt => opt.value === ((properties.queueType as string) || 'pub_sub'))?.label || '',
                MQ_QUEUE_TYPES.find(opt => opt.value === ((properties.queueType as string) || 'pub_sub'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {MQ_QUEUE_TYPES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The reliability guarantee for message handoff between producer and consumer."
            >
              Delivery Guarantee
            </label>
            <select
              value={(properties.deliveryGuarantee as string) || 'at_least_once'}
              onChange={(e) => handlePropertyChange('deliveryGuarantee', e.target.value)}
              title={getTooltip(
                MQ_DELIVERY_GUARANTEES.find(opt => opt.value === ((properties.deliveryGuarantee as string) || 'at_least_once'))?.label || '',
                MQ_DELIVERY_GUARANTEES.find(opt => opt.value === ((properties.deliveryGuarantee as string) || 'at_least_once'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {MQ_DELIVERY_GUARANTEES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)', cursor: 'pointer' }}
              title="Ensures that messages are processed in the exact order they were sent."
            >
              <input
                type="checkbox"
                checked={(properties.ordered as boolean) || false}
                onChange={(e) => handlePropertyChange('ordered', e.target.checked)}
              />
              Ordered Delivery
            </label>
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
              Guarantee message ordering within a partition or queue.
            </p>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)', cursor: 'pointer' }}
              title="A secondary queue for messages that cannot be processed, allowing for offline inspection and manual retry."
            >
              <input
                type="checkbox"
                checked={(properties.hasDLQ as boolean) || false}
                onChange={(e) => handlePropertyChange('hasDLQ', e.target.checked)}
              />
              Enable Dead Letter Queue (DLQ)
            </label>
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
              Automatically move failed messages to a separate queue for inspection.
            </p>
          </div>
        </>
      )}

      {componentType === 'cache' && (
        <>
          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Defines whether the cache is local to an instance or shared across a cluster."
            >
              Cache Type
            </label>
            <select
              value={(properties.cacheType as string) || 'distributed'}
              onChange={(e) => handlePropertyChange('cacheType', e.target.value)}
              title={getTooltip(
                CACHE_TYPES.find(opt => opt.value === ((properties.cacheType as string) || 'distributed'))?.label || '',
                CACHE_TYPES.find(opt => opt.value === ((properties.cacheType as string) || 'distributed'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {CACHE_TYPES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The specific caching engine or product."
            >
              Product
            </label>
            <select
              value={(properties.product as string) || 'redis'}
              onChange={(e) => handlePropertyChange('product', e.target.value)}
              title={getTooltip(
                CACHE_PRODUCTS.find(p => p.value === ((properties.product as string) || 'redis'))?.label || '',
                CACHE_PRODUCTS.find(p => p.value === ((properties.product as string) || 'redis'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {CACHE_PRODUCTS.map((p) => (
                <option key={p.value} value={p.value} title={p.description}>{p.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="The logic used to decide which items to remove when the cache is full."
            >
              Eviction Policy
            </label>
            <select
              value={(properties.evictionPolicy as string) || 'lru'}
              onChange={(e) => handlePropertyChange('evictionPolicy', e.target.value)}
              title={getTooltip(
                EVICTION_POLICIES.find(opt => opt.value === ((properties.evictionPolicy as string) || 'lru'))?.label || '',
                EVICTION_POLICIES.find(opt => opt.value === ((properties.evictionPolicy as string) || 'lru'))?.description
              )}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              {EVICTION_POLICIES.map((opt) => (
                <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div style={{ marginBottom: 16 }}>
            <label 
              style={{ display: 'block', marginBottom: 4, fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}
              title="Time-To-Live: Duration after which a cache entry is automatically invalidated."
            >
              TTL (Seconds)
            </label>
            <input
              type="number"
              min="0"
              placeholder="0 (No Expiration)"
              value={(properties.ttlSeconds as number) || 0}
              onChange={(e) => handlePropertyChange('ttlSeconds', parseInt(e.target.value, 10))}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid var(--border-color)', borderRadius: 4, fontSize: 13, backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--text-secondary)' }}>
              Setting a TTL helps maintain data consistency by expiring old cache entries.
            </p>
          </div>
        </>
      )}
    </aside>
  )
}
