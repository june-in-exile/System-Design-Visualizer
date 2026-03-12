# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ArchitectMind is a system design visualizer: an interactive canvas for building infrastructure diagrams with a backend that validates topology semantics. Users drag components (databases, load balancers, services, etc.) from a sidebar onto a React Flow canvas, connect them, and hit "Analyze" to send the graph to a Go backend for validation.

## Commands

### Backend (Go + Gin)

```bash
cd backend && go run main.go        # Start server on :8080
cd backend && go build ./...         # Build check
cd backend && go test ./...          # Run all tests
cd backend && go test ./handler/...  # Run tests for a single package
```

### Frontend (React + Vite + TypeScript)

```bash
cd frontend && npm install           # Install dependencies
cd frontend && npm run dev           # Dev server on :5173
cd frontend && npm run build         # Type-check (tsc -b) then bundle
cd frontend && npm run lint          # ESLint
```

Both servers must run simultaneously for the full workflow. Frontend calls backend at `http://localhost:8080/api`.

## Architecture

### Data Flow

1. **Sidebar** → user drags a component type onto the canvas (drag data: `application/architectmind` mime type)
2. **Canvas** → creates a React Flow node with type `"architecture"`, stores `componentType` + `properties` in node `data`
3. **Analyze button** → serializes all nodes/edges into a `SystemTopology` JSON payload, POSTs to `POST /api/topology`
4. **Backend handler** → deserializes, validates component/connection types, checks edge references, returns `AnalyzeResponse` with warnings

### Shared Contract (keep in sync manually)

- **Backend types**: `backend/model/topology.go` — `SystemTopology`, `SystemNode`, `SystemEdge`, `ValidComponentTypes`, `ValidConnectionTypes`
- **Backend properties**: `backend/model/properties.go` — typed property structs per component, `ParseNodeProperties` dispatcher
- **Frontend types**: `frontend/src/types/topology.ts` — mirrors the Go structs as TypeScript interfaces
- **Node config**: `frontend/src/nodes/nodeConfig.ts` — `NODE_TYPE_CONFIG` maps each `ComponentType` to label, color, icon, and default properties

When adding a new component type, update all four files: the Go `ValidComponentTypes` map, the Go property struct (if needed), the TS `ComponentType` union, and `NODE_TYPE_CONFIG`.

### Frontend Structure

- All canvas nodes use a single custom React Flow node type `"architecture"` rendered by `ArchitectureNode.tsx`
- Nodes carry `componentType` and `properties` in their `data` field, styled dynamically via `NODE_TYPE_CONFIG`
- Edge `data.connectionType` maps to backend `ConnectionType` values (`sync`, `async`, `replication`, `cdn_origin`)

### Backend Structure

- Single endpoint: `POST /api/topology` → `handler.PostTopology`
- Validation in `handler.validate()` checks component types, connection types, and dangling edge references
- CORS configured for `http://localhost:5173` only
- MongoDB driver is included in deps but not yet wired up
