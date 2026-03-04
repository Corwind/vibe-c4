package api

import "github.com/Corwind/vibe-c4/backend/internal/c4model"

// Position represents a 2D coordinate for React Flow nodes.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData holds the payload for a React Flow node.
type NodeData struct {
	Label           string `json:"label"`
	Description     string `json:"description,omitempty"`
	Technology      string `json:"technology,omitempty"`
	C4Type          string `json:"c4Type"`
	External        bool   `json:"external,omitempty"`
	IsEntrypoint    bool   `json:"isEntrypoint,omitempty"`
	EntrypointKind  string `json:"entrypointKind,omitempty"`
	EntrypointRoute string `json:"entrypointRoute,omitempty"`
	SystemKind      string `json:"systemKind,omitempty"`
}

// Node represents a React Flow node.
type Node struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

// Edge represents a React Flow edge.
type Edge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Label    string `json:"label,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

// DiagramResponse is the common response for all diagram level endpoints.
type DiagramResponse struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// AnalyzeRequest is the JSON body for project analysis.
type AnalyzeRequest struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	GitURL string `json:"gitUrl"`
}

// AnalyzeResponse is returned after triggering analysis.
type AnalyzeResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ProjectResponse is returned for a single project in list results.
type ProjectResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Status string `json:"status"`
}

// ListProjectsResponse wraps the project list.
type ListProjectsResponse struct {
	Data []ProjectResponse `json:"data"`
}

// buildContextDiagram builds the Level 1 (System Context) diagram.
func buildContextDiagram(model *c4model.C4Model) DiagramResponse {
	var nodes []Node
	var edges []Edge

	for i, sys := range model.Systems {
		nodeType := "system"
		if sys.External {
			nodeType = "externalSystem"
		}
		nodes = append(nodes, Node{
			ID:       sys.ID,
			Type:     nodeType,
			Position: Position{X: float64(i * 300), Y: float64(boolToInt(!sys.External) * 200)},
			Data: NodeData{
				Label:       sys.Name,
				Description: sys.Description,
				C4Type:      "system",
				External:    sys.External,
				SystemKind:  sys.SystemKind,
			},
		})
	}

	for _, rel := range model.Relationships {
		// Only include system-level relationships
		if model.FindSystem(rel.SourceID) != nil && model.FindSystem(rel.TargetID) != nil {
			edges = append(edges, Edge{
				ID:     "edge-" + rel.SourceID + "-" + rel.TargetID,
				Source: rel.SourceID,
				Target: rel.TargetID,
				Label:  rel.Description,
			})
		}
	}

	return DiagramResponse{Nodes: ensureNodes(nodes), Edges: ensureEdges(edges)}
}

// buildContainerDiagram builds the Level 2 (Container) diagram.
func buildContainerDiagram(model *c4model.C4Model) DiagramResponse {
	var nodes []Node
	var edges []Edge

	for i, c := range model.Containers {
		col := i % 4
		row := i / 4
		nodes = append(nodes, Node{
			ID:       c.ID,
			Type:     "container",
			Position: Position{X: float64(col * 300), Y: float64(row * 200)},
			Data: NodeData{
				Label:      c.Name,
				Technology: c.Technology,
				C4Type:     "container",
			},
		})
	}

	for _, rel := range model.Relationships {
		if model.FindContainer(rel.SourceID) != nil && model.FindContainer(rel.TargetID) != nil {
			edges = append(edges, Edge{
				ID:       "edge-" + rel.SourceID + "-" + rel.TargetID,
				Source:   rel.SourceID,
				Target:   rel.TargetID,
				Label:    rel.Description,
				Animated: true,
			})
		}
	}

	return DiagramResponse{Nodes: ensureNodes(nodes), Edges: ensureEdges(edges)}
}

// buildComponentDiagram builds the Level 3 (Component) diagram for a given container.
func buildComponentDiagram(model *c4model.C4Model, containerID string) DiagramResponse {
	var nodes []Node
	var edges []Edge

	components := model.ComponentsForContainer(containerID)
	componentIDs := make(map[string]bool)
	for i, comp := range components {
		componentIDs[comp.ID] = true
		col := i % 3
		row := i / 3
		nodes = append(nodes, Node{
			ID:       comp.ID,
			Type:     "component",
			Position: Position{X: float64(col * 250), Y: float64(row * 180)},
			Data: NodeData{
				Label:           comp.Name,
				Technology:      comp.Technology,
				C4Type:          string(comp.Type),
				IsEntrypoint:    comp.IsEntrypoint,
				EntrypointKind:  comp.EntrypointKind,
				EntrypointRoute: comp.EntrypointRoute,
			},
		})
	}

	// Add edges for component-level relationships
	ghostNodes := make(map[string]bool)
	for _, rel := range model.Relationships {
		if rel.Level != "component" {
			continue
		}
		sourceInContainer := componentIDs[rel.SourceID]
		targetInContainer := componentIDs[rel.TargetID]

		if !sourceInContainer && !targetInContainer {
			continue // neither endpoint is in this container
		}

		// Add ghost nodes for cross-container references
		if !sourceInContainer && !ghostNodes[rel.SourceID] {
			ghostNodes[rel.SourceID] = true
			comp := model.FindComponent(rel.SourceID)
			if comp != nil {
				nodes = append(nodes, Node{
					ID:       comp.ID,
					Type:     "ghostComponent",
					Position: Position{X: float64(len(nodes) % 3 * 250), Y: float64(len(nodes) / 3 * 180)},
					Data: NodeData{
						Label:      comp.Name,
						Technology: comp.Technology,
						C4Type:     string(comp.Type),
						External:   true,
					},
				})
			}
		}
		if !targetInContainer && !ghostNodes[rel.TargetID] {
			ghostNodes[rel.TargetID] = true
			comp := model.FindComponent(rel.TargetID)
			if comp != nil {
				nodes = append(nodes, Node{
					ID:       comp.ID,
					Type:     "ghostComponent",
					Position: Position{X: float64(len(nodes) % 3 * 250), Y: float64(len(nodes) / 3 * 180)},
					Data: NodeData{
						Label:      comp.Name,
						Technology: comp.Technology,
						C4Type:     string(comp.Type),
						External:   true,
					},
				})
			}
		}

		edges = append(edges, Edge{
			ID:       "edge-" + rel.SourceID + "-" + rel.TargetID,
			Source:   rel.SourceID,
			Target:   rel.TargetID,
			Label:    rel.Description,
			Animated: rel.Description == "implements",
		})
	}

	return DiagramResponse{Nodes: ensureNodes(nodes), Edges: ensureEdges(edges)}
}

// buildCodeDiagram builds the Level 4 (Code) diagram for a given component.
func buildCodeDiagram(model *c4model.C4Model, componentID string) DiagramResponse {
	var nodes []Node

	for i, ce := range model.CodeElements {
		if ce.ComponentID != componentID {
			continue
		}
		nodes = append(nodes, Node{
			ID:       ce.ID,
			Type:     "codeElement",
			Position: Position{X: float64((i % 4) * 200), Y: float64((i / 4) * 150)},
			Data: NodeData{
				Label:  ce.Name,
				C4Type: string(ce.Type),
			},
		})
	}

	return DiagramResponse{Nodes: ensureNodes(nodes), Edges: ensureEdges(nil)}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ensureNodes(nodes []Node) []Node {
	if nodes == nil {
		return []Node{}
	}
	return nodes
}

func ensureEdges(edges []Edge) []Edge {
	if edges == nil {
		return []Edge{}
	}
	return edges
}
