package c4model_test

import (
	"strings"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func smartAnalysisResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/sample-project",
			GoVersion:  "1.22",
			DirectDeps: []analyzer.Dependency{
				{Path: "github.com/go-chi/chi/v5", Version: "v5.0.10"},
				{Path: "github.com/lib/pq", Version: "v1.10.9"},
				{Path: "github.com/stretchr/testify", Version: "v1.8.4"},
			},
		},
		Packages: []analyzer.PackageInfo{
			{
				Name:       "main",
				ImportPath: "github.com/example/sample-project/cmd/server",
				Dir:        "cmd/server",
				Role:       "cmd",
				GoFiles:    []string{"main.go"},
				Imports:    []string{"github.com/example/sample-project/internal/handler", "github.com/example/sample-project/internal/service"},
			},
			{
				Name:       "handler",
				ImportPath: "github.com/example/sample-project/internal/handler",
				Dir:        "internal/handler",
				Role:       "internal",
				GoFiles:    []string{"handler.go"},
				Imports:    []string{"github.com/example/sample-project/internal/service"},
				Structs: []analyzer.StructInfo{
					{
						Name:     "UserHandler",
						FilePath: "internal/handler/handler.go",
						Line:     8,
						Fields: []analyzer.FieldInfo{
							{Name: "svc", Type: "*service.UserService"},
						},
						Methods: []analyzer.MethodInfo{
							{Name: "ServeHTTP", FilePath: "internal/handler/handler.go", Line: 16},
							{Name: "GetUser", FilePath: "internal/handler/handler.go", Line: 20},
						},
					},
					// Config struct should be skipped
					{
						Name:     "Config",
						FilePath: "internal/handler/config.go",
						Fields: []analyzer.FieldInfo{
							{Name: "Port", Type: "int"},
							{Name: "Host", Type: "string"},
						},
					},
					// Logger should be skipped
					{
						Name:     "Logger",
						FilePath: "internal/handler/logger.go",
						Fields: []analyzer.FieldInfo{
							{Name: "level", Type: "string"},
						},
					},
				},
				Interfaces: []analyzer.InterfaceInfo{
					{
						Name: "RequestHandler",
						Methods: []analyzer.MethodInfo{
							{Name: "ServeHTTP"},
						},
					},
				},
			},
			{
				Name:       "service",
				ImportPath: "github.com/example/sample-project/internal/service",
				Dir:        "internal/service",
				Role:       "internal",
				GoFiles:    []string{"service.go"},
				Imports:    []string{"github.com/example/sample-project/internal/repo"},
				Structs: []analyzer.StructInfo{
					{
						Name:     "UserService",
						FilePath: "internal/service/service.go",
						Line:     13,
						Fields: []analyzer.FieldInfo{
							{Name: "repo", Type: "*repo.UserRepo"},
						},
						Methods: []analyzer.MethodInfo{
							{Name: "GetUser", FilePath: "internal/service/service.go", Line: 20},
							{Name: "CreateUser", FilePath: "internal/service/service.go", Line: 30},
						},
					},
				},
			},
			{
				Name:       "repo",
				ImportPath: "github.com/example/sample-project/internal/repo",
				Dir:        "internal/repo",
				Role:       "internal",
				GoFiles:    []string{"repo.go"},
				Structs: []analyzer.StructInfo{
					{
						Name:     "UserRepo",
						FilePath: "internal/repo/repo.go",
						Line:     8,
						Fields: []analyzer.FieldInfo{
							{Name: "db", Type: "*sql.DB"},
						},
						Methods: []analyzer.MethodInfo{
							{Name: "FindByID", FilePath: "internal/repo/repo.go", Line: 15},
							{Name: "Save", FilePath: "internal/repo/repo.go", Line: 25},
						},
					},
				},
			},
		},
		ImportGraph: map[string][]string{
			"github.com/example/sample-project/cmd/server":       {"github.com/example/sample-project/internal/handler", "github.com/example/sample-project/internal/service"},
			"github.com/example/sample-project/internal/handler": {"github.com/example/sample-project/internal/service"},
			"github.com/example/sample-project/internal/service": {"github.com/example/sample-project/internal/repo"},
		},
		Entrypoints: []analyzer.Entrypoint{
			{Kind: "http_handler", PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "UserHandler.ServeHTTP", Route: "/api/v1/users"},
			{Kind: "main", PkgPath: "github.com/example/sample-project/cmd/server", FuncName: "main"},
		},
		ExternalInteractions: []analyzer.ExternalInteraction{
			{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/example/sample-project/internal/repo", TypeName: "UserRepo", FuncName: "FindByID", Technology: "PostgreSQL"},
			{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/example/sample-project/internal/repo", TypeName: "UserRepo", FuncName: "Save", Technology: "PostgreSQL"},
			{Kind: analyzer.ExtKindHTTPHandler, PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "UserHandler.ServeHTTP", Technology: "HTTP"},
		},
		CallGraph: []analyzer.FunctionCall{
			{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "UserHandler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "UserService", CalleeFunc: "GetUser"},
			{CallerPkg: "github.com/example/sample-project/internal/service", CallerType: "UserService", CallerFunc: "GetUser", CalleePkg: "github.com/example/sample-project/internal/repo", CalleeType: "UserRepo", CalleeFunc: "FindByID"},
			// Self-call: should be skipped
			{CallerPkg: "github.com/example/sample-project/internal/service", CallerType: "UserService", CallerFunc: "GetUser", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "UserService", CalleeFunc: "CreateUser"},
			// Missing type: should be skipped
			{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "", CallerFunc: "New", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "UserService", CalleeFunc: "GetUser"},
			// External call: should be skipped
			{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "UserHandler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/go-chi/chi/v5", CalleeType: "Mux", CalleeFunc: "Use"},
		},
		InterfaceImpls: []analyzer.InterfaceImpl{
			{StructPkg: "github.com/example/sample-project/internal/handler", StructName: "UserHandler", InterfacePkg: "github.com/example/sample-project/internal/handler", InterfaceName: "RequestHandler"},
		},
	}
}

// --- Level 1: System Context ---

func TestSmartBuilder_NilResult(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	_, err := builder.BuildFromAnalysis(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "analysis result is nil")
}

func TestSmartBuilder_MainSystemCreated(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var mainSystem *c4model.System
	for i, s := range model.Systems {
		if !s.External {
			mainSystem = &model.Systems[i]
			break
		}
	}
	require.NotNil(t, mainSystem)
	assert.Equal(t, "sample-project", mainSystem.Name)
	assert.Contains(t, mainSystem.Description, "serves HTTP requests")
	assert.Contains(t, mainSystem.Description, "stores data in PostgreSQL")
}

func TestSmartBuilder_ActorsInferred(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var actors []c4model.System
	for _, s := range model.Systems {
		if s.SystemKind == "actor" {
			actors = append(actors, s)
		}
	}
	require.Len(t, actors, 1)
	assert.Equal(t, "User", actors[0].Name)
	assert.True(t, actors[0].External)
}

func TestSmartBuilder_NoLibraryExternalSystems(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	// chi, pq, testify should NOT appear as external systems
	for _, s := range model.Systems {
		assert.NotContains(t, s.Name, "chi")
		assert.NotContains(t, s.Name, "pq")
		assert.NotContains(t, s.Name, "testify")
	}
}

func TestSmartBuilder_InfraExternalSystems(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	extSystems := make(map[string]c4model.System)
	for _, s := range model.Systems {
		if s.External && s.SystemKind != "" && s.SystemKind != "actor" {
			extSystems[s.SystemKind] = s
		}
	}

	assert.Contains(t, extSystems, "database")
	assert.Equal(t, "Database", extSystems["database"].Name)
}

// --- Level 2: Containers ---

func TestSmartBuilder_ContainersGroupedByCmd(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	// Single cmd/server → 1 container named "server", NOT 4 containers
	assert.Len(t, model.Containers, 1)
	assert.Equal(t, "server", model.Containers[0].Name)
	assert.Equal(t, "Go", model.Containers[0].Technology)
}

func TestSmartBuilder_NoContainerPerPackage(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	// Should NOT have containers named after individual packages
	for _, c := range model.Containers {
		assert.NotEqual(t, "internal/handler", c.Name)
		assert.NotEqual(t, "internal/service", c.Name)
		assert.NotEqual(t, "internal/repo", c.Name)
	}
}

// --- Level 3: Components ---

func TestSmartBuilder_ComponentsFiltered(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	compNames := make(map[string]bool)
	for _, c := range model.Components {
		compNames[c.Name] = true
	}

	// Config and Logger should be filtered out
	assert.False(t, compNames["Config"], "Config should be filtered out")
	assert.False(t, compNames["Logger"], "Logger should be filtered out")

	// Real components should be present
	assert.True(t, compNames["UserHandler"], "UserHandler should be present")
	assert.True(t, compNames["UserService"], "UserService should be present")
	assert.True(t, compNames["UserRepo"], "UserRepo should be present")

	// Interface should be present
	assert.True(t, compNames["RequestHandler"], "RequestHandler interface should be present")
}

func TestSmartBuilder_ComponentDescriptions(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	for _, c := range model.Components {
		switch c.Name {
		case "UserHandler":
			assert.Contains(t, c.Description, "user")
			assert.Contains(t, c.Description, "requests")
		case "UserService":
			assert.Contains(t, c.Description, "user")
			assert.Contains(t, c.Description, "business logic")
		case "UserRepo":
			assert.Contains(t, c.Description, "user")
			assert.Contains(t, c.Description, "data access")
		}
	}
}

func TestSmartBuilder_EntrypointMarked(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var handler *c4model.Component
	for i, c := range model.Components {
		if c.Name == "UserHandler" {
			handler = &model.Components[i]
			break
		}
	}
	require.NotNil(t, handler)
	assert.True(t, handler.IsEntrypoint)
	assert.Equal(t, "http_handler", handler.EntrypointKind)
	assert.Equal(t, "/api/v1/users", handler.EntrypointRoute)
}

// --- Relationships ---

func TestSmartBuilder_ComponentCallRelationships(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var callRels []c4model.Relationship
	for _, r := range model.Relationships {
		if strings.HasPrefix(r.Description, "calls ") && r.Level == "component" {
			callRels = append(callRels, r)
		}
	}

	// Handler→Service and Service→Repo (self-calls, missing types, external calls filtered out)
	assert.Len(t, callRels, 2)
}

func TestSmartBuilder_InterfaceImplRelationships(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var implRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "implements" {
			implRels = append(implRels, r)
		}
	}

	assert.Len(t, implRels, 1)
}

func TestSmartBuilder_ComponentInfraRelationships(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	// UserRepo should have a relationship to the Database external system
	var infraRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "Reads and writes data" && r.Level == "component" {
			infraRels = append(infraRels, r)
		}
	}

	// 2 DB interactions for UserRepo, but should be deduplicated to 1 relationship
	assert.Len(t, infraRels, 1)
}

func TestSmartBuilder_NoActorsWithoutEntrypoints(t *testing.T) {
	result := &analyzer.AnalysisResult{
		Module: analyzer.ModuleInfo{
			ModulePath: "github.com/example/lib",
		},
		Packages: []analyzer.PackageInfo{
			{Name: "lib", ImportPath: "github.com/example/lib", Dir: ".", Role: "pkg"},
		},
		ImportGraph: map[string][]string{},
	}

	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	for _, s := range model.Systems {
		assert.NotEqual(t, "actor", s.SystemKind, "should not have actors without entrypoints")
	}
}

func TestSmartBuilder_RelationshipDescriptionsAreMeaningful(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		// No generic "imports" descriptions
		assert.NotEqual(t, "imports", r.Description, "should not use 'imports' as description")
		// "depends on" is allowed at component level (field-based deps), but not at system/container level
		if r.Level == "system" || r.Level == "container" {
			assert.NotEqual(t, "depends on", r.Description, "should not use 'depends on' at %s level", r.Level)
		}
	}
}

// --- Field Dependency Relationships ---

func TestSmartBuilder_FieldDependencyRelationships(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	var fieldDeps []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "depends on" && r.Level == "component" {
			fieldDeps = append(fieldDeps, r)
		}
	}

	// UserHandler has field "svc *service.UserService" → depends on UserService
	// UserService has field "repo *repo.UserRepo" → depends on UserRepo
	require.GreaterOrEqual(t, len(fieldDeps), 2, "should have at least 2 field-based dependency relationships")

	depPairs := make(map[string]bool)
	for _, r := range fieldDeps {
		src := model.FindComponent(r.SourceID)
		tgt := model.FindComponent(r.TargetID)
		if src != nil && tgt != nil {
			depPairs[src.Name+"->"+tgt.Name] = true
		}
	}

	assert.True(t, depPairs["UserHandler->UserService"], "UserHandler should depend on UserService via field")
	assert.True(t, depPairs["UserService->UserRepo"], "UserService should depend on UserRepo via field")
}

func TestSmartBuilder_EmbeddedStructExtendsRelationship(t *testing.T) {
	result := smartAnalysisResult()
	// Add a struct with an embedded field
	result.Packages = append(result.Packages, analyzer.PackageInfo{
		Name:       "models",
		ImportPath: "github.com/example/sample-project/internal/models",
		Dir:        "internal/models",
		Role:       "internal",
		GoFiles:    []string{"models.go"},
		Structs: []analyzer.StructInfo{
			{
				Name:     "BaseModel",
				FilePath: "internal/models/models.go",
				Line:     5,
				Methods: []analyzer.MethodInfo{
					{Name: "GetID", Params: []string{}, Returns: []string{"string"}},
				},
			},
			{
				Name:     "UserModel",
				FilePath: "internal/models/models.go",
				Line:     15,
				Fields: []analyzer.FieldInfo{
					{Name: "", Type: "BaseModel", Embedded: true},
					{Name: "Email", Type: "string"},
				},
				Methods: []analyzer.MethodInfo{
					{Name: "Validate", Params: []string{}, Returns: []string{"error"}},
				},
			},
		},
	})
	// Add models package to import graph so it gets grouped into a container.
	result.ImportGraph["github.com/example/sample-project/cmd/server"] = append(
		result.ImportGraph["github.com/example/sample-project/cmd/server"],
		"github.com/example/sample-project/internal/models",
	)

	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	var extendsRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "extends" {
			extendsRels = append(extendsRels, r)
		}
	}

	require.Len(t, extendsRels, 1, "should have exactly 1 extends relationship")
	src := model.FindComponent(extendsRels[0].SourceID)
	tgt := model.FindComponent(extendsRels[0].TargetID)
	require.NotNil(t, src)
	require.NotNil(t, tgt)
	assert.Equal(t, "UserModel", src.Name)
	assert.Equal(t, "BaseModel", tgt.Name)
}

func TestSmartBuilder_CallRelationshipsIncludeMethodNames(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if !strings.HasPrefix(r.Description, "calls ") || r.Level != "component" {
			continue
		}
		src := model.FindComponent(r.SourceID)
		tgt := model.FindComponent(r.TargetID)
		if src == nil || tgt == nil {
			continue
		}

		// Handler→Service should mention "GetUser"
		if src.Name == "UserHandler" && tgt.Name == "UserService" {
			assert.Contains(t, r.Description, "GetUser", "call from Handler to Service should include GetUser")
		}
		// Service→Repo should mention "FindByID"
		if src.Name == "UserService" && tgt.Name == "UserRepo" {
			assert.Contains(t, r.Description, "FindByID", "call from Service to Repo should include FindByID")
		}
	}
}

func TestSmartBuilder_ComponentMetadataPopulated(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	for _, comp := range model.Components {
		switch comp.Name {
		case "UserHandler":
			assert.Equal(t, "controller", comp.Role)
			assert.Equal(t, "github.com/example/sample-project/internal/handler", comp.PackagePath)
			assert.NotEmpty(t, comp.Methods, "UserHandler should have methods")
			assert.NotEmpty(t, comp.Fields, "UserHandler should have fields")
			// Check method format
			found := false
			for _, m := range comp.Methods {
				if strings.HasPrefix(m, "ServeHTTP(") {
					found = true
				}
			}
			assert.True(t, found, "should have ServeHTTP method signature")

		case "UserService":
			assert.Equal(t, "service", comp.Role)
			assert.Equal(t, "github.com/example/sample-project/internal/service", comp.PackagePath)
			assert.Len(t, comp.Methods, 2, "UserService should have 2 methods")

		case "UserRepo":
			assert.Equal(t, "repository", comp.Role)
			assert.Equal(t, "github.com/example/sample-project/internal/repo", comp.PackagePath)

		case "RequestHandler":
			// Interface — should have methods but no Role (interfaces don't have roles)
			assert.Empty(t, comp.Role)
			assert.NotEmpty(t, comp.Methods)
		}
	}
}

func TestSmartBuilder_ContainerDescriptionSet(t *testing.T) {
	builder := c4model.NewSmartModelBuilder()
	model, err := builder.BuildFromAnalysis(smartAnalysisResult())
	require.NoError(t, err)

	for _, c := range model.Containers {
		assert.NotEmpty(t, c.Description, "container %s should have a description", c.Name)
	}
}

func TestFormatMethodSignatures(t *testing.T) {
	methods := []analyzer.MethodInfo{
		{Name: "GetUser", Params: []string{"ctx context.Context", "id string"}, Returns: []string{"*User", "error"}},
		{Name: "Close", Params: nil, Returns: nil},
		{Name: "Count", Params: nil, Returns: []string{"int"}},
	}

	sigs := c4model.FormatMethodSignatures(methods)
	require.Len(t, sigs, 3)
	assert.Equal(t, "GetUser(ctx context.Context, id string) (*User, error)", sigs[0])
	assert.Equal(t, "Close()", sigs[1])
	assert.Equal(t, "Count() int", sigs[2])
}

func TestFormatFieldDescriptors(t *testing.T) {
	fields := []analyzer.FieldInfo{
		{Name: "svc", Type: "*service.UserService"},
		{Name: "", Type: "BaseModel", Embedded: true},
		{Name: "db", Type: "*sql.DB"},
	}

	descs := c4model.FormatFieldDescriptors(fields)
	require.Len(t, descs, 3)
	assert.Equal(t, "svc *service.UserService", descs[0])
	assert.Equal(t, "BaseModel (embedded)", descs[1])
	assert.Equal(t, "db *sql.DB", descs[2])
}

func TestResolveFieldTypeToComponentID(t *testing.T) {
	pkgNameToPath := map[string]string{
		"service": "github.com/example/project/internal/service",
		"repo":    "github.com/example/project/internal/repo",
	}
	structToCompID := map[string]string{
		"github.com/example/project/internal/service/UserService": "component-service-userservice",
		"github.com/example/project/internal/repo/UserRepo":       "component-repo-userrepo",
		"github.com/example/project/internal/handler/LocalStruct": "component-handler-localstruct",
	}

	tests := []struct {
		name       string
		typeName   string
		currentPkg string
		expected   string
	}{
		{"qualified pointer", "*service.UserService", "github.com/example/project/internal/handler", "component-service-userservice"},
		{"qualified slice", "[]repo.UserRepo", "github.com/example/project/internal/handler", "component-repo-userrepo"},
		{"unqualified local", "LocalStruct", "github.com/example/project/internal/handler", "component-handler-localstruct"},
		{"primitive type", "string", "github.com/example/project/internal/handler", ""},
		{"external package", "*http.Client", "github.com/example/project/internal/handler", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c4model.ResolveFieldTypeToComponentID(tt.typeName, tt.currentPkg, pkgNameToPath, structToCompID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
