# AGENTS.md - System Design Visualizer

## Project Overview

A full-stack system design visualizer with a React frontend (using React Flow) and a Go Gin backend. The app allows users to visually design system architectures and analyze them for best practices.

## Directory Structure

```
├── frontend/          # React + TypeScript + Vite
│   └── src/
│       ├── api/       # API client functions
│       ├── components/  # React components (Canvas, Sidebar, TabBar)
│       ├── nodes/     # Custom React Flow node types
│       └── utils/     # Export utilities (Excalidraw, Mermaid, Image)
├── api/               # Go + Gin
│   ├── _cmd/          # Server entry (main.go)
│   ├── handler/       # HTTP handlers (validation rules)
│   └── model/         # Data models
```

---

## Build, Lint & Development Commands

| Command | Description |
|---------|-------------|
| `./start-dev.sh` | Start both frontend and backend (Ctrl+C to stop) |

### Frontend (React + TypeScript + Vite)

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server (localhost:5173) |
| `npm run build` | Build for production (typecheck + Vite build) |
| `npm run lint` | Run ESLint on all files |

### Backend (Go + Gin)

| Command | Description |
|---------|-------------|
| `cd api && go run _cmd/main.go` | Run backend server (localhost:8080) |
| `cd api && go test ./...` | Run all tests |
| `cd api && go fmt` | Format Go code |

---

## Code Style Guidelines

- **Keep files small and focused** - Single responsibility per file
- **Functional Components** - React components should be functions
- **Explicit TypeScript types** - Avoid `any`, use interfaces/types
- **Error Handling** - Handle Go errors explicitly; wrap async calls in React with try/catch

---

## Working with this Codebase

### Adding a New Validation Rule

1. Add a new `check_*.go` file in `api/handler/`
2. Implement the rule using `model.TopologyContext`
3. Call the rule function in `handler.validate()` within `api/handler/topology_handler.go`
4. Add corresponding tests in `api/handler/check_*_test.go`

### Adding a New Export Format

1. Create an export utility in `frontend/src/utils/`
2. Register the format in `frontend/src/components/ExportMenu.tsx`
