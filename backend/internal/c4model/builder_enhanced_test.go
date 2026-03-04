package c4model_test

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enhancedAnalysisResult() *analyzer.AnalysisResult {
	base := sampleAnalysisResult()
	base.ExternalInteractions = []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/example/sample-project/internal/service", FuncName: "Query", Technology: "PostgreSQL"},
		{Kind: analyzer.ExtKindDatabase, PkgPath: "github.com/example/sample-project/internal/service", FuncName: "Exec", Technology: "PostgreSQL"},
		{Kind: analyzer.ExtKindKafkaProducer, PkgPath: "github.com/example/sample-project/internal/service", FuncName: "Produce", Technology: "Kafka"},
		{Kind: analyzer.ExtKindKafkaConsumer, PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "Consume", Technology: "Kafka"},
		{Kind: analyzer.ExtKindHTTPClient, PkgPath: "github.com/example/sample-project/internal/service", FuncName: "Do", Technology: "HTTP"},
		{Kind: analyzer.ExtKindGRPCClient, PkgPath: "github.com/example/sample-project/internal/service", FuncName: "Invoke", Technology: "gRPC"},
		{Kind: analyzer.ExtKindHTTPHandler, PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "Handler.ServeHTTP", Technology: "HTTP"},
		{Kind: analyzer.ExtKindGRPCServer, PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "Serve", Technology: "gRPC"},
	}
	base.CallGraph = []analyzer.FunctionCall{
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "Handler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "Service", CalleeFunc: "DoWork"},
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "Handler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "Service", CalleeFunc: "Stop"},
		// Duplicate edge: same source->target, different func
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "Handler", CallerFunc: "Init", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "Service", CalleeFunc: "DoWork"},
		// Self-call: should be skipped
		{CallerPkg: "github.com/example/sample-project/internal/service", CallerType: "Service", CallerFunc: "DoWork", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "Service", CalleeFunc: "Stop"},
		// No CallerType: should be skipped
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "", CallerFunc: "New", CalleePkg: "github.com/example/sample-project/internal/service", CalleeType: "Service", CalleeFunc: "DoWork"},
		// No CalleeType: should be skipped
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "Handler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/example/sample-project/pkg/utils", CalleeType: "", CalleeFunc: "Log"},
		// External call: should be skipped
		{CallerPkg: "github.com/example/sample-project/internal/handler", CallerType: "Handler", CallerFunc: "ServeHTTP", CalleePkg: "github.com/go-chi/chi/v5", CalleeType: "Mux", CalleeFunc: "Use"},
	}
	base.InterfaceImpls = []analyzer.InterfaceImpl{
		{StructPkg: "github.com/example/sample-project/internal/service", StructName: "Service", InterfacePkg: "github.com/example/sample-project/internal/service", InterfaceName: "Worker"},
		// Non-existent component: should be skipped
		{StructPkg: "github.com/example/sample-project/internal/service", StructName: "NonExistent", InterfacePkg: "github.com/example/sample-project/internal/service", InterfaceName: "Worker"},
	}
	base.Entrypoints = []analyzer.Entrypoint{
		{Kind: "http_handler", PkgPath: "github.com/example/sample-project/internal/handler", FuncName: "Handler.ServeHTTP", Route: "/api/v1"},
		{Kind: "main", PkgPath: "github.com/example/sample-project/cmd/server", FuncName: "main"},
	}
	return base
}

func TestBuildExternalSystems_CreatesSystemsForEachKind(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	extSystems := make(map[string]c4model.System)
	for _, s := range model.Systems {
		if s.External && s.SystemKind != "" {
			extSystems[s.SystemKind] = s
		}
	}

	assert.Contains(t, extSystems, "database")
	assert.Equal(t, "Database", extSystems["database"].Name)

	assert.Contains(t, extSystems, "message_queue")
	assert.Equal(t, "Message Queue", extSystems["message_queue"].Name)

	assert.Contains(t, extSystems, "http_api")
	assert.Equal(t, "External HTTP API", extSystems["http_api"].Name)

	assert.Contains(t, extSystems, "grpc_service")
	assert.Equal(t, "gRPC Service", extSystems["grpc_service"].Name)
}

func TestBuildExternalSystems_DeduplicatesByKind(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	// Count database systems (we have 2 database interactions but should get only 1 system)
	var dbCount int
	for _, s := range model.Systems {
		if s.SystemKind == "database" {
			dbCount++
		}
	}
	assert.Equal(t, 1, dbCount, "should deduplicate database systems")

	// Count message queue systems (we have kafka_producer and kafka_consumer but should get 1 system)
	var mqCount int
	for _, s := range model.Systems {
		if s.SystemKind == "message_queue" {
			mqCount++
		}
	}
	assert.Equal(t, 1, mqCount, "should deduplicate kafka producer/consumer into one message queue system")
}

func TestBuildExternalSystems_SkipsInboundHandlers(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	// http_handler and grpc_server should NOT create external systems
	for _, s := range model.Systems {
		assert.NotEqual(t, "http_handler", s.SystemKind)
		assert.NotEqual(t, "grpc_server", s.SystemKind)
	}
}

func TestBuildExternalSystems_CreatesRelationships(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	// Find system-level relationships to external systems with SystemKind
	extSystemIDs := make(map[string]bool)
	for _, s := range model.Systems {
		if s.SystemKind != "" {
			extSystemIDs[s.ID] = true
		}
	}

	var relCount int
	for _, r := range model.Relationships {
		if r.Level == "system" && extSystemIDs[r.TargetID] {
			relCount++
		}
	}
	assert.Equal(t, 4, relCount, "should have system-level relationships to each external system kind")
}

func TestBuildComponentRelationships_CreatesCallEdges(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	var callRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "calls" && r.Level == "component" {
			callRels = append(callRels, r)
		}
	}

	// Only Handler->Service should exist (deduplicated, self-calls skipped, missing types skipped, external skipped)
	assert.Len(t, callRels, 1, "should have exactly one deduplicated component call edge")
	assert.Equal(t, "Go", callRels[0].Technology)
}

func TestBuildComponentRelationships_SkipsSelfCalls(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if r.Description == "calls" && r.Level == "component" {
			assert.NotEqual(t, r.SourceID, r.TargetID, "should not have self-referencing call edges")
		}
	}
}

func TestBuildComponentRelationships_SkipsExternalCalls(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if r.Description == "calls" && r.Level == "component" {
			// Neither source nor target should reference external packages
			assert.NotContains(t, r.SourceID, "go-chi")
			assert.NotContains(t, r.TargetID, "go-chi")
		}
	}
}

func TestBuildInterfaceImplRelationships_CreatesImplementsEdges(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	var implRels []c4model.Relationship
	for _, r := range model.Relationships {
		if r.Description == "implements" {
			implRels = append(implRels, r)
		}
	}

	// Only Service->Worker should exist (NonExistent component should be skipped)
	assert.Len(t, implRels, 1, "should have exactly one implements edge")
	assert.Equal(t, "component", implRels[0].Level)
}

func TestBuildInterfaceImplRelationships_SkipsNonExistentComponents(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	// NonExistent struct should not appear in any implements relationship
	for _, r := range model.Relationships {
		if r.Description == "implements" {
			assert.NotContains(t, r.SourceID, "NonExistent")
		}
	}
}

func TestMarkEntrypoints_HTTPHandler(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	comp := model.FindComponent(findComponentIDByName(model, "Handler"))
	require.NotNil(t, comp, "should find Handler component")

	assert.True(t, comp.IsEntrypoint, "Handler should be marked as entrypoint")
	assert.Equal(t, "http_handler", comp.EntrypointKind)
	assert.Equal(t, "/api/v1", comp.EntrypointRoute)
}

func TestMarkEntrypoints_Main(t *testing.T) {
	builder := c4model.NewModelBuilder()
	result := enhancedAnalysisResult()

	// Add a struct to cmd/server so we can check it gets marked as entrypoint
	for i := range result.Packages {
		if result.Packages[i].ImportPath == "github.com/example/sample-project/cmd/server" {
			result.Packages[i].Structs = []analyzer.StructInfo{
				{Name: "App", FilePath: "cmd/server/main.go", Line: 5},
			}
			break
		}
	}

	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	comp := model.FindComponent(findComponentIDByName(model, "App"))
	require.NotNil(t, comp, "should find App component")
	assert.True(t, comp.IsEntrypoint, "App should be marked as entrypoint")
	assert.Equal(t, "main", comp.EntrypointKind)
}

func TestRelationshipLevels_SystemLevel(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if r.Description == "depends on" {
			assert.Equal(t, "system", r.Level, "dependency relationships should have system level")
		}
	}
}

func TestRelationshipLevels_ContainerLevel(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if r.Description == "imports" {
			assert.Equal(t, "container", r.Level, "import relationships should have container level")
		}
	}
}

func TestRelationshipLevels_ComponentLevel(t *testing.T) {
	builder := c4model.NewModelBuilder()
	model, err := builder.BuildFromAnalysis(enhancedAnalysisResult())
	require.NoError(t, err)

	for _, r := range model.Relationships {
		if r.Description == "implements" {
			assert.Equal(t, "component", r.Level, "implements relationships should have component level")
		}
		// "calls" at component level comes from buildComponentRelationships; "calls" at system level comes from buildExternalSystems
		if r.Description == "calls" && r.Level == "component" {
			assert.Equal(t, "component", r.Level)
		}
	}

	// Verify there is at least one component-level "calls" relationship
	var found bool
	for _, r := range model.Relationships {
		if r.Description == "calls" && r.Level == "component" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have at least one component-level calls relationship")
}

func TestBuildExternalSystems_NoInteractions(t *testing.T) {
	builder := c4model.NewModelBuilder()
	result := sampleAnalysisResult() // no ExternalInteractions
	model, err := builder.BuildFromAnalysis(result)
	require.NoError(t, err)

	for _, s := range model.Systems {
		if s.External {
			assert.Empty(t, s.SystemKind, "systems from direct deps should not have SystemKind")
		}
	}
}

// findComponentIDByName returns the ID of the first component with the given name.
func findComponentIDByName(model *c4model.C4Model, name string) string {
	for _, c := range model.Components {
		if c.Name == name {
			return c.ID
		}
	}
	return ""
}
