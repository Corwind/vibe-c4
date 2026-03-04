package c4model

import (
	"strings"
	"unicode"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// StructRole describes the architectural role of a struct within a C4 model.
type StructRole string

const (
	RoleSkip       StructRole = "skip"
	RoleController StructRole = "controller"
	RoleRepository StructRole = "repository"
	RoleGateway    StructRole = "gateway"
	RoleService    StructRole = "service"
	RoleModel      StructRole = "model"
)

type classificationContext struct {
	structInfo           analyzer.StructInfo
	pkgPath              string
	entrypoints          []analyzer.Entrypoint
	externalInteractions []analyzer.ExternalInteraction
	callGraph            []analyzer.FunctionCall
	modulePath           string
}

func classifyStructRole(ctx classificationContext) StructRole {
	name := ctx.structInfo.Name

	// Rule 1a: Skip well-known infra names and mocks.
	switch name {
	case "Config", "Options", "Settings", "Middleware", "Logger":
		return RoleSkip
	}
	if strings.HasPrefix(name, "Mock") {
		return RoleSkip
	}

	// Rule 1b: Skip pure data structs with 0 methods and all primitive fields.
	if len(ctx.structInfo.Methods) == 0 && len(ctx.structInfo.Fields) > 0 && allFieldsPrimitive(ctx.structInfo.Fields) {
		return RoleSkip
	}

	// Rule 2: Controller — method appears in HTTP/gRPC entrypoints.
	for _, ep := range ctx.entrypoints {
		if ep.PkgPath != ctx.pkgPath {
			continue
		}
		if ep.Kind != "http_handler" && ep.Kind != "grpc_handler" {
			continue
		}
		prefix := name + "."
		if strings.Contains(ep.FuncName, prefix) {
			return RoleController
		}
	}

	// Rule 3: Repository — DB interaction or DB field.
	for _, ei := range ctx.externalInteractions {
		if ei.PkgPath == ctx.pkgPath && ei.TypeName == name && ei.Kind == analyzer.ExtKindDatabase {
			return RoleRepository
		}
	}
	if hasFieldTypeContaining(ctx.structInfo.Fields, "*sql.DB", "*gorm.DB", "*sqlx.DB") {
		return RoleRepository
	}

	// Rule 4: Gateway — HTTP client / Kafka / gRPC client interaction or *http.Client field.
	for _, ei := range ctx.externalInteractions {
		if ei.PkgPath == ctx.pkgPath && ei.TypeName == name {
			switch ei.Kind {
			case analyzer.ExtKindHTTPClient, analyzer.ExtKindKafkaProducer, analyzer.ExtKindKafkaConsumer, analyzer.ExtKindGRPCClient:
				return RoleGateway
			}
		}
	}
	if hasFieldTypeContaining(ctx.structInfo.Fields, "*http.Client") {
		return RoleGateway
	}

	// Rule 5: Name suffix heuristics.
	if strings.HasSuffix(name, "Handler") || strings.HasSuffix(name, "Controller") {
		return RoleController
	}
	if strings.HasSuffix(name, "Service") || strings.HasSuffix(name, "UseCase") {
		return RoleService
	}
	if strings.HasSuffix(name, "Repo") || strings.HasSuffix(name, "Repository") || strings.HasSuffix(name, "Store") || strings.HasSuffix(name, "DAO") {
		return RoleRepository
	}
	if strings.HasSuffix(name, "Client") || strings.HasSuffix(name, "Gateway") {
		return RoleGateway
	}

	// Rule 6: Service by call graph — calls repo/gateway-like types.
	for _, call := range ctx.callGraph {
		if call.CallerType != name {
			continue
		}
		target := call.CalleeType
		if strings.HasSuffix(target, "Repo") || strings.HasSuffix(target, "Repository") ||
			strings.HasSuffix(target, "Store") || strings.HasSuffix(target, "DAO") ||
			strings.HasSuffix(target, "Client") || strings.HasSuffix(target, "Gateway") {
			return RoleService
		}
	}

	// Rule 7: Model — majority exported fields and fewer than 3 methods.
	if len(ctx.structInfo.Methods) < 3 && len(ctx.structInfo.Fields) > 0 {
		exported := 0
		for _, f := range ctx.structInfo.Fields {
			if len(f.Name) > 0 && unicode.IsUpper(rune(f.Name[0])) {
				exported++
			}
		}
		if exported > len(ctx.structInfo.Fields)/2 {
			return RoleModel
		}
	}

	// Rule 8: Default.
	return RoleService
}

func isPrimitiveType(typeName string) bool {
	t := strings.TrimPrefix(typeName, "*")
	t = strings.TrimPrefix(t, "[]")
	t = strings.TrimPrefix(t, "*")

	switch t {
	case "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "error",
		"time.Time", "time.Duration":
		return true
	}
	return false
}

func allFieldsPrimitive(fields []analyzer.FieldInfo) bool {
	for _, f := range fields {
		if !isPrimitiveType(f.Type) {
			return false
		}
	}
	return true
}

func hasFieldTypeContaining(fields []analyzer.FieldInfo, patterns ...string) bool {
	for _, f := range fields {
		for _, p := range patterns {
			if strings.Contains(f.Type, p) {
				return true
			}
		}
	}
	return false
}
