# Vibe-C4: Interactive C4 Diagram Generator for Go Projects

## Vision

A web service that opens Go code projects and generates interactive C4 architecture
diagrams with zoom-in/zoom-out navigation across all C4 levels (Context, Container,
Component, Code).

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                      │
│  vibe-c4-frontend                                        │
│  ┌──────────────┐ ┌──────────────┐ ┌─────────────────┐  │
│  │ Project       │ │ C4 Diagram   │ │ Detail Panel    │  │
│  │ Upload/Select │ │ Viewer       │ │ (contextual)    │  │
│  │              │ │ (React Flow) │ │                 │  │
│  └──────────────┘ └──────────────┘ └─────────────────┘  │
│  Tech: React 19 + TypeScript + React Flow + Tailwind     │
└──────────────────┬──────────────────────────────────────┘
                   │ REST API (JSON)
┌──────────────────▼──────────────────────────────────────┐
│                    Backend (Go)                          │
│  vibe-c4-api                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌─────────────────┐  │
│  │ HTTP API      │ │ Go Analyzer  │ │ C4 Model        │  │
│  │ (chi router)  │ │ (go/ast +    │ │ Builder         │  │
│  │              │ │  go/parser)  │ │                 │  │
│  └──────────────┘ └──────────────┘ └─────────────────┘  │
│  Tech: Go 1.22+ / chi / go/ast / go/parser               │
└─────────────────────────────────────────────────────────┘
```

## C4 Model Mapping for Go Projects

### Level 1 - System Context
- **System**: The Go module itself (from go.mod)
- **External Systems**: External dependencies (third-party modules)
- **Actors**: Identified from main packages and entry points

### Level 2 - Container
- **Containers**: Major packages/directories (cmd/, internal/, pkg/, api/)
- **Relationships**: Package import dependencies

### Level 3 - Component
- **Components**: Interfaces, structs, and key functions within a package
- **Relationships**: Method calls, interface implementations, struct compositions

### Level 4 - Code (optional, auto-generated)
- **Code elements**: Individual functions, methods, fields
- **Relationships**: Call graph within a component

## Component Breakdown

### Backend (Go Service)

#### 1. HTTP API Layer (`internal/api/`)
- `POST /api/v1/projects/analyze` — Upload/point to a Go project, trigger analysis
- `GET  /api/v1/projects/{id}/diagram` — Get full C4 model for a project
- `GET  /api/v1/projects/{id}/diagram/context` — Level 1: System Context
- `GET  /api/v1/projects/{id}/diagram/containers` — Level 2: Containers
- `GET  /api/v1/projects/{id}/diagram/containers/{name}/components` — Level 3: Components
- `GET  /api/v1/projects/{id}/diagram/components/{name}/code` — Level 4: Code
- `GET  /api/v1/health` — Health check
- CORS middleware for frontend

#### 2. Go Code Analyzer (`internal/analyzer/`)
- Parse Go source files using `go/ast` and `go/parser`
- Extract module info from `go.mod`
- Build package dependency graph
- Extract structs, interfaces, functions per package
- Detect interface implementations
- Track struct compositions and embeddings

#### 3. C4 Model Builder (`internal/c4model/`)
- Domain types: System, Container, Component, CodeElement, Relationship
- Transform analyzer output into C4 model hierarchy
- Compute relationships at each level
- Generate node positions (layout algorithm)

#### 4. Project Manager (`internal/project/`)
- Handle project uploads (tar.gz / zip)
- Handle git clone from URL
- Manage project storage and lifecycle

### Frontend (React Application)

#### 1. Project Management (`features/project/`)
- Upload project archive or provide git URL
- Project list and selection
- Analysis status tracking

#### 2. C4 Diagram Viewer (`features/diagram/`)
- Interactive diagram using React Flow (XyFlow)
- Custom nodes for each C4 element type (System, Container, Component)
- Custom edges with labels for relationships
- Zoom levels mapped to C4 levels:
  - Zoomed out → System Context (Level 1)
  - Zoom in → Container view (Level 2)
  - Zoom deeper → Component view (Level 3)
  - Maximum zoom → Code view (Level 4)
- Click-to-drill-down navigation
- Breadcrumb navigation for current depth
- MiniMap for orientation
- Fit-to-view controls

#### 3. Detail Panel (`features/detail/`)
- Side panel showing details of selected element
- Source code preview for code-level elements
- Dependency list
- File path references

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **Router**: chi (lightweight, idiomatic)
- **AST Parsing**: go/ast, go/parser, go/token (stdlib)
- **Module Info**: golang.org/x/mod for go.mod parsing
- **Testing**: Go stdlib testing + testify
- **Container**: Docker

### Frontend
- **Framework**: React 19 + TypeScript (via vibe-scaffolding)
- **Build**: Vite 6
- **Diagram**: React Flow (@xyflow/react)
- **Styling**: Tailwind CSS 4
- **State**: TanStack Query 5 (server) + Zustand 5 (client)
- **HTTP**: Fetch API via api-client
- **Testing**: Vitest + Testing Library + Playwright
- **Container**: Docker + nginx

## Development Phases

### Phase 1: Scaffolding & Foundation (Tasks 1-3)
1. Generate frontend from vibe-scaffolding
2. Initialize Go backend with project structure
3. Define shared API types/contracts

### Phase 2: Backend Core (Tasks 4-7)
4. Go code analyzer — module and package extraction
5. Go code analyzer — struct/interface/function extraction
6. C4 model builder — transform analyzer output
7. REST API endpoints

### Phase 3: Frontend Core (Tasks 8-11)
8. Project upload/selection UI
9. C4 diagram viewer with React Flow
10. Custom C4 nodes and edges
11. Zoom-level navigation (drill-down)

### Phase 4: Integration & Polish (Tasks 12-14)
12. End-to-end integration testing
13. Detail panel and breadcrumb navigation
14. Docker compose for full stack

### Phase 5: Quality & Documentation (Task 15)
15. E2E tests, documentation, code review

## Team Structure

| Role | Name | Responsibilities |
|------|------|-----------------|
| **Lead** | team-lead | Coordination, architecture, code review |
| **Backend Dev** | backend-dev | Go service: analyzer, C4 builder, API |
| **Frontend Dev** | frontend-dev | React app: diagram viewer, navigation |
| **Integration/QA** | qa-engineer | E2E tests, integration, Docker setup |

## Parallelization Strategy

```
Phase 1: Sequential (foundation must be laid first)
  [scaffolding] → [go-init] → [api-contract]

Phase 2 & 3: Parallel (backend + frontend develop against contract)
  [backend-analyzer] ──→ [backend-c4-builder] ──→ [backend-api]
  [frontend-upload]  ──→ [frontend-diagram]   ──→ [frontend-zoom]

Phase 4: Sequential (integration requires both sides)
  [integration] → [detail-panel] → [docker-compose]

Phase 5: Sequential (quality gate)
  [e2e-tests] → [review] → [ship]
```
