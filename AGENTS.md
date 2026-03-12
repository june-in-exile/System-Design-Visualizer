# AGENTS.md - System Design Visualizer

## Project Overview

A full-stack system design visualizer with a React frontend (using React Flow) and a Go Gin backend. The app allows users to visually design system architectures and analyze them for best practices.

## Directory Structure

```
├── frontend/          # React + TypeScript + Vite
│   └── src/
│       ├── api/       # API client functions
│       ├── components/  # React components (Canvas, Sidebar)
│       ├── nodes/     # Custom React Flow node types
│       └── types/     # TypeScript type definitions
├── backend/           # Go + Gin
│   ├── handler/       # HTTP handlers
│   └── model/         # Data models
```

---

## Build, Lint & Development Commands

| `./start-dev.sh` | Start both frontend and backend (Ctrl+C to stop) |

### Frontend (React + TypeScript + Vite)

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server (localhost:5173) |
| `npm run build` | Build for production (typecheck + Vite build) |
| `npm run lint` | Run ESLint on all files |
| `npm run preview` | Preview production build |

### Backend (Go + Gin)

| Command | Description |
|---------|-------------|
| `go run .` | Run backend server (localhost:8080) |
| `go build` | Build binary |
| `go vet` | Run Go static analysis |
| `go fmt` | Format Go code |

**Note:** This project currently has no test framework configured.

---

## Code Style Guidelines

### General Principles

- **Keep files small and focused** - Single responsibility per file
- **Use functional components** - React components should be functions, not classes
- **Avoid magic numbers** - Use named constants for any non-obvious values

---

### Frontend (TypeScript + React)

#### Imports

- **Group imports** in this order:
  1. External libraries (React, React Flow, etc.)
  2. CSS/style imports
  3. Internal components/utils
  4. Type imports

```typescript
import { useCallback, useState } from 'react'
import { ReactFlow, Controls, useNodesState } from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import Canvas from './components/Canvas'
import type { SystemTopology, AnalyzeResponse } from '../types/topology'
```

#### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Components | PascalCase | `Canvas.tsx`, `ArchitectureNode.tsx` |
| Functions | camelCase | `analyzeTopology`, `generateNodeId` |
| Types/Interfaces | PascalCase | `SystemTopology`, `ComponentType` |
| Constants | PascalCase or SCREAMING_SNAKE | `NODE_TYPE_CONFIG`, `API_BASE` |
| Files (utils) | camelCase | `topologyApi.ts`, `nodeConfig.ts` |

#### TypeScript

- **Always define explicit types** for function parameters and return values
- **Use type inference** only for obvious cases (e.g., `const count = 0`)
- **Use `type`** for unions, intersections, and type aliases
- **Use `interface`** for object shapes that may be extended

```typescript
// Good: explicit return type
function generateNodeId(): string {
  return `node-${++nodeIdCounter}`
}

// Good: interface for extensible objects
interface DatabaseProperties {
  dbType: 'sql' | 'nosql' | 'graph' | 'timeseries'
  acidRequired: boolean
}
```

#### React Patterns

- **Use hooks for state** - `useState`, `useCallback`, `useRef`, `useEffect`
- **Memoize callbacks** - Wrap event handlers with `useCallback` when passed as props
- **Destructure props** - Extract needed values from props objects

```typescript
// Good: destructure props and memoize
function Canvas() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])

  const onConnect: OnConnect = useCallback(
    (params) => { /* ... */ },
    [setEdges]
  )
}
```

#### Error Handling

- **Always wrap async calls** in try/catch
- **Provide meaningful error messages** - Include context in error strings
- **Handle API errors gracefully** - Show user-friendly messages

```typescript
try {
  const result = await analyzeTopology(topology)
  setAnalysisResult(result)
} catch (error) {
  setAnalysisResult({
    success: false,
    nodeCount: 0,
    edgeCount: 0,
    warnings: [
      error instanceof Error ? error.message : 'Analysis failed',
    ],
  })
}
```

#### Styling

- **Use inline styles** for simple, dynamic styling
- **Use CSS files** (`index.css`) for global styles
- **Keep styles co-located** with components when complex

---

### Backend (Go)

#### Naming

- **Packages**: lowercase, short, no underscores (e.g., `handler`, `model`)
- **Files**: lowercase with underscores (e.g., `topology_handler.go`)
- **Functions/Variables**: camelCase
- **Types/Interfaces**: PascalCase
- **Constants**: PascalCase or SCREAMING_SNAKE for exported

#### Error Handling

- **Handle errors explicitly** - Check error values, don't ignore them
- **Return meaningful errors** - Use `fmt.Errorf` with context
- **Log failures** - Use `log` or structured logging

```go
if err := r.Run(":8080"); err != nil {
    log.Fatalf("failed to start server: %v", err)
}
```

#### Imports

- **Group imports**:
  1. Standard library
  2. External packages (grouped by project)

```go
import (
    "log"

    "github.com/architectmind/backend/handler"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)
```

---

## Working with this Codebase

### Adding a New Component Type

1. Add the type to `frontend/src/types/topology.ts` (e.g., `ComponentType` union)
2. Add properties interface if needed in the same file
3. Add config in `frontend/src/nodes/nodeConfig.ts`
4. Create custom node component in `frontend/src/nodes/`
5. Register in `Canvas.tsx` nodeTypes

### Adding a New API Endpoint

1. Add handler function in `backend/handler/`
2. Register route in `backend/main.go`
3. Add API client function in `frontend/src/api/`
4. Add frontend types if needed

### Adding Tests

Currently there are no tests. To add tests:

- Frontend: Use Vitest (`npm create vitest@latest`)
- Backend: Use Go's built-in testing package

---

## Development Workflow

1. **Start both servers**: `./start-dev.sh` (Ctrl+C to stop)
2. **Lint before committing**: `cd frontend && npm run lint`
3. **Build to verify**: `cd frontend && npm run build`
