import type { SystemTopology, AnalyzeResponse } from '../types/topology'

const API_BASE = '/api'

export async function analyzeTopology(
  topology: SystemTopology
): Promise<AnalyzeResponse> {
  const response = await fetch(`${API_BASE}/topology`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(topology),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => null)
    throw new Error(
      errorData?.error ?? `Request failed with status ${response.status}`
    )
  }

  return response.json()
}
