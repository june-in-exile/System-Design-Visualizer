import type { ComponentType, ComponentProperties } from '../types/topology'

interface NodeTypeConfig {
  label: string
  color: string
  icon: string
  defaultProperties: ComponentProperties
}

export const NODE_TYPE_CONFIG: Record<ComponentType, NodeTypeConfig> = {
  client: {
    label: 'Client',
    color: '#6366f1',
    icon: '🖥️',
    defaultProperties: {},
  },
  load_balancer: {
    label: 'Load Balancer',
    color: '#f59e0b',
    icon: '⚖️',
    defaultProperties: {
      algorithm: 'round_robin',
      healthCheck: true,
      layer: 'l7',
    },
  },
  api_gateway: {
    label: 'API Gateway',
    color: '#10b981',
    icon: '🚪',
    defaultProperties: {},
  },
  service: {
    label: 'Service',
    color: '#3b82f6',
    icon: '⚙️',
    defaultProperties: {
      replicas: 1,
      stateless: true,
    },
  },
  database: {
    label: 'Database',
    color: '#ef4444',
    icon: '🗄️',
    defaultProperties: {
      dbType: 'sql',
      acidRequired: true,
      readWriteRatio: 0.8,
      scalingStrategy: 'vertical',
    },
  },
  cache: {
    label: 'Cache',
    color: '#8b5cf6',
    icon: '⚡',
    defaultProperties: {
      cacheType: 'distributed',
      product: 'redis',
      evictionPolicy: 'lru',
      ttlSeconds: 3600,
    },
  },
  message_queue: {
    label: 'Message Queue',
    color: '#ec4899',
    icon: '📨',
    defaultProperties: {
      queueType: 'pub_sub',
      deliveryGuarantee: 'at_least_once',
      ordered: false,
    },
  },
  cdn: {
    label: 'CDN',
    color: '#14b8a6',
    icon: '🌐',
    defaultProperties: {},
  },
  dns: {
    label: 'DNS',
    color: '#64748b',
    icon: '📡',
    defaultProperties: {},
  },
  storage: {
    label: 'Storage',
    color: '#f97316',
    icon: '💾',
    defaultProperties: {
      accessLevel: 'private',
      storageClass: 'standard',
      versioning: false,
    },
  },
}
