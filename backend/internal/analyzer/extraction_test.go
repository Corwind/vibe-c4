package analyzer_test

import (
	"context"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findPackage(packages []analyzer.PackageInfo, name string) *analyzer.PackageInfo {
	for i := range packages {
		if packages[i].Name == name {
			return &packages[i]
		}
	}
	return nil
}

func findStruct(structs []analyzer.StructInfo, name string) *analyzer.StructInfo {
	for i := range structs {
		if structs[i].Name == name {
			return &structs[i]
		}
	}
	return nil
}

func findInterface(interfaces []analyzer.InterfaceInfo, name string) *analyzer.InterfaceInfo {
	for i := range interfaces {
		if interfaces[i].Name == name {
			return &interfaces[i]
		}
	}
	return nil
}

func findFunction(functions []analyzer.FunctionInfo, name string) *analyzer.FunctionInfo {
	for i := range functions {
		if functions[i].Name == name {
			return &functions[i]
		}
	}
	return nil
}

func findMethod(methods []analyzer.MethodInfo, name string) *analyzer.MethodInfo {
	for i := range methods {
		if methods[i].Name == name {
			return &methods[i]
		}
	}
	return nil
}

func TestExtraction_StructsDiscovered(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	require.Len(t, svcPkg.Structs, 2, "should find BaseService and Service structs")

	baseService := findStruct(svcPkg.Structs, "BaseService")
	require.NotNil(t, baseService, "should find BaseService")
	assert.NotEmpty(t, baseService.Fields)

	service := findStruct(svcPkg.Structs, "Service")
	require.NotNil(t, service, "should find Service")
}

func TestExtraction_StructFields(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	baseService := findStruct(svcPkg.Structs, "BaseService")
	require.NotNil(t, baseService)

	assert.Len(t, baseService.Fields, 1)
	assert.Equal(t, "Name", baseService.Fields[0].Name)
	assert.Equal(t, "string", baseService.Fields[0].Type)
}

func TestExtraction_EmbeddedStruct(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	service := findStruct(svcPkg.Structs, "Service")
	require.NotNil(t, service)

	var hasEmbedded bool
	for _, f := range service.Fields {
		if f.Embedded && f.Type == "BaseService" {
			hasEmbedded = true
			break
		}
	}
	assert.True(t, hasEmbedded, "Service should have an embedded BaseService")
}

func TestExtraction_StructMethods(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	service := findStruct(svcPkg.Structs, "Service")
	require.NotNil(t, service)
	require.GreaterOrEqual(t, len(service.Methods), 2, "Service should have at least DoWork and Stop methods")

	doWork := findMethod(service.Methods, "DoWork")
	require.NotNil(t, doWork)
	assert.Equal(t, []string{"string"}, doWork.Returns)

	stop := findMethod(service.Methods, "Stop")
	require.NotNil(t, stop)
	assert.Equal(t, []string{"error"}, stop.Returns)
}

func TestExtraction_InterfacesDiscovered(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	require.Len(t, svcPkg.Interfaces, 1, "should find Worker interface")

	worker := findInterface(svcPkg.Interfaces, "Worker")
	require.NotNil(t, worker)
}

func TestExtraction_InterfaceMethods(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	worker := findInterface(svcPkg.Interfaces, "Worker")
	require.NotNil(t, worker)
	require.Len(t, worker.Methods, 2, "Worker should have DoWork and Stop methods")

	doWork := findMethod(worker.Methods, "DoWork")
	require.NotNil(t, doWork)
	assert.Equal(t, []string{"string"}, doWork.Returns)

	stop := findMethod(worker.Methods, "Stop")
	require.NotNil(t, stop)
	assert.Equal(t, []string{"error"}, stop.Returns)
}

func TestExtraction_StandaloneFunctions(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	// Should find New and HelperFunc (standalone functions, not methods)
	newFn := findFunction(svcPkg.Functions, "New")
	require.NotNil(t, newFn, "should find New function")
	assert.Equal(t, []string{"*Service"}, newFn.Returns)

	helperFn := findFunction(svcPkg.Functions, "HelperFunc")
	require.NotNil(t, helperFn, "should find HelperFunc function")
	assert.Equal(t, []string{"string"}, helperFn.Params)
	assert.Equal(t, []string{"string"}, helperFn.Returns)
}

func TestExtraction_HandlerPackageStruct(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	handlerPkg := findPackage(result.Packages, "handler")
	require.NotNil(t, handlerPkg)

	require.Len(t, handlerPkg.Structs, 1)
	h := findStruct(handlerPkg.Structs, "Handler")
	require.NotNil(t, h)

	// Handler should have a method ServeHTTP
	serveHTTP := findMethod(h.Methods, "ServeHTTP")
	require.NotNil(t, serveHTTP)
}

func TestExtraction_FilePathsPopulated(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	svcPkg := findPackage(result.Packages, "service")
	require.NotNil(t, svcPkg)

	for _, s := range svcPkg.Structs {
		assert.NotEmpty(t, s.FilePath, "struct %s should have a file path", s.Name)
		assert.Greater(t, s.Line, 0, "struct %s should have a line number", s.Name)
	}

	for _, iface := range svcPkg.Interfaces {
		assert.NotEmpty(t, iface.FilePath, "interface %s should have a file path", iface.Name)
		assert.Greater(t, iface.Line, 0, "interface %s should have a line number", iface.Name)
	}

	for _, fn := range svcPkg.Functions {
		assert.NotEmpty(t, fn.FilePath, "function %s should have a file path", fn.Name)
		assert.Greater(t, fn.Line, 0, "function %s should have a line number", fn.Name)
	}
}

func TestExtraction_UtilsPackageFunctions(t *testing.T) {
	a := analyzer.NewGoAnalyzer()
	result, err := a.AnalyzeProject(context.Background(), testdataPath(t))
	require.NoError(t, err)

	utilsPkg := findPackage(result.Packages, "utils")
	require.NotNil(t, utilsPkg)

	logFn := findFunction(utilsPkg.Functions, "Log")
	require.NotNil(t, logFn, "should find Log function")
	assert.Equal(t, []string{"string"}, logFn.Params)
}
