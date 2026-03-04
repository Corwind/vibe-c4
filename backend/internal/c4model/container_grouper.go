package c4model

import (
	"path"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// ContainerGroup represents a logical grouping of Go packages that form a C4 container.
type ContainerGroup struct {
	Name       string
	Technology string
	Packages   []string // import paths
	IsInfra    bool
	SystemKind string // for infrastructure: "database", "message_queue", "http_api", "grpc_service"
}

// groupPackagesIntoContainers groups analyzed packages into C4 container groups
// based on cmd/ entry points and external infrastructure interactions.
func groupPackagesIntoContainers(packages []analyzer.PackageInfo, importGraph map[string][]string, externalInteractions []analyzer.ExternalInteraction, modulePath string) []ContainerGroup {
	var cmdPkgs []analyzer.PackageInfo
	for _, pkg := range packages {
		if strings.HasPrefix(pkg.Dir, "cmd/") || pkg.Role == "cmd" {
			cmdPkgs = append(cmdPkgs, pkg)
		}
	}

	var groups []ContainerGroup

	switch {
	case len(cmdPkgs) == 1:
		// Single cmd package: one container with ALL packages
		allPaths := make([]string, 0, len(packages))
		for _, pkg := range packages {
			allPaths = append(allPaths, pkg.ImportPath)
		}
		name := path.Base(cmdPkgs[0].Dir)
		groups = append(groups, ContainerGroup{
			Name:       name,
			Technology: "Go",
			Packages:   allPaths,
		})

	case len(cmdPkgs) > 1:
		// Multiple cmd packages: trace imports to assign internal packages
		for _, cmd := range cmdPkgs {
			reachable := traceImportsTransitively(cmd.ImportPath, importGraph, modulePath)
			name := path.Base(cmd.Dir)
			groups = append(groups, ContainerGroup{
				Name:       name,
				Technology: "Go",
				Packages:   reachable,
			})
		}

	default:
		// No cmd packages: single Library container
		allPaths := make([]string, 0, len(packages))
		for _, pkg := range packages {
			allPaths = append(allPaths, pkg.ImportPath)
		}
		groups = append(groups, ContainerGroup{
			Name:       "Library",
			Technology: "Go",
			Packages:   allPaths,
		})
	}

	// Add infrastructure containers from external interactions
	groups = append(groups, buildInfraContainers(externalInteractions)...)

	return groups
}

// traceImportsTransitively returns all internal packages (those with modulePath prefix)
// reachable from root via BFS on importGraph. Includes the root itself.
func traceImportsTransitively(root string, importGraph map[string][]string, modulePath string) []string {
	visited := map[string]bool{}
	queue := []string{root}
	visited[root] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dep := range importGraph[current] {
			if visited[dep] {
				continue
			}
			if !strings.HasPrefix(dep, modulePath) {
				continue
			}
			visited[dep] = true
			queue = append(queue, dep)
		}
	}

	result := make([]string, 0, len(visited))
	for pkg := range visited {
		result = append(result, pkg)
	}
	return result
}

// buildInfraContainers creates deduplicated infrastructure ContainerGroups
// from external interactions, skipping inbound kinds (http_handler, grpc_server).
func buildInfraContainers(interactions []analyzer.ExternalInteraction) []ContainerGroup {
	type infraDef struct {
		name       string
		defaultTech string
		systemKind string
	}

	kindMap := map[analyzer.ExternalDependencyKind]infraDef{
		analyzer.ExtKindDatabase:      {name: "Database", defaultTech: "SQL", systemKind: "database"},
		analyzer.ExtKindKafkaProducer: {name: "Message Queue", defaultTech: "Kafka", systemKind: "message_queue"},
		analyzer.ExtKindKafkaConsumer: {name: "Message Queue", defaultTech: "Kafka", systemKind: "message_queue"},
		analyzer.ExtKindHTTPClient:    {name: "External HTTP API", defaultTech: "HTTP", systemKind: "http_api"},
		analyzer.ExtKindGRPCClient:    {name: "External gRPC Service", defaultTech: "gRPC", systemKind: "grpc_service"},
	}

	seen := map[string]bool{}
	var groups []ContainerGroup

	for _, interaction := range interactions {
		def, ok := kindMap[interaction.Kind]
		if !ok {
			continue
		}
		if seen[def.systemKind] {
			continue
		}
		seen[def.systemKind] = true

		tech := interaction.Technology
		if tech == "" {
			tech = def.defaultTech
		}

		groups = append(groups, ContainerGroup{
			Name:       def.name,
			Technology: tech,
			IsInfra:    true,
			SystemKind: def.systemKind,
		})
	}

	return groups
}
