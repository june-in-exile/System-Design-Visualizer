import type { Node, Edge } from '@xyflow/react'
import type { ComponentType } from '../types/topology'

function shapeFor(componentType: ComponentType, label: string): string {
  switch (componentType) {
    case 'client':
      return `(${label})`
    case 'database':
      return `[(${label})]`
    case 'cache':
      return `{{${label}}}`
    case 'message_queue':
      return `[/${label}/]`
    case 'cdn':
      return `([${label}])`
    case 'dns':
      return `([${label}])`
    case 'storage':
      return `[(${label})]`
    default:
      return `[${label}]`
  }
}

function sanitizeId(id: string): string {
  return id.replace(/[^a-zA-Z0-9_]/g, '_')
}

function buildEdgeLabel(edge: Edge): string {
  const data = (edge.data as Record<string, unknown>) ?? {}
  const protocol = data.protocol as string | undefined
  const connectionType = data.connectionType as string | undefined

  const parts: string[] = []
  if (protocol && protocol !== '' && protocol !== 'unspecified') {
    parts.push(protocol.toUpperCase())
  }
  if (connectionType && connectionType !== 'unspecified') {
    parts.push(connectionType)
  }
  return parts.join(' | ')
}

export function toMermaid(nodes: Node[], edges: Edge[]): string {
  let output = 'flowchart TD\n'

  for (const node of nodes) {
    const data = node.data as Record<string, unknown>
    const label = (data.label as string) || node.id
    const componentType = data.componentType as ComponentType
    const safeId = sanitizeId(node.id)
    output += `  ${safeId}${shapeFor(componentType, label)}\n`
  }

  output += '\n'

  for (const edge of edges) {
    const data = (edge.data as Record<string, unknown>) ?? {}
    const connectionType = data.connectionType as string | undefined
    const label = buildEdgeLabel(edge)
    const arrow = connectionType === 'async' ? '-.->' : '-->'
    const safeSource = sanitizeId(edge.source)
    const safeTarget = sanitizeId(edge.target)

    if (label) {
      output += `  ${safeSource} ${arrow}|${label}| ${safeTarget}\n`
    } else {
      output += `  ${safeSource} ${arrow} ${safeTarget}\n`
    }
  }

  return output
}

export async function copyMermaidToClipboard(nodes: Node[], edges: Edge[]): Promise<void> {
  const mermaidText = toMermaid(nodes, edges)
  await navigator.clipboard.writeText(mermaidText)
}

export function downloadMermaidFile(nodes: Node[], edges: Edge[], filename = 'architecture.mmd'): void {
  const mermaidText = toMermaid(nodes, edges)
  const blob = new Blob([mermaidText], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}
