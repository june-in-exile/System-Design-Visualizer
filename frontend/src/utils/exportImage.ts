import { toBlob } from 'html-to-image'
import { saveAs } from 'file-saver'
import { getNodesBounds } from '@xyflow/react'
import type { Node } from '@xyflow/react'

export async function exportPng(
  nodes: Node[],
  theme: 'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk',
  filename = 'architecture.png'
): Promise<void> {
  const viewport = document.querySelector('.react-flow__viewport') as HTMLElement | null
  if (!viewport) {
    throw new Error('React Flow viewport not found')
  }

  const bounds = getNodesBounds(nodes)
  const padding = 50

  const backgroundColor = theme === 'dark' ? '#1f2937' : theme === 'warm' ? '#EBE4D1' : theme === 'dream' ? '#F5F3FF' : theme === 'cyberpunk' ? '#050505' : '#ffffff'

  const blob = await toBlob(viewport, {
    width: bounds.width + padding * 2,
    height: bounds.height + padding * 2,
    backgroundColor,
    pixelRatio: 2,
    style: {
      transform: `translate(${-bounds.x + padding}px, ${-bounds.y + padding}px)`,
    },
  })

  if (!blob) {
    throw new Error('Failed to generate PNG')
  }

  saveAs(blob, filename)
}
