import { toPng } from 'html-to-image'
import { jsPDF } from 'jspdf'
import { getNodesBounds } from '@xyflow/react'
import type { Node } from '@xyflow/react'

export async function exportPdf(
  nodes: Node[],
  isDarkMode: boolean,
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

  const dataUrl = await toPng(viewport, {
    width,
    height,
    backgroundColor: isDarkMode ? '#1f2937' : '#ffffff',
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

  pdf.setFontSize(16)
  pdf.setTextColor(isDarkMode ? '#f9fafb' : '#1f2937')
  pdf.text('Architecture Diagram', 50, 30)

  pdf.setFontSize(10)
  pdf.setTextColor(isDarkMode ? '#d1d5db' : '#4b5563')
  pdf.text(new Date().toLocaleString(), 50, 46)

  pdf.addImage(dataUrl, 'PNG', 50, 60, bounds.width, bounds.height)

  pdf.save(filename)
}
