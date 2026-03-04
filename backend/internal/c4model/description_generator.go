package c4model

import (
	"strings"
	"unicode"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

func generateSystemDescription(entrypoints []analyzer.Entrypoint, interactions []analyzer.ExternalInteraction) string {
	seen := make(map[string]bool)
	var caps []string

	add := func(cap string) {
		if !seen[cap] {
			seen[cap] = true
			caps = append(caps, cap)
		}
	}

	for _, ep := range entrypoints {
		switch ep.Kind {
		case "http_handler":
			add("serves HTTP requests")
		case "grpc_handler", "grpc_server":
			add("provides gRPC services")
		}
	}

	for _, ei := range interactions {
		switch ei.Kind {
		case analyzer.ExtKindDatabase:
			tech := ei.Technology
			if tech == "" {
				tech = "a database"
			}
			add("stores data in " + tech)
		case analyzer.ExtKindKafkaProducer, analyzer.ExtKindKafkaConsumer:
			add("uses Kafka messaging")
		case analyzer.ExtKindHTTPClient:
			add("calls external HTTP APIs")
		case analyzer.ExtKindGRPCClient:
			add("calls external gRPC services")
		}
	}

	if len(caps) == 0 {
		return "Go application"
	}
	return "Go application that " + strings.Join(caps, ", ")
}

func generateContainerDescription(group ContainerGroup) string {
	if group.IsInfra {
		switch group.SystemKind {
		case "database":
			if group.Technology != "" {
				return group.Technology + " database"
			}
			return "database"
		case "message_queue":
			if group.Technology != "" {
				return group.Technology + " message broker"
			}
			return "message broker"
		case "http_api":
			return "External HTTP API"
		case "grpc_service":
			return "External gRPC service"
		}
		return "external system"
	}
	return "Go service"
}

func generateComponentDescription(name string, role StructRole) string {
	domain := extractDomain(name)

	switch role {
	case RoleController:
		return "Handles " + domain + " requests"
	case RoleService:
		return "Provides " + domain + " business logic"
	case RoleRepository:
		return "Provides data access for " + domain + " records"
	case RoleGateway:
		return "Communicates with external " + domain + " services"
	case RoleModel:
		return domain + " data model"
	default:
		return name + " component"
	}
}

func extractDomain(name string) string {
	suffixes := []string{
		"Handler", "Controller", "Service", "UseCase",
		"Repository", "Repo", "Store", "DAO",
		"Client", "Gateway",
	}

	stripped := name
	for _, s := range suffixes {
		if strings.HasSuffix(stripped, s) {
			stripped = strings.TrimSuffix(stripped, s)
			break
		}
	}

	if stripped == "" {
		stripped = name
	}

	// Insert spaces before uppercase letters for camelCase splitting.
	var result strings.Builder
	for i, r := range stripped {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune(' ')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func generateRelationshipDescription(kind analyzer.ExternalDependencyKind) string {
	switch kind {
	case analyzer.ExtKindDatabase:
		return "Reads and writes data"
	case analyzer.ExtKindKafkaProducer:
		return "Publishes events"
	case analyzer.ExtKindKafkaConsumer:
		return "Consumes events"
	case analyzer.ExtKindHTTPClient:
		return "Calls HTTP API"
	case analyzer.ExtKindGRPCClient:
		return "Calls gRPC service"
	default:
		return "uses"
	}
}
