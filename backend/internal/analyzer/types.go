package analyzer

import "context"

// Analyzer defines the interface for Go project analysis.
type Analyzer interface {
	AnalyzeProject(ctx context.Context, projectPath string) (*AnalysisResult, error)
}

// AnalysisResult holds the full analysis output for a Go project.
type AnalysisResult struct {
	Module      ModuleInfo        `json:"module"`
	Packages    []PackageInfo     `json:"packages"`
	ImportGraph map[string][]string `json:"import_graph"`
}

// ModuleInfo holds information extracted from go.mod.
type ModuleInfo struct {
	ModulePath       string       `json:"module_path"`
	GoVersion        string       `json:"go_version"`
	DirectDeps       []Dependency `json:"direct_deps"`
	IndirectDeps     []Dependency `json:"indirect_deps"`
}

// Dependency represents an external module dependency.
type Dependency struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// PackageInfo holds information about a single Go package in the project.
type PackageInfo struct {
	Name       string   `json:"name"`
	ImportPath string   `json:"import_path"`
	Dir        string   `json:"dir"`
	Role       string   `json:"role"`
	GoFiles    []string `json:"go_files"`
	Imports    []string `json:"imports"`
}
