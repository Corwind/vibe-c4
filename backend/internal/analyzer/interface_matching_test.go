package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchInterfaceImpls_FullMatch(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project/service",
			Interfaces: []InterfaceInfo{
				{
					Name: "Worker",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
						{Name: "Stop", Returns: []string{"error"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "Service",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
						{Name: "Stop", Returns: []string{"error"}},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	require.Len(t, impls, 1)
	assert.Equal(t, "Service", impls[0].StructName)
	assert.Equal(t, "Worker", impls[0].InterfaceName)
	assert.Equal(t, "example.com/project/service", impls[0].StructPkg)
	assert.Equal(t, "example.com/project/service", impls[0].InterfacePkg)
}

func TestMatchInterfaceImpls_PartialMatch_DoesNotImplement(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project/service",
			Interfaces: []InterfaceInfo{
				{
					Name: "Worker",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
						{Name: "Stop", Returns: []string{"error"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "PartialService",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
						// Missing Stop() method
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	assert.Empty(t, impls)
}

func TestMatchInterfaceImpls_SupersetMatch(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project/service",
			Interfaces: []InterfaceInfo{
				{
					Name: "Worker",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "FullService",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
						{Name: "Stop", Returns: []string{"error"}},
						{Name: "Reset"},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	require.Len(t, impls, 1)
	assert.Equal(t, "FullService", impls[0].StructName)
	assert.Equal(t, "Worker", impls[0].InterfaceName)
}

func TestMatchInterfaceImpls_ParamTypeMismatch(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project",
			Interfaces: []InterfaceInfo{
				{
					Name: "Processor",
					Methods: []MethodInfo{
						{Name: "Process", Params: []string{"string"}, Returns: []string{"error"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "MyProcessor",
					Methods: []MethodInfo{
						{Name: "Process", Params: []string{"int"}, Returns: []string{"error"}},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	assert.Empty(t, impls, "should not match when param types differ")
}

func TestMatchInterfaceImpls_ReturnTypeMismatch(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project",
			Interfaces: []InterfaceInfo{
				{
					Name: "Processor",
					Methods: []MethodInfo{
						{Name: "Process", Returns: []string{"error"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "MyProcessor",
					Methods: []MethodInfo{
						{Name: "Process", Returns: []string{"string"}},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	assert.Empty(t, impls, "should not match when return types differ")
}

func TestMatchInterfaceImpls_CrossPackage(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project/api",
			Interfaces: []InterfaceInfo{
				{
					Name: "Repository",
					Methods: []MethodInfo{
						{Name: "Get", Params: []string{"string"}, Returns: []string{"*Entity", "error"}},
						{Name: "Save", Params: []string{"*Entity"}, Returns: []string{"error"}},
					},
				},
			},
		},
		{
			ImportPath: "example.com/project/storage",
			Structs: []StructInfo{
				{
					Name: "PostgresRepo",
					Methods: []MethodInfo{
						{Name: "Get", Params: []string{"string"}, Returns: []string{"*Entity", "error"}},
						{Name: "Save", Params: []string{"*Entity"}, Returns: []string{"error"}},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	require.Len(t, impls, 1)
	assert.Equal(t, "example.com/project/storage", impls[0].StructPkg)
	assert.Equal(t, "PostgresRepo", impls[0].StructName)
	assert.Equal(t, "example.com/project/api", impls[0].InterfacePkg)
	assert.Equal(t, "Repository", impls[0].InterfaceName)
}

func TestMatchInterfaceImpls_NoInterfaces(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project",
			Structs: []StructInfo{
				{
					Name: "Service",
					Methods: []MethodInfo{
						{Name: "DoWork"},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	assert.Empty(t, impls)
}

func TestMatchInterfaceImpls_NoStructs(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project",
			Interfaces: []InterfaceInfo{
				{
					Name: "Worker",
					Methods: []MethodInfo{
						{Name: "DoWork"},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	assert.Empty(t, impls)
}

func TestMatchInterfaceImpls_MultipleStructsImplementSameInterface(t *testing.T) {
	packages := []PackageInfo{
		{
			ImportPath: "example.com/project",
			Interfaces: []InterfaceInfo{
				{
					Name: "Worker",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
					},
				},
			},
			Structs: []StructInfo{
				{
					Name: "ServiceA",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
					},
				},
				{
					Name: "ServiceB",
					Methods: []MethodInfo{
						{Name: "DoWork", Returns: []string{"string"}},
					},
				},
			},
		},
	}

	impls := matchInterfaceImpls(packages)

	require.Len(t, impls, 2)
	names := []string{impls[0].StructName, impls[1].StructName}
	assert.Contains(t, names, "ServiceA")
	assert.Contains(t, names, "ServiceB")
}
