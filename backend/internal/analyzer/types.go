package analyzer

import "context"

// Analyzer defines the interface for Go project analysis.
type Analyzer interface {
	AnalyzeProject(ctx context.Context, projectPath string) (*AnalysisResult, error)
}

// ExternalDependencyKind categorizes the type of external system interaction.
type ExternalDependencyKind string

const (
	ExtKindHTTPHandler    ExternalDependencyKind = "http_handler"
	ExtKindHTTPClient     ExternalDependencyKind = "http_client"
	ExtKindKafkaProducer  ExternalDependencyKind = "kafka_producer"
	ExtKindKafkaConsumer  ExternalDependencyKind = "kafka_consumer"
	ExtKindDatabase       ExternalDependencyKind = "database"
	ExtKindGRPCServer     ExternalDependencyKind = "grpc_server"
	ExtKindGRPCClient     ExternalDependencyKind = "grpc_client"
)

// FunctionCall represents a detected call from one function to another.
type FunctionCall struct {
	CallerPkg      string `json:"caller_pkg"`
	CallerType     string `json:"caller_type,omitempty"`
	CallerFunc     string `json:"caller_func"`
	CalleePkg      string `json:"callee_pkg"`
	CalleeType     string `json:"callee_type,omitempty"`
	CalleeFunc     string `json:"callee_func"`
	CalleePkgAlias string `json:"callee_pkg_alias,omitempty"`
	FilePath       string `json:"file_path,omitempty"`
	Line           int    `json:"line,omitempty"`
}

// ExternalInteraction represents a detected interaction with an external system.
type ExternalInteraction struct {
	Kind       ExternalDependencyKind `json:"kind"`
	PkgPath    string                 `json:"pkg_path"`
	TypeName   string                 `json:"type_name,omitempty"`
	FuncName   string                 `json:"func_name"`
	Detail     string                 `json:"detail,omitempty"`
	Technology string                 `json:"technology,omitempty"`
	FilePath   string                 `json:"file_path,omitempty"`
	Line       int                    `json:"line,omitempty"`
}

// InterfaceImpl represents a detected struct-implements-interface relationship.
type InterfaceImpl struct {
	StructPkg     string `json:"struct_pkg"`
	StructName    string `json:"struct_name"`
	InterfacePkg  string `json:"interface_pkg"`
	InterfaceName string `json:"interface_name"`
}

// Entrypoint represents a detected application entry point.
type Entrypoint struct {
	Kind     string `json:"kind"`
	PkgPath  string `json:"pkg_path"`
	FuncName string `json:"func_name"`
	Route    string `json:"route,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// AnalysisResult holds the full analysis output for a Go project.
type AnalysisResult struct {
	ProjectPath          string                 `json:"project_path"`
	Module               ModuleInfo             `json:"module"`
	Packages             []PackageInfo          `json:"packages"`
	ImportGraph          map[string][]string    `json:"import_graph"`
	CallGraph            []FunctionCall         `json:"call_graph,omitempty"`
	ExternalInteractions []ExternalInteraction  `json:"external_interactions,omitempty"`
	InterfaceImpls       []InterfaceImpl        `json:"interface_impls,omitempty"`
	Entrypoints          []Entrypoint           `json:"entrypoints,omitempty"`
}

// ModuleInfo holds information extracted from go.mod.
type ModuleInfo struct {
	ModulePath   string       `json:"module_path"`
	GoVersion    string       `json:"go_version"`
	DirectDeps   []Dependency `json:"direct_deps"`
	IndirectDeps []Dependency `json:"indirect_deps"`
}

// Dependency represents an external module dependency.
type Dependency struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// PackageInfo holds information about a single Go package in the project.
type PackageInfo struct {
	Name       string          `json:"name"`
	ImportPath string          `json:"import_path"`
	Dir        string          `json:"dir"`
	Role       string          `json:"role"`
	GoFiles    []string        `json:"go_files"`
	Imports    []string        `json:"imports"`
	Structs    []StructInfo    `json:"structs,omitempty"`
	Interfaces []InterfaceInfo `json:"interfaces,omitempty"`
	Functions  []FunctionInfo  `json:"functions,omitempty"`
}

// FieldInfo represents a struct field.
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Tag      string `json:"tag,omitempty"`
	Embedded bool   `json:"embedded,omitempty"`
}

// MethodInfo represents a method signature.
type MethodInfo struct {
	Name       string   `json:"name"`
	Params     []string `json:"params,omitempty"`
	Returns    []string `json:"returns,omitempty"`
	FilePath   string   `json:"file_path,omitempty"`
	Line       int      `json:"line,omitempty"`
}

// StructInfo holds metadata about a Go struct.
type StructInfo struct {
	Name     string      `json:"name"`
	Fields   []FieldInfo `json:"fields,omitempty"`
	Methods  []MethodInfo `json:"methods,omitempty"`
	FilePath string      `json:"file_path,omitempty"`
	Line     int         `json:"line,omitempty"`
}

// InterfaceInfo holds metadata about a Go interface.
type InterfaceInfo struct {
	Name     string       `json:"name"`
	Methods  []MethodInfo `json:"methods,omitempty"`
	FilePath string       `json:"file_path,omitempty"`
	Line     int          `json:"line,omitempty"`
}

// FunctionInfo holds metadata about a standalone function.
type FunctionInfo struct {
	Name     string   `json:"name"`
	Params   []string `json:"params,omitempty"`
	Returns  []string `json:"returns,omitempty"`
	FilePath string   `json:"file_path,omitempty"`
	Line     int      `json:"line,omitempty"`
}
