# ArchitectMind: System Design Visualizer

[🌐 Live Demo](https://architect-mind.vercel.app/)

![ArchitectMind Demo](./demo.png)

ArchitectMind is a web-based tool for visualizing and analyzing system architectures. It provides an intuitive interactive canvas where users can drag and drop infrastructure components (such as Load Balancers, Databases, Services, etc.) and receive architectural logic validation and best practice recommendations via a backend API.

## 🚀 Features

- **Interactive Canvas**: Based on React Flow v12, supports node drag-and-drop, connections, and custom properties.
- **Component Sidebar**: Provides a standard library of system design components (Client, CDN, API Gateway, Message Queue, etc.).
- **Backend Architecture Validation (Analyze)**: Supports 46+ automated rules to check if component types and connection logic in the diagram align with system design best practices.
- **Custom Analysis Parameters**: Set system parameters like DAU, QPS, Storage, and Availability targets to generate more precise capacity planning recommendations during analysis.
- **Multi-tab Support**: Open multiple design canvases simultaneously for architecture comparison and multi-project workflows.
- **One-click Presets**: Classic system design templates (Standard Demo, Twitter, YouTube, Google) to help you get started quickly.
- **Export Capabilities**: Supports exporting to Excalidraw, Image, Mermaid, and PDF formats.
- **Quick Operations**: Supports **Duplicate**, **Undo**, and **Merge/Split** functions to improve drawing efficiency.
- **Multi-theme Support**: 5 themes available - Light, Dark, Warm, Dream, and CyberPunk to suit different visual preferences.
- **Responsive Design**: Clean and modern user interface.

## 🛠 Tech Stack

### Frontend

- **React 19** + **TypeScript**
- **Vite** (Build tool)
- **React Flow (@xyflow/react v12)** (Canvas engine)

### Backend

- **Go 1.25+**
- **Gin Web Framework** (API routing)
- **MongoDB** (Interface reserved for data persistence)

## 📦 Installation and Setup

### 1. Quick Start

A convenience script is provided in the root directory to start both the backend and frontend development servers:

```bash
chmod +x start-dev.sh
./start-dev.sh
```

After starting:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

## ✅ Tests

Tests are currently focused on the backend. Run them in the root directory or `api` subdirectory:

```bash
go test ./... -v
```

## 🔍 Validation Rule Examples

The backend implements over 38 [Architecture Validation Rules](./docs/RULES.md), covering:

- **Availability**: Single Point of Failure (SPOF) checks, LB/API Gateway redundancy validation, Health Check configuration.
- **Performance**: Read/Write separation suggestions, Cache consistency and eviction policies, CDN global acceleration recommendations.
- **Security**: `invalid_connection` (detecting unreasonable component connection directions, such as DB→Client, LB→Database, DNS→Service, etc.), Firewall/WAF missing checks.
- **Scalability**: Asynchronous decoupling (MQ), traffic peak shaving suggestions, database vertical partitioning reminders.
- **Observability**: Logger/Monitor missing checks, Alerting configuration validation.

## 🗺️ Roadmap

- [ ] **Infrastructure as Code (IaC) Integration**: Connect with tools like Terraform to generate deployment scripts directly from architecture diagrams.
- [ ] **AI Deployment Prompt Export**: Export system designs as specialized AI Prompts to help non-technical users quickly understand and implement deployments using AI tools (e.g., ChatGPT, Claude).

## 📂 Project Structure

- `/frontend`: React source code, React Flow components, custom hooks, and API logic.
- `/api`: Go API handlers, data models, and the validation engine.
