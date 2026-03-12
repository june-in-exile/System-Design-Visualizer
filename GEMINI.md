# ArchitectMind: System Design Visualizer

ArchitectMind is a web-based tool for visualizing and analyzing system architectures. It provides an interactive canvas to build infrastructure diagrams and a backend that validates the topology's semantic correctness.

## Project Structure

- `backend/`: Go-based API using the Gin framework.
  - `main.go`: Entry point, CORS configuration, and route definitions.
  - `handler/`: API request handlers (e.g., `PostTopology` for diagram analysis).
  - `model/`: Data structures for system nodes, edges, and topologies.
- `frontend/`: React-based visualizer using Vite and React Flow.
  - `src/components/Canvas.tsx`: Main diagramming area.
  - `src/components/Sidebar.tsx`: Palette of available system components.
  - `src/nodes/`: Custom React Flow node implementations.
  - `src/api/`: Frontend clients for backend communication.

## Tech Stack

- **Backend:** Go 1.25+, [Gin](https://gin-gonic.com/), MongoDB (driver ready).
- **Frontend:** React 19, [Vite](https://vitejs.dev/), [React Flow (@xyflow/react)](https://reactflow.dev/), TypeScript.

## Getting Started

### Backend

1. Navigate to the `backend` directory.
2. Run the server:

   ```bash
   go run main.go
   ```

   The backend will start on `http://localhost:8080`.

### Frontend

1. Navigate to the `frontend` directory.
2. Install dependencies:

   ```bash
   npm install
   ```

3. Start the development server:

   ```bash
   npm run dev
   ```

   The frontend will be available at `http://localhost:5173`.

## Development Conventions

- **Backend:**
  - Follow idiomatic Go practices.
  - Define new architecture components in `model/topology.go` (both struct and `ValidComponentTypes`).
  - Validation logic resides in `handler/topology_handler.go`.
- **Frontend:**
  - Use TypeScript for all components and API interactions.
  - Custom node styling should be managed in `src/nodes/ArchitectureNode.tsx` or via CSS modules.
  - Keep types in sync between `frontend/src/types/` and `backend/model/`.
