export type ComponentType =
  | 'client'
  | 'load_balancer'
  | 'api_gateway'
  | 'service'
  | 'database'
  | 'cache'
  | 'message_queue'
  | 'cdn'
  | 'dns'
  | 'storage'

export type ConnectionType = 'sync' | 'async' | 'replication' | 'cdn_origin'

export interface DatabaseProperties {
  dbType: 'sql' | 'nosql' | 'graph' | 'timeseries'
  acidRequired: boolean
  readWriteRatio: number
  scalingStrategy: 'vertical' | 'horizontal' | 'none'
  replicationFactor?: number
  consistencyLevel?: 'strong' | 'eventual'
}

export interface LoadBalancerProperties {
  algorithm: 'round_robin' | 'least_connections' | 'ip_hash' | 'weighted'
  healthCheck: boolean
  layer: 'l4' | 'l7'
}

export interface CacheProperties {
  cacheType: 'in_memory' | 'distributed'
  evictionPolicy: 'lru' | 'lfu' | 'ttl'
  ttlSeconds?: number
}

export interface MessageQueueProperties {
  queueType: 'point_to_point' | 'pub_sub'
  deliveryGuarantee: 'at_most_once' | 'at_least_once' | 'exactly_once'
  ordered: boolean
}

export interface ServiceProperties {
  replicas: number
  stateless: boolean
}

export type ComponentProperties =
  | DatabaseProperties
  | LoadBalancerProperties
  | CacheProperties
  | MessageQueueProperties
  | ServiceProperties
  | Record<string, unknown>

export interface SystemNode {
  id: string
  componentType: ComponentType
  label: string
  position: { x: number; y: number }
  properties: ComponentProperties
}

export interface SystemEdge {
  id: string
  source: string
  target: string
  connectionType: ConnectionType
  protocol?: string
  label?: string
}

export interface SystemTopology {
  id: string
  name: string
  version: number
  nodes: SystemNode[]
  edges: SystemEdge[]
}

export interface AnalyzeResponse {
  success: boolean
  nodeCount: number
  edgeCount: number
  warnings?: string[]
}
