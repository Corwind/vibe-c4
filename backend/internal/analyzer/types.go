package analyzer

import "context"

// Analyzer defines the interface for Go project analysis.
type Analyzer interface {
	AnalyzeProject(ctx context.Context, projectPath string) (*AnalysisResult, error)
}

// AnalysisResult holds the full analysis output for a Go project.
type AnalysisResult struct {
	Module      ModuleInfo          `json:"module"`
	Packages    []PackageInfo       `json:"packages"`
	ImportGraph map[string][]string `json:"import_graph"`
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
