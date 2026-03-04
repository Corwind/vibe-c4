package c4model

import (
	"sort"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
)

func TestGroupPackages_SingleCmd(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "main", ImportPath: "example.com/app/cmd/server", Dir: "cmd/server", Role: "cmd"},
		{Name: "handler", ImportPath: "example.com/app/internal/handler", Dir: "internal/handler", Role: "internal"},
		{Name: "service", ImportPath: "example.com/app/internal/service", Dir: "internal/service", Role: "internal"},
		{Name: "repo", ImportPath: "example.com/app/internal/repo", Dir: "internal/repo", Role: "internal"},
	}
	importGraph := map[string][]string{
		"example.com/app/cmd/server":        {"example.com/app/internal/handler", "example.com/app/internal/service"},
		"example.com/app/internal/handler":   {"example.com/app/internal/service"},
		"example.com/app/internal/service":   {"example.com/app/internal/repo"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, nil, "example.com/app")

	// Single cmd → 1 container with ALL packages
	assert.Len(t, groups, 1)
	assert.Equal(t, "server", groups[0].Name)
	assert.Equal(t, "Go", groups[0].Technology)
	assert.False(t, groups[0].IsInfra)

	sort.Strings(groups[0].Packages)
	expected := []string{
		"example.com/app/cmd/server",
		"example.com/app/internal/handler",
		"example.com/app/internal/repo",
		"example.com/app/internal/service",
	}
	sort.Strings(expected)
	assert.Equal(t, expected, groups[0].Packages)
}

func TestGroupPackages_MultipleCmds(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "main", ImportPath: "example.com/app/cmd/api", Dir: "cmd/api", Role: "cmd"},
		{Name: "main", ImportPath: "example.com/app/cmd/worker", Dir: "cmd/worker", Role: "cmd"},
		{Name: "handler", ImportPath: "example.com/app/internal/handler", Dir: "internal/handler", Role: "internal"},
		{Name: "service", ImportPath: "example.com/app/internal/service", Dir: "internal/service", Role: "internal"},
		{Name: "queue", ImportPath: "example.com/app/internal/queue", Dir: "internal/queue", Role: "internal"},
	}
	importGraph := map[string][]string{
		"example.com/app/cmd/api":            {"example.com/app/internal/handler"},
		"example.com/app/cmd/worker":         {"example.com/app/internal/queue"},
		"example.com/app/internal/handler":   {"example.com/app/internal/service"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, nil, "example.com/app")

	// 2 cmd packages → 2 containers
	assert.Len(t, groups, 2)

	// Sort groups by name for deterministic assertions
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

	// api container: cmd/api → handler → service
	assert.Equal(t, "api", groups[0].Name)
	assert.Equal(t, "Go", groups[0].Technology)
	sort.Strings(groups[0].Packages)
	expectedAPI := []string{
		"example.com/app/cmd/api",
		"example.com/app/internal/handler",
		"example.com/app/internal/service",
	}
	sort.Strings(expectedAPI)
	assert.Equal(t, expectedAPI, groups[0].Packages)

	// worker container: cmd/worker → queue
	assert.Equal(t, "worker", groups[1].Name)
	assert.Equal(t, "Go", groups[1].Technology)
	sort.Strings(groups[1].Packages)
	expectedWorker := []string{
		"example.com/app/cmd/worker",
		"example.com/app/internal/queue",
	}
	sort.Strings(expectedWorker)
	assert.Equal(t, expectedWorker, groups[1].Packages)
}

func TestGroupPackages_NoCmd(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "pkg", ImportPath: "example.com/lib/pkg", Dir: "pkg", Role: "pkg"},
		{Name: "util", ImportPath: "example.com/lib/util", Dir: "util", Role: "internal"},
	}
	importGraph := map[string][]string{
		"example.com/lib/pkg": {"example.com/lib/util"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, nil, "example.com/lib")

	assert.Len(t, groups, 1)
	assert.Equal(t, "Library", groups[0].Name)
	assert.Equal(t, "Go", groups[0].Technology)
	assert.False(t, groups[0].IsInfra)

	sort.Strings(groups[0].Packages)
	expected := []string{"example.com/lib/pkg", "example.com/lib/util"}
	sort.Strings(expected)
	assert.Equal(t, expected, groups[0].Packages)
}

func TestGroupPackages_InfraContainers(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "main", ImportPath: "example.com/app/cmd/server", Dir: "cmd/server", Role: "cmd"},
	}
	importGraph := map[string][]string{}
	interactions := []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL", PkgPath: "example.com/app/internal/repo"},
		{Kind: analyzer.ExtKindKafkaProducer, Technology: "Kafka", PkgPath: "example.com/app/internal/events"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, interactions, "example.com/app")

	// 1 cmd container + 2 infra containers
	assert.Len(t, groups, 3)

	infraGroups := make(map[string]ContainerGroup)
	for _, g := range groups {
		if g.IsInfra {
			infraGroups[g.SystemKind] = g
		}
	}

	db, ok := infraGroups["database"]
	assert.True(t, ok, "expected database infrastructure container")
	assert.Equal(t, "Database", db.Name)
	assert.Equal(t, "PostgreSQL", db.Technology)
	assert.True(t, db.IsInfra)

	mq, ok := infraGroups["message_queue"]
	assert.True(t, ok, "expected message_queue infrastructure container")
	assert.Equal(t, "Message Queue", mq.Name)
	assert.Equal(t, "Kafka", mq.Technology)
	assert.True(t, mq.IsInfra)
}

func TestGroupPackages_InfraDedup(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "main", ImportPath: "example.com/app/cmd/server", Dir: "cmd/server", Role: "cmd"},
	}
	importGraph := map[string][]string{}
	interactions := []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL", PkgPath: "example.com/app/internal/repo"},
		{Kind: analyzer.ExtKindDatabase, Technology: "MySQL", PkgPath: "example.com/app/internal/other"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, interactions, "example.com/app")

	// 1 cmd container + 1 deduplicated infra container
	assert.Len(t, groups, 2)

	var infraCount int
	for _, g := range groups {
		if g.IsInfra && g.SystemKind == "database" {
			infraCount++
			// First match wins for technology
			assert.Equal(t, "PostgreSQL", g.Technology)
		}
	}
	assert.Equal(t, 1, infraCount, "database infra should be deduplicated to 1")
}

func TestGroupPackages_SkipsInbound(t *testing.T) {
	packages := []analyzer.PackageInfo{
		{Name: "main", ImportPath: "example.com/app/cmd/server", Dir: "cmd/server", Role: "cmd"},
	}
	importGraph := map[string][]string{}
	interactions := []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindHTTPHandler, PkgPath: "example.com/app/internal/handler"},
		{Kind: analyzer.ExtKindGRPCServer, PkgPath: "example.com/app/internal/grpc"},
	}

	groups := groupPackagesIntoContainers(packages, importGraph, interactions, "example.com/app")

	// Only 1 cmd container, no infra for inbound handlers
	assert.Len(t, groups, 1)
	assert.False(t, groups[0].IsInfra)
}

func TestTraceImportsTransitively(t *testing.T) {
	importGraph := map[string][]string{
		"example.com/app/cmd/server":        {"example.com/app/internal/handler", "example.com/app/internal/service", "fmt"},
		"example.com/app/internal/handler":   {"example.com/app/internal/service", "net/http"},
		"example.com/app/internal/service":   {"example.com/app/internal/repo"},
		"example.com/app/internal/repo":      {"database/sql"},
	}

	result := traceImportsTransitively("example.com/app/cmd/server", importGraph, "example.com/app")

	sort.Strings(result)
	expected := []string{
		"example.com/app/cmd/server",
		"example.com/app/internal/handler",
		"example.com/app/internal/repo",
		"example.com/app/internal/service",
	}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}
