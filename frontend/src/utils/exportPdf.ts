import { toPng } from 'html-to-image'
import { jsPDF } from 'jspdf'
import { getNodesBounds } from '@xyflow/react'
import type { Node } from '@xyflow/react'

export async function exportPdf(
  nodes: Node[],
  theme: 'light' | 'dark' | 'warm' | 'dream' | 'cyberpunk',
  filename = 'architecture.pdf'
): Promise<void> {
  const viewport = document.querySelector('.react-flow__viewport') as HTMLElement | null
  if (!viewport) {
    throw new Error('React Flow viewport not found')
  }

  const bounds = getNodesBounds(nodes)
  const padding = 50
  const width = bounds.width + padding * 2
  const height = bounds.height + padding * 2

  const backgroundColor = theme === 'dark' ? '#1f2937' : theme === 'warm' ? '#EBE4D1' : theme === 'dream' ? '#F5F3FF' : theme === 'cyberpunk' ? '#050505' : '#ffffff'

  const dataUrl = await toPng(viewport, {
    width,
    height,
    backgroundColor,
    pixelRatio: 2,
    style: {
      transform: `translate(${-bounds.x + padding}px, ${-bounds.y + padding}px)`,
    },
  })

  const isLandscape = bounds.width > bounds.height
  const pdf = new jsPDF({
    orientation: isLandscape ? 'landscape' : 'portrait',
    unit: 'px',
    format: [width + 100, height + 100],
  })

  const textColor = theme === 'dark' ? '#f9fafb' : theme === 'warm' ? '#433422' : theme === 'dream' ? '#312E81' : theme === 'cyberpunk' ? '#00f3ff' : '#1f2937'
  const subTextColor = theme === 'dark' ? '#d1d5db' : theme === 'warm' ? '#6B543D' : theme === 'dream' ? '#4338CA' : theme === 'cyberpunk' ? '#ff0055' : '#4b5563'

  pdf.setFontSize(16)
  pdf.setTextColor(textColor)
  pdf.text('Architecture Diagram', 50, 30)

  pdf.setFontSize(10)
  pdf.setTextColor(subTextColor)
  pdf.text(new Date().toLocaleString(), 50, 46)

  pdf.addImage(dataUrl, 'PNG', 50, 60, bounds.width, bounds.height)

  pdf.save(filename)
}
