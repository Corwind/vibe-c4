package c4model

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
)

func TestClassifyStructRole_SkipConfig(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{Name: "Config"},
	}
	assert.Equal(t, RoleSkip, classifyStructRole(ctx))
}

func TestClassifyStructRole_SkipMock(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{Name: "MockUserService"},
	}
	assert.Equal(t, RoleSkip, classifyStructRole(ctx))
}

func TestClassifyStructRole_SkipPrimitiveOnly(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "Pagination",
			Fields: []analyzer.FieldInfo{
				{Name: "Page", Type: "int"},
				{Name: "Size", Type: "int"},
				{Name: "Query", Type: "string"},
			},
			// No methods
		},
	}
	assert.Equal(t, RoleSkip, classifyStructRole(ctx))
}

func TestClassifyStructRole_Controller(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "UserHandler",
			Methods: []analyzer.MethodInfo{
				{Name: "ServeHTTP"},
			},
		},
		pkgPath: "myapp/internal/handler",
		entrypoints: []analyzer.Entrypoint{
			{
				Kind:     "http_handler",
				PkgPath:  "myapp/internal/handler",
				FuncName: "UserHandler.ServeHTTP",
			},
		},
	}
	assert.Equal(t, RoleController, classifyStructRole(ctx))
}

func TestClassifyStructRole_Repository_DBField(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "UserStore",
			Fields: []analyzer.FieldInfo{
				{Name: "db", Type: "*sql.DB"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "FindByID"},
			},
		},
		pkgPath: "myapp/internal/store",
	}
	// Note: name suffix "Store" would also match, but DB field check fires first via rule 3.
	assert.Equal(t, RoleRepository, classifyStructRole(ctx))
}

func TestClassifyStructRole_Repository_DBInteraction(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "AccountPersistence",
			Fields: []analyzer.FieldInfo{
				{Name: "conn", Type: "SomeConnWrapper"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "Save"},
			},
		},
		pkgPath: "myapp/internal/persistence",
		externalInteractions: []analyzer.ExternalInteraction{
			{
				Kind:     analyzer.ExtKindDatabase,
				PkgPath:  "myapp/internal/persistence",
				TypeName: "AccountPersistence",
				FuncName: "Save",
			},
		},
	}
	assert.Equal(t, RoleRepository, classifyStructRole(ctx))
}

func TestClassifyStructRole_Gateway_HTTPClientField(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "NotificationSender",
			Fields: []analyzer.FieldInfo{
				{Name: "client", Type: "*http.Client"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "Send"},
			},
		},
		pkgPath: "myapp/internal/notify",
	}
	assert.Equal(t, RoleGateway, classifyStructRole(ctx))
}

func TestClassifyStructRole_Gateway_Interaction(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "PaymentAdapter",
			Fields: []analyzer.FieldInfo{
				{Name: "baseURL", Type: "string"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "Charge"},
			},
		},
		pkgPath: "myapp/internal/payment",
		externalInteractions: []analyzer.ExternalInteraction{
			{
				Kind:     analyzer.ExtKindHTTPClient,
				PkgPath:  "myapp/internal/payment",
				TypeName: "PaymentAdapter",
				FuncName: "Charge",
			},
		},
	}
	assert.Equal(t, RoleGateway, classifyStructRole(ctx))
}

func TestClassifyStructRole_NameSuffix_Handler(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "OrderHandler",
			Methods: []analyzer.MethodInfo{
				{Name: "Create"},
			},
		},
	}
	assert.Equal(t, RoleController, classifyStructRole(ctx))
}

func TestClassifyStructRole_NameSuffix_Service(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "OrderService",
			Methods: []analyzer.MethodInfo{
				{Name: "PlaceOrder"},
			},
		},
	}
	assert.Equal(t, RoleService, classifyStructRole(ctx))
}

func TestClassifyStructRole_NameSuffix_Repo(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "OrderRepo",
			Methods: []analyzer.MethodInfo{
				{Name: "FindAll"},
			},
		},
	}
	assert.Equal(t, RoleRepository, classifyStructRole(ctx))
}

func TestClassifyStructRole_NameSuffix_Client(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "StripeClient",
			Methods: []analyzer.MethodInfo{
				{Name: "Charge"},
			},
		},
	}
	assert.Equal(t, RoleGateway, classifyStructRole(ctx))
}

func TestClassifyStructRole_ServiceByCallGraph(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "Orchestrator",
			Methods: []analyzer.MethodInfo{
				{Name: "Run"},
			},
		},
		callGraph: []analyzer.FunctionCall{
			{
				CallerType: "Orchestrator",
				CallerFunc: "Run",
				CalleeType: "UserRepo",
				CalleeFunc: "FindByID",
			},
		},
	}
	assert.Equal(t, RoleService, classifyStructRole(ctx))
}

func TestClassifyStructRole_Model(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "User",
			Fields: []analyzer.FieldInfo{
				{Name: "ID", Type: "int"},
				{Name: "Name", Type: "string"},
				{Name: "Email", Type: "string"},
				{Name: "active", Type: "bool"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "FullName"},
			},
		},
	}
	assert.Equal(t, RoleModel, classifyStructRole(ctx))
}

func TestClassifyStructRole_DefaultService(t *testing.T) {
	ctx := classificationContext{
		structInfo: analyzer.StructInfo{
			Name: "Processor",
			Fields: []analyzer.FieldInfo{
				{Name: "logger", Type: "*zap.Logger"},
			},
			Methods: []analyzer.MethodInfo{
				{Name: "Process"},
				{Name: "Validate"},
				{Name: "Transform"},
			},
		},
	}
	assert.Equal(t, RoleService, classifyStructRole(ctx))
}
