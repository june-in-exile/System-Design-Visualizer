import type { ComponentType, ComponentProperties } from '../types/topology'

interface NodeTypeConfig {
  label: string
  color: string
  icon: string
  description: string
  defaultProperties: ComponentProperties
}

export const NODE_TYPE_CONFIG: Record<ComponentType, NodeTypeConfig> = {
  client: {
    label: 'Client',
    color: '#6366f1',
    icon: '🖥️',
    description: 'Represents an end-user device, web browser, or external consumer of the system.',
    defaultProperties: {},
  },
  dns: {
    label: 'DNS',
    color: '#64748b',
    icon: '📡',
    description: 'Domain Name System: Translates human-readable domain names to IP addresses.',
    defaultProperties: {},
  },
  cdn: {
    label: 'CDN',
    color: '#14b8a6',
    icon: '🌐',
    description: 'Content Delivery Network: Geographically distributed group of servers which work together to provide fast delivery of Internet content.',
    defaultProperties: {},
  },
  load_balancer: {
    label: 'Load Balancer',
    color: '#f59e0b',
    icon: '⚖️',
    description: 'Distributes incoming network traffic across a group of backend servers to ensure availability and reliability.',
    defaultProperties: {
      algorithm: 'round_robin',
      healthCheck: true,
      layer: 'l7',
    },
  },
  reverse_proxy: {
    label: 'Reverse Proxy',
    color: '#0ea5e9',
    icon: '🔀',
    description: 'Sits between clients and backend servers, handling SSL termination, caching, compression, and request routing.',
    defaultProperties: {
      product: 'nginx',
      sslTermination: true,
      caching: false,
      compression: false,
      rateLimiting: false,
    },
  },
  api_gateway: {
    label: 'API Gateway',
    color: '#10b981',
    icon: '🚪',
    description: 'A single entry point for all clients, handling request routing, composition, and protocol translation.',
    defaultProperties: {},
  },
  service: {
    label: 'Service',
    color: '#3b82f6',
    icon: '⚙️',
    description: 'A generic compute unit, microservice, or application server that performs business logic.',
    defaultProperties: {
      replicas: 1,
      stateless: true,
    },
  },
  message_queue: {
    label: 'Message Queue',
    color: '#ec4899',
    icon: '📨',
    description: 'Provides an asynchronous communication protocol for inter-service messaging.',
    defaultProperties: {
      category: 'broker',
      product: 'rabbitmq',
      queueType: 'pub_sub',
      deliveryGuarantee: 'at_least_once',
      ordered: false,
      hasDLQ: false,
    },
  },
  cache: {
    label: 'Cache',
    color: '#8b5cf6',
    icon: '⚡',
    description: 'High-speed data storage layer which stores a subset of data for faster retrieval.',
    defaultProperties: {
      cacheType: 'distributed',
      product: 'redis',
      evictionPolicy: 'lru',
      ttlSeconds: 3600,
    },
  },
  database: {
    label: 'Database',
    color: '#ef4444',
    icon: '🗄️',
    description: 'A structured system for storing, managing, and retrieving data.',
    defaultProperties: {
      dbType: 'sql',
      acidRequired: true,
      readWriteRatio: 0.8,
      scalingStrategy: 'vertical',
    },
  },
  storage: {
    label: 'Storage',
    color: '#f97316',
    icon: '💾',
    description: 'Object or block storage for unstructured data like images, logs, and backups.',
    defaultProperties: {
      accessLevel: 'private',
      storageClass: 'standard',
      versioning: false,
    },
  },
  
}

export function getMergedConfig(roles: ComponentType[]): {
  label: string
  icons: string[]
  colors: string[]
  descriptions: string[]
} {
  return {
    label: roles.map(r => NODE_TYPE_CONFIG[r].label).join(' + '),
    icons: roles.map(r => NODE_TYPE_CONFIG[r].icon),
    colors: roles.map(r => NODE_TYPE_CONFIG[r].color),
    descriptions: roles.map(r => NODE_TYPE_CONFIG[r].description),
  }
}
