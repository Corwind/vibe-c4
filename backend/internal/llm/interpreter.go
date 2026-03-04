package llm

import "context"

// Interpreter is the port interface for AI-assisted C4 interpretation.
// Implementations transform compact analysis facts into proper C4 architecture structures.
type Interpreter interface {
	Interpret(ctx context.Context, input *InterpretationInput) (*InterpretationResult, error)
}

// InterpretationInput is the compact representation of analysis facts sent to the LLM.
type InterpretationInput struct {
	ProjectName          string                       `json:"project_name"`
	ModulePath           string                       `json:"module_path"`
	PackageSummaries     []PackageSummary             `json:"package_summaries"`
	ExternalInteractions []ExternalInteractionSummary `json:"external_interactions"`
	Entrypoints          []EntrypointSummary          `json:"entrypoints"`
	CallGraphEdges       []CallEdgeSummary            `json:"call_graph_edges"`
	InterfaceImpls       []InterfaceImplSummary       `json:"interface_impls"`
	DirectDeps           []string                     `json:"direct_deps"`
	DeploymentHints      []string                     `json:"deployment_hints"`
}

// PackageSummary is a compressed representation of a Go package.
type PackageSummary struct {
	Name           string   `json:"name"`
	ImportPath     string   `json:"import_path"`
	Structs        []string `json:"structs,omitempty"`
	Interfaces     []string `json:"interfaces,omitempty"`
	Functions      []string `json:"functions,omitempty"`
	InternalImports []string `json:"internal_imports,omitempty"`
	ExternalImports []string `json:"external_imports,omitempty"`
}

// ExternalInteractionSummary is a deduplicated external interaction fact.
type ExternalInteractionSummary struct {
	Kind       string `json:"kind"`
	Technology string `json:"technology,omitempty"`
}

// EntrypointSummary is a compact entrypoint fact.
type EntrypointSummary struct {
	Kind    string `json:"kind"`
	Package string `json:"package"`
	Route   string `json:"route,omitempty"`
}

// CallEdgeSummary is a deduplicated cross-type call edge.
type CallEdgeSummary struct {
	CallerPkg  string `json:"caller_pkg"`
	CallerType string `json:"caller_type"`
	CalleePkg  string `json:"callee_pkg"`
	CalleeType string `json:"callee_type"`
}

// InterfaceImplSummary is a compact interface implementation fact.
type InterfaceImplSummary struct {
	StructPkg     string `json:"struct_pkg"`
	StructName    string `json:"struct_name"`
	InterfacePkg  string `json:"interface_pkg"`
	InterfaceName string `json:"interface_name"`
}

// InterpretationResult is the structured output from the LLM containing proper C4 architecture.
type InterpretationResult struct {
	SystemContext SystemContextInterpretation `json:"system_context"`
	Containers    []ContainerInterpretation   `json:"containers"`
	Components    []ComponentInterpretation   `json:"components"`
	Relationships []RelationshipInterpretation `json:"relationships"`
}

// SystemContextInterpretation is the Level 1 interpretation.
type SystemContextInterpretation struct {
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	Actors          []ActorInterpretation        `json:"actors"`
	ExternalSystems []ExternalSystemInterpretation `json:"external_systems"`
}

// ActorInterpretation represents a person or external actor that uses the system.
type ActorInterpretation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // "person", "external_system"
}

// ExternalSystemInterpretation represents a true external system (not a library).
type ExternalSystemInterpretation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Technology  string `json:"technology,omitempty"`
	Kind        string `json:"kind"` // "database", "message_queue", "http_api", "grpc_service"
}

// ContainerInterpretation represents a deployable unit.
type ContainerInterpretation struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Technology   string   `json:"technology"`
	Type         string   `json:"type"` // "service", "database", "message_queue", "cache"
	PackagePaths []string `json:"package_paths,omitempty"`
}

// ComponentInterpretation represents an architectural building block.
type ComponentInterpretation struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Role        string   `json:"role"` // "controller", "service", "repository", "gateway", "domain"
	ContainerID string   `json:"container_id"`
	StructNames []string `json:"struct_names,omitempty"`
}

// RelationshipInterpretation represents a business-oriented relationship.
type RelationshipInterpretation struct {
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
	Description string `json:"description"`
	Level       string `json:"level"` // "system", "container", "component"
}
