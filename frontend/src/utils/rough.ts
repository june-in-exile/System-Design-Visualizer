import rough from 'roughjs'

export interface RoughOptions {
  stroke: string
  strokeWidth?: number
  fill?: string
  fillStyle?: 'hachure' | 'solid' | 'zigzag' | 'cross-hatch' | 'dots'
  roughness?: number
  seed?: number
}

/**
 * Produce a stable seed from a string id so rough.js
 * renders the same hand-drawn shape on every re-render.
 */
export function stableSeed(id: string): number {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = ((hash << 5) - hash + id.charCodeAt(i)) | 0
  }
  return Math.abs(hash)
}

export { rough }
