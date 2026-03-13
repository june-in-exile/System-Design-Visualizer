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
