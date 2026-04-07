package c4model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// BuildSystemPrompt returns a comprehensive system prompt that instructs Claude
// to generate a C4 model from Go project analysis data.
func BuildSystemPrompt() string {
	return `You are an expert software architect specializing in Simon Brown's C4 model for visualizing software architecture.

## C4 Model Levels

### Level 1 - System Context
The highest level of abstraction. Shows the software system being modeled, its users (actors/personas), and the external systems it interacts with. Each system has an ID prefixed with "system-".

### Level 2 - Container
Shows the deployable units within the main system: services, applications, databases, file systems, message queues. Each container belongs to exactly one system (referenced by system_id). Container IDs are prefixed with "container-".

### Level 3 - Component
Shows the key structural building blocks within each container: interfaces, services, repositories, controllers, domain models. Each component belongs to exactly one container (referenced by container_id). Component IDs are prefixed with "component-".

### Level 4 - Code
Shows the methods, fields, and functions within each component. Each code element belongs to exactly one component (referenced by component_id). Code element IDs are prefixed with "code-".

## ID Generation Rules
All IDs MUST be generated as "<prefix>-<sanitized-name>" where:
- The prefix is one of: "system-", "container-", "component-", "code-"
- The sanitized name replaces "/" and "." characters with "-"
- Example: a container for package "internal/handler" gets ID "container-internal-handler"

## Relationship Rules
- Every relationship MUST reference valid source_id and target_id that exist as entity IDs in the model
- Each relationship has a "level" field indicating the diagram level it belongs to: "context", "container", or "component"
- Relationships describe directed dependencies or interactions between elements

## Quality Guidelines
- Descriptions should explain the PURPOSE and RESPONSIBILITY of the element, not just restate its name
- For example, instead of "The user service" write "Manages user lifecycle including registration, authentication, and profile management"
- External systems should clearly state what capability they provide
- Relationships should describe what data or functionality flows between elements

## Output JSON Schema
Return a JSON object matching this exact structure:

{
  "project_id": "string",
  "systems": [
    {
      "id": "string (system-<sanitized>)",
      "name": "string",
      "description": "string",
      "external": false,
      "module_path": "string (optional)",
      "system_kind": "string (optional)",
      "container_ids": ["string (optional)"]
    }
  ],
  "containers": [
    {
      "id": "string (container-<sanitized>)",
      "name": "string",
      "description": "string",
      "technology": "string",
      "package_path": "string",
      "system_id": "string (must reference a valid system ID)",
      "component_ids": ["string (optional)"]
    }
  ],
  "components": [
    {
      "id": "string (component-<sanitized>)",
      "name": "string",
      "description": "string",
      "technology": "string",
      "type": "struct|interface|function",
      "container_id": "string (must reference a valid container ID)",
      "code_elements": ["string (optional)"],
      "is_entrypoint": false,
      "entrypoint_kind": "string (optional)",
      "entrypoint_route": "string (optional)",
      "role": "string (optional)",
      "package_path": "string (optional)",
      "methods": ["string (optional)"],
      "fields": ["string (optional)"]
    }
  ],
  "code_elements": [
    {
      "id": "string (code-<sanitized>)",
      "name": "string",
      "description": "string",
      "type": "function|method|field",
      "component_id": "string (must reference a valid component ID)",
      "file_path": "string (optional)",
      "line": 0
    }
  ],
  "relationships": [
    {
      "source_id": "string (must reference a valid entity ID)",
      "target_id": "string (must reference a valid entity ID)",
      "description": "string",
      "technology": "string (optional)",
      "level": "context|container|component"
    }
  ]
}

IMPORTANT: Return ONLY valid JSON. Do NOT wrap the output in markdown code fences or any other formatting. The response must start with "{" and end with "}".`
}

// BuildUserPrompt builds the user message for Claude from analysis results and source code.
func BuildUserPrompt(result *analyzer.AnalysisResult, repoContents map[string]string) string {
	var sb strings.Builder

	// Section 1: Static Analysis Results
	sb.WriteString("## Static Analysis Results\n\n")
	analysisJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error serializing analysis: %v\n", err))
	} else {
		sb.Write(analysisJSON)
	}
	sb.WriteString("\n\n")

	// Section 2: Source Code
	sb.WriteString("## Source Code\n\n")

	// Sort file paths for deterministic output
	paths := make([]string, 0, len(repoContents))
	for p := range repoContents {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		content := repoContents[p]
		sb.WriteString(fmt.Sprintf("### %s\n```go\n%s\n```\n\n", p, content))
	}

	// Section 3: Instructions
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Based on the static analysis results and source code above, generate a complete C4 model JSON for this Go project. Ensure all 4 levels are populated with proper parent-child relationships.\n")

	return sb.String()
}

// EstimateTokens returns a conservative approximate token count for the given content.
// It uses a heuristic of 1 token per 3 characters (conservative for Go source code
// which has many keywords and symbols that tokenize to individual tokens).
func EstimateTokens(content string) int {
	return len(content) / 3
}

// ReadFullRepository reads all relevant Go source files from the project directory,
// respecting a token budget and prioritizing important files.
func ReadFullRepository(projectPath string, tokenBudget int) (map[string]string, error) {
	type fileEntry struct {
		relPath string
		content string
		tier    int
	}

	var files []fileEntry

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "testdata" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}

		// Include go.mod always
		if info.Name() == "go.mod" {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			files = append(files, fileEntry{relPath: relPath, content: string(content), tier: 1})
			return nil
		}

		// Only include .go files (excluding tests)
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		contentStr := string(content)

		tier := classifyFileTier(relPath, contentStr)

		files = append(files, fileEntry{relPath: relPath, content: contentStr, tier: tier})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking project directory: %w", err)
	}

	// Sort by tier (ascending), then by path for determinism
	sort.Slice(files, func(i, j int) bool {
		if files[i].tier != files[j].tier {
			return files[i].tier < files[j].tier
		}
		return files[i].relPath < files[j].relPath
	})

	// Check if everything fits in budget
	totalTokens := 0
	for _, f := range files {
		totalTokens += EstimateTokens(f.content)
	}

	result := make(map[string]string)

	if totalTokens <= tokenBudget {
		for _, f := range files {
			result[f.relPath] = f.content
		}
		return result, nil
	}

	// Budget exceeded: add files tier by tier
	usedTokens := 0
	for _, f := range files {
		tokens := EstimateTokens(f.content)
		if usedTokens+tokens > tokenBudget {
			continue
		}
		usedTokens += tokens
		result[f.relPath] = f.content
	}

	return result, nil
}

// classifyFileTier determines the priority tier of a Go source file.
// Tier 1 (highest): go.mod, main.go, files containing "interface " keyword
// Tier 2: handler, controller, service, repository, server paths
// Tier 3: model, types, domain paths
// Tier 4: everything else
func classifyFileTier(relPath, content string) int {
	base := filepath.Base(relPath)
	lower := strings.ToLower(relPath)

	// Tier 1: main.go or files with interfaces
	if base == "main.go" {
		return 1
	}
	if strings.Contains(content, "interface ") {
		return 1
	}

	// Tier 2: handler, controller, service, repository, server
	for _, keyword := range []string{"handler", "controller", "service", "repository", "server"} {
		if strings.Contains(lower, keyword) {
			return 2
		}
	}

	// Tier 3: model, types, domain
	for _, keyword := range []string{"model", "types", "domain"} {
		if strings.Contains(lower, keyword) {
			return 3
		}
	}

	return 4
}
