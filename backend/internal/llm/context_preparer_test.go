package llm

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepare_CompressesPackages(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/app"},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "handlers",
				ImportPath: "github.com/example/app/handlers",
				Structs: []analyzer.StructInfo{
					{Name: "UserHandler", FilePath: "/src/handlers/user.go", Line: 10},
					{Name: "OrderHandler", FilePath: "/src/handlers/order.go", Line: 5},
				},
				Interfaces: []analyzer.InterfaceInfo{
					{Name: "Service", FilePath: "/src/handlers/iface.go", Line: 3},
				},
				Functions: []analyzer.FunctionInfo{
					{Name: "NewUserHandler", FilePath: "/src/handlers/user.go", Line: 20},
				},
			},
		},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	require.Len(t, input.PackageSummaries, 1)
	pkg := input.PackageSummaries[0]

	assert.Equal(t, "handlers", pkg.Name)
	assert.Equal(t, "github.com/example/app/handlers", pkg.ImportPath)
	assert.Equal(t, []string{"UserHandler", "OrderHandler"}, pkg.Structs)
	assert.Equal(t, []string{"Service"}, pkg.Interfaces)
	assert.Equal(t, []string{"NewUserHandler"}, pkg.Functions)
}

func TestPrepare_SplitsImports(t *testing.T) {
	modulePath := "github.com/example/app"

	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: modulePath},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "service",
				ImportPath: modulePath + "/service",
				Imports: []string{
					modulePath + "/repository",
					modulePath + "/domain",
					"github.com/go-chi/chi/v5",
					"fmt",
				},
			},
		},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	require.Len(t, input.PackageSummaries, 1)
	pkg := input.PackageSummaries[0]

	assert.Equal(t, []string{
		modulePath + "/repository",
		modulePath + "/domain",
	}, pkg.InternalImports)

	assert.Equal(t, []string{
		"github.com/go-chi/chi/v5",
		"fmt",
	}, pkg.ExternalImports)
}

func TestPrepare_DeduplicatesExternalInteractions(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/app"},
		ExternalInteractions: []analyzer.ExternalInteraction{
			{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL", PkgPath: "pkg/repo", FuncName: "Query"},
			{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL", PkgPath: "pkg/other", FuncName: "Exec"},
			{Kind: analyzer.ExtKindHTTPClient, Technology: "net/http", PkgPath: "pkg/gateway", FuncName: "Do"},
		},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	require.Len(t, input.ExternalInteractions, 2)

	kinds := make(map[string]string)
	for _, ei := range input.ExternalInteractions {
		kinds[ei.Kind] = ei.Technology
	}
	assert.Equal(t, "PostgreSQL", kinds["database"])
	assert.Equal(t, "net/http", kinds["http_client"])
}

func TestPrepare_FiltersCrossTypeCallEdges(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/app"},
		CallGraph: []analyzer.FunctionCall{
			// Cross-type edge: should pass
			{CallerPkg: "pkg/a", CallerType: "ServiceA", CallerFunc: "Do", CalleePkg: "pkg/b", CalleeType: "RepoB", CalleeFunc: "Save"},
			// Same-type, same-pkg: should be filtered out
			{CallerPkg: "pkg/a", CallerType: "ServiceA", CallerFunc: "Foo", CalleePkg: "pkg/a", CalleeType: "ServiceA", CalleeFunc: "Bar"},
			// Empty CallerType: should be filtered out
			{CallerPkg: "pkg/a", CallerType: "", CallerFunc: "Init", CalleePkg: "pkg/b", CalleeType: "RepoB", CalleeFunc: "New"},
			// Empty CalleeType: should be filtered out
			{CallerPkg: "pkg/a", CallerType: "ServiceA", CallerFunc: "Run", CalleePkg: "pkg/c", CalleeType: "", CalleeFunc: "Helper"},
		},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	require.Len(t, input.CallGraphEdges, 1)
	edge := input.CallGraphEdges[0]
	assert.Equal(t, "pkg/a", edge.CallerPkg)
	assert.Equal(t, "ServiceA", edge.CallerType)
	assert.Equal(t, "pkg/b", edge.CalleePkg)
	assert.Equal(t, "RepoB", edge.CalleeType)
}

func TestPrepare_CollectsDirectDeps(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/app",
			DirectDeps: []analyzer.Dependency{
				{Path: "github.com/go-chi/chi/v5", Version: "v5.2.0"},
				{Path: "github.com/stretchr/testify", Version: "v1.9.0"},
			},
		},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	assert.Equal(t, []string{
		"github.com/go-chi/chi/v5",
		"github.com/stretchr/testify",
	}, input.DirectDeps)
}

func TestPrepare_SetsProjectNameFromModulePath(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{ModulePath: "github.com/example/my-awesome-app"},
	}

	cp := &ContextPreparer{}
	input := cp.Prepare(result, "/tmp/fake")

	assert.Equal(t, "my-awesome-app", input.ProjectName)
	assert.Equal(t, "github.com/example/my-awesome-app", input.ModulePath)
}
