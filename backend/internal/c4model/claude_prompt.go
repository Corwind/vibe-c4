package c4model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// BuildContextPrompt returns the system prompt for Level 1 (System Context) and Level 2 (Container) generation.
func BuildContextPrompt() string {
	return `You are a software architecture expert specializing in the C4 model.
Your task is to generate ONLY the System Context (Level 1) and Container (Level 2) layers of a C4 model.

## C4 Levels

### Level 1 - System Context
Identifies the software system being modeled, the people (actors) who use it, and the other software systems it interacts with.

### Level 2 - Container
Zooms into the main software system and identifies the deployable/runnable units (containers): applications, services, databases, etc. Each container represents a separately deployable unit or a major package grouping.

## JSON Output Schema

Return a JSON object with EXACTLY these top-level keys:

{
  "systems": [
    {
      "id": "system-<sanitized-name>",
      "name": "<human-readable name>",
      "description": "<what this system does>",
      "external": <true|false>,
      "module_path": "<Go module path if applicable>",
      "container_ids": ["container-..."]
    }
  ],
  "containers": [
    {
      "id": "container-<sanitized-name>",
      "name": "<human-readable name>",
      "description": "<what this container does>",
      "technology": "<e.g., Go, PostgreSQL>",
      "package_path": "<Go import path>",
      "system_id": "system-<parent-system>"
    }
  ],
  "relationships": [
    {
      "source_id": "<id of source element>",
      "target_id": "<id of target element>",
      "description": "<what this relationship represents>",
      "technology": "<optional protocol or technology>",
      "level": "<context|container>"
    }
  ]
}

## ID Rules
- System IDs: prefix "system-", sanitized name (replace "/" and "." with "-")
- Container IDs: prefix "container-", sanitized name (replace "/" and "." with "-")
- Relationship level must be "context" or "container" only

## Quality Guidelines
- Descriptions should be concise but informative (1-2 sentences)
- Every container MUST include a package_path (the Go import path)
- Identify the main system, external systems, and actors
- Group related packages into logical containers

Return ONLY valid JSON. Do NOT wrap in markdown code fences.`
}

// BuildContextUserPrompt returns the user prompt for Level 1 and Level 2 generation.
func BuildContextUserPrompt(result *analyzer.AnalysisResult, repoContents map[string]string) string {
	var sb strings.Builder

	sb.WriteString("## Static Analysis Results\n\n")
	sb.WriteString(formatAnalysisJSON(result))

	sb.WriteString("\n\n## Source Code\n\n")
	for path, content := range repoContents {
		sb.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content))
	}

	sb.WriteString(`## Instructions

Generate ONLY the System Context (Level 1) and Container (Level 2) layers.
Do NOT generate components or code elements.
Identify the main system, external systems, actors, and the deployable containers within the main system.
`)

	return sb.String()
}

// BuildComponentPrompt returns the system prompt for Level 3 (Component) extraction for a single container.
func BuildComponentPrompt(container Container, siblingContainers []Container) string {
	var sb strings.Builder

	sb.WriteString(`You are a software architecture expert specializing in the C4 model.
Your task is to generate the Component (Level 3) layer for a specific container.

## Target Container
`)
	sb.WriteString(fmt.Sprintf("- Name: %s\n", container.Name))
	sb.WriteString(fmt.Sprintf("- ID: %s\n", container.ID))
	sb.WriteString(fmt.Sprintf("- Package Path: %s\n", container.PackagePath))
	sb.WriteString(fmt.Sprintf("- Technology: %s\n", container.Technology))

	if len(siblingContainers) > 0 {
		sb.WriteString("\n## Sibling Containers (for cross-container relationships)\n")
		for _, s := range siblingContainers {
			sb.WriteString(fmt.Sprintf("- %s (ID: %s)\n", s.Name, s.ID))
		}
	}

	sb.WriteString(`
## JSON Output Schema

Return a JSON object with EXACTLY these top-level keys:

{
  "components": [
    {
      "id": "component-<sanitized-name>",
      "name": "<human-readable name>",
      "description": "<what this component does>",
      "technology": "<e.g., Go struct, Go interface>",
      "type": "<struct|interface|function>",
      "container_id": "` + container.ID + `",
      "package_path": "<Go import path>",
      "is_entrypoint": <true|false>,
      "role": "<optional role>"
    }
  ],
  "relationships": [
    {
      "source_id": "<component id>",
      "target_id": "<component or container id>",
      "description": "<what this relationship represents>",
      "technology": "<optional>",
      "level": "component"
    }
  ]
}

## Rules
- Every component MUST have container_id set to "` + container.ID + `"
- Every component MUST include a package_path
- Relationship level must be "component" only
- Identify structs, interfaces, and key functions as components

## Quality Guidelines
- Descriptions should be concise but informative (1-2 sentences)
- Use meaningful names that reflect the Go type or function name
- Set type to "struct", "interface", or "function" appropriately

Return ONLY valid JSON. Do NOT wrap in markdown code fences.`)

	return sb.String()
}

// BuildComponentUserPrompt returns the user prompt for Level 3 component generation for a single container.
func BuildComponentUserPrompt(scopedResult *analyzer.AnalysisResult, container Container, scopedFiles map[string]string) string {
	var sb strings.Builder

	sb.WriteString("## Static Analysis Results\n\n")
	sb.WriteString(formatAnalysisJSON(scopedResult))

	sb.WriteString("\n\n## Source Code\n\n")
	for path, content := range scopedFiles {
		sb.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content))
	}

	sb.WriteString(fmt.Sprintf(`## Instructions

Generate components for container %s (ID: %s).
Every component must have container_id set to "%s".
`, container.Name, container.ID, container.ID))

	return sb.String()
}

// BuildCodePrompt returns the system prompt for Level 4 (Code Element) extraction for a single component.
func BuildCodePrompt(component Component) string {
	var sb strings.Builder

	sb.WriteString(`You are a software architecture expert specializing in the C4 model.
Your task is to generate the Code (Level 4) layer for a specific component.

## Target Component
`)
	sb.WriteString(fmt.Sprintf("- Name: %s\n", component.Name))
	sb.WriteString(fmt.Sprintf("- ID: %s\n", component.ID))
	sb.WriteString(fmt.Sprintf("- Type: %s\n", string(component.Type)))
	sb.WriteString(fmt.Sprintf("- Container ID: %s\n", component.ContainerID))

	if len(component.Methods) > 0 {
		sb.WriteString("\n## Known Methods\n")
		for _, m := range component.Methods {
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
	}

	if len(component.Fields) > 0 {
		sb.WriteString("\n## Known Fields\n")
		for _, f := range component.Fields {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	sb.WriteString(`
## JSON Output Schema

Return a JSON object with EXACTLY this top-level key:

{
  "code_elements": [
    {
      "id": "code-<sanitized-name>",
      "name": "<element name>",
      "description": "<what this element does>",
      "type": "<function|method|field>",
      "component_id": "` + component.ID + `",
      "file_path": "<optional file path>",
      "line": <optional line number>
    }
  ]
}

## Rules
- Every code_element MUST have component_id set to "` + component.ID + `"
- Identify methods, fields, and functions as code elements
- Set type to "function", "method", or "field" appropriately

Return ONLY valid JSON. Do NOT wrap in markdown code fences.`)

	return sb.String()
}

// BuildCodeUserPrompt returns the user prompt for Level 4 code element generation for a single component.
func BuildCodeUserPrompt(scopedResult *analyzer.AnalysisResult, component Component, scopedFiles map[string]string) string {
	var sb strings.Builder

	sb.WriteString("## Static Analysis Results\n\n")
	sb.WriteString(formatAnalysisJSON(scopedResult))

	sb.WriteString("\n\n## Source Code\n\n")
	for path, content := range scopedFiles {
		sb.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content))
	}

	sb.WriteString(fmt.Sprintf(`## Instructions

Generate code elements for component %s (ID: %s).
Every code_element must have component_id set to "%s".
`, component.Name, component.ID, component.ID))

	return sb.String()
}

// ScopeFilesToPackage filters repoContents to only files under the given package's directory.
func ScopeFilesToPackage(repoContents map[string]string, packagePath, modulePath string) map[string]string {
	// If package path equals module path, return all files (root package)
	if packagePath == modulePath {
		result := make(map[string]string, len(repoContents))
		for k, v := range repoContents {
			result[k] = v
		}
		return result
	}

	// If package path doesn't start with module path, return empty
	if !strings.HasPrefix(packagePath, modulePath+"/") {
		return map[string]string{}
	}

	// Strip module path prefix to get relative directory
	relDir := strings.TrimPrefix(packagePath, modulePath+"/")

	result := make(map[string]string)
	for path, content := range repoContents {
		if path == relDir || strings.HasPrefix(path, relDir+"/") {
			result[path] = content
		}
	}

	return result
}

// ScopeAnalysis returns a new filtered AnalysisResult containing only data relevant to the given package paths.
// A package matches if its ImportPath equals one of the paths OR starts with one of them + "/".
func ScopeAnalysis(result *analyzer.AnalysisResult, packagePaths []string) *analyzer.AnalysisResult {
	matches := func(pkg string) bool {
		for _, pp := range packagePaths {
			if pkg == pp || strings.HasPrefix(pkg, pp+"/") {
				return true
			}
		}
		return false
	}

	scoped := &analyzer.AnalysisResult{
		Module: result.Module,
	}

	// Filter Packages
	for _, p := range result.Packages {
		if matches(p.ImportPath) {
			scoped.Packages = append(scoped.Packages, p)
		}
	}

	// Filter ImportGraph
	scoped.ImportGraph = make(map[string][]string)
	for key, vals := range result.ImportGraph {
		if matches(key) {
			scoped.ImportGraph[key] = vals
		}
	}

	// Filter CallGraph
	for _, c := range result.CallGraph {
		if matches(c.CallerPkg) {
			scoped.CallGraph = append(scoped.CallGraph, c)
		}
	}

	// Filter ExternalInteractions
	for _, e := range result.ExternalInteractions {
		if matches(e.PkgPath) {
			scoped.ExternalInteractions = append(scoped.ExternalInteractions, e)
		}
	}

	// Filter InterfaceImpls
	for _, i := range result.InterfaceImpls {
		if matches(i.StructPkg) {
			scoped.InterfaceImpls = append(scoped.InterfaceImpls, i)
		}
	}

	// Filter Entrypoints
	for _, e := range result.Entrypoints {
		if matches(e.PkgPath) {
			scoped.Entrypoints = append(scoped.Entrypoints, e)
		}
	}

	return scoped
}

// formatAnalysisJSON marshals the analysis result to indented JSON for prompt inclusion.
func formatAnalysisJSON(result *analyzer.AnalysisResult) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling analysis: %v", err)
	}
	return string(data)
}
