package c4model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// BuildContextPrompt returns the system prompt for Pass 1 (System Context + Containers, Level 1 + Level 2).
func BuildContextPrompt() string {
	return `You are a software architecture expert. Analyze the provided Go project and produce a C4 System Context and Container diagram (Level 1 + Level 2).

Output valid JSON with this structure:
{
  "systems": [...],
  "containers": [...],
  "relationships": [...]
}

Each system must have: id, name, description, external (bool), module_path.
Each container must have: id, name, description, technology, package_path, system_id.
Each relationship must have: source_id, target_id, description, technology, level.`
}

// BuildContextUserPrompt builds the user prompt for Pass 1 with analysis data and repository contents.
func BuildContextUserPrompt(result *analyzer.AnalysisResult, repoContents map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Module: %s\n\n", result.Module.ModulePath))
	sb.WriteString("Packages:\n")
	for _, pkg := range result.Packages {
		sb.WriteString(fmt.Sprintf("- %s (role: %s)\n", pkg.ImportPath, pkg.Role))
	}
	sb.WriteString("\nKey files:\n")
	for path, content := range repoContents {
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", path, content))
	}
	return sb.String()
}

// BuildComponentPrompt returns the system prompt for Pass 2 (Component diagram, Level 3).
func BuildComponentPrompt(container Container, siblingContainers []Container) string {
	return fmt.Sprintf(`You are a software architecture expert. Analyze the provided container "%s" and produce a C4 Component diagram (Level 3).

Output valid JSON with this structure:
{
  "components": [...],
  "relationships": [...]
}

Each component must have: id, name, description, technology, type, container_id, package_path.
Each relationship must have: source_id, target_id, description, technology, level.`, container.Name)
}

// BuildComponentUserPrompt builds the user prompt for Pass 2 with scoped analysis and files.
func BuildComponentUserPrompt(scopedResult *analyzer.AnalysisResult, container Container, scopedFiles map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Container: %s (package: %s)\n\n", container.Name, container.PackagePath))
	sb.WriteString("Scoped files:\n")
	for path, content := range scopedFiles {
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", path, content))
	}
	return sb.String()
}

// BuildCodePrompt returns the system prompt for Pass 3 (Code diagram, Level 4).
func BuildCodePrompt(component Component) string {
	return fmt.Sprintf(`You are a software architecture expert. Analyze the provided component "%s" and produce a C4 Code diagram (Level 4).

Output valid JSON with this structure:
{
  "code_elements": [...]
}

Each code element must have: id, name, description, type, component_id, file_path.`, component.Name)
}

// BuildCodeUserPrompt builds the user prompt for Pass 3 with scoped analysis and files.
func BuildCodeUserPrompt(scopedResult *analyzer.AnalysisResult, component Component, scopedFiles map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Component: %s (package: %s)\n\n", component.Name, component.PackagePath))
	sb.WriteString("Scoped files:\n")
	for path, content := range scopedFiles {
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", path, content))
	}
	return sb.String()
}

// ReadFullRepository reads repository files up to a token budget.
// It uses a tier-based approach, prioritizing Go source files.
func ReadFullRepository(projectPath string, tokenBudget int) (map[string]string, error) {
	result := make(map[string]string)
	usedTokens := 0
	avgCharsPerToken := 4

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "go.mod") {
			return nil
		}
		relPath, _ := filepath.Rel(projectPath, path)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		estimatedTokens := len(content) / avgCharsPerToken
		if usedTokens+estimatedTokens > tokenBudget {
			return nil
		}
		usedTokens += estimatedTokens
		result[relPath] = string(content)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read repository: %w", err)
	}
	return result, nil
}

// ScopeFilesToPackage filters repository contents to files matching a package path.
func ScopeFilesToPackage(repoContents map[string]string, packagePath, modulePath string) map[string]string {
	result := make(map[string]string)
	// Derive the relative directory from the package path
	relDir := strings.TrimPrefix(packagePath, modulePath)
	relDir = strings.TrimPrefix(relDir, "/")

	for path, content := range repoContents {
		if strings.HasPrefix(path, relDir) {
			result[path] = content
		}
	}
	return result
}

// ScopeAnalysis filters an AnalysisResult to only include packages matching the given paths.
func ScopeAnalysis(result *analyzer.AnalysisResult, packagePaths []string) *analyzer.AnalysisResult {
	pathSet := make(map[string]bool, len(packagePaths))
	for _, p := range packagePaths {
		pathSet[p] = true
	}

	scoped := &analyzer.AnalysisResult{
		Module: result.Module,
	}
	for _, pkg := range result.Packages {
		if pathSet[pkg.ImportPath] {
			scoped.Packages = append(scoped.Packages, pkg)
		}
	}
	return scoped
}
