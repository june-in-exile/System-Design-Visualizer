# CLAUDE.md

ArchitectMind is a system design visualizer with a React Flow canvas and a Go Gin backend for architecture validation.

## Commands

### Backend (Go + Gin)

- `cd api && go run _cmd/main.go` - Start server on :8080
- `cd api && go build ./...` - Build check
- `cd api && go test ./...` - Run all tests

### Frontend (React + Vite + TypeScript)

- `cd frontend && npm install` - Install dependencies
- `cd frontend && npm run dev` - Dev server on :5173
- `cd frontend && npm run build` - Type-check and bundle
- `cd frontend && npm run lint` - ESLint

## Architecture

- **Analysis**: `POST /api/topology` sends nodes/edges to backend.
- **Validation Rules**: Implemented in `api/handler/check_*.go`. Over 38 rules covering Availability, Performance, Security, etc.
- **Frontend State**: Managed in `Canvas.tsx` (nodes/edges) with history for Undo. Multi-tab support via `useCanvasTabs`.
- **Export**: Utilities in `src/utils/` for Excalidraw, Image, Mermaid, and PDF.
- **Presets**: Standard Demo, Twitter, YouTube, and Google architectures available in the toolbar.
- **Duplicate & Merge**: Support for cloning nodes and merging multiple components into a single role-based node.

## Adding Features

- **New Component**: Update `api/model/topology.go`, `api/model/properties.go`, `frontend/src/types/topology.ts`, and `frontend/src/nodes/nodeConfig.ts`.
- **New Rule**: Add `check_newrule.go` in `api/handler/`, register in `topology_handler.go`, and add a test.
