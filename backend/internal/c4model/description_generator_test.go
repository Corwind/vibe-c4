package c4model

import (
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSystemDescription_HTTPOnly(t *testing.T) {
	entrypoints := []analyzer.Entrypoint{
		{Kind: "http_handler", FuncName: "HandleGet"},
	}
	desc := generateSystemDescription(entrypoints, nil)
	assert.Contains(t, desc, "serves HTTP requests")
}

func TestGenerateSystemDescription_WithDB(t *testing.T) {
	interactions := []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL"},
	}
	desc := generateSystemDescription(nil, interactions)
	assert.Contains(t, desc, "stores data in PostgreSQL")
}

func TestGenerateSystemDescription_Multiple(t *testing.T) {
	entrypoints := []analyzer.Entrypoint{
		{Kind: "http_handler", FuncName: "HandleGet"},
		{Kind: "grpc_handler", FuncName: "ServeGRPC"},
	}
	interactions := []analyzer.ExternalInteraction{
		{Kind: analyzer.ExtKindDatabase, Technology: "PostgreSQL"},
		{Kind: analyzer.ExtKindKafkaProducer},
	}
	desc := generateSystemDescription(entrypoints, interactions)
	assert.Contains(t, desc, "serves HTTP requests")
	assert.Contains(t, desc, "provides gRPC services")
	assert.Contains(t, desc, "stores data in PostgreSQL")
	assert.Contains(t, desc, "uses Kafka messaging")
}

func TestGenerateSystemDescription_Empty(t *testing.T) {
	desc := generateSystemDescription(nil, nil)
	assert.Equal(t, "Go application", desc)
}

func TestGenerateContainerDescription_Database(t *testing.T) {
	group := ContainerGroup{
		IsInfra:    true,
		SystemKind: "database",
		Technology: "PostgreSQL",
	}
	desc := generateContainerDescription(group)
	assert.Equal(t, "PostgreSQL database", desc)
}

func TestGenerateContainerDescription_MessageQueue(t *testing.T) {
	group := ContainerGroup{
		IsInfra:    true,
		SystemKind: "message_queue",
		Technology: "Kafka",
	}
	desc := generateContainerDescription(group)
	assert.Equal(t, "Kafka message broker", desc)
}

func TestGenerateContainerDescription_CodeContainer(t *testing.T) {
	group := ContainerGroup{
		IsInfra: false,
	}
	desc := generateContainerDescription(group)
	assert.Equal(t, "Go service", desc)
}

func TestGenerateComponentDescription_Controller(t *testing.T) {
	desc := generateComponentDescription("UserHandler", RoleController)
	assert.Equal(t, "Handles user requests", desc)
}

func TestGenerateComponentDescription_Service(t *testing.T) {
	desc := generateComponentDescription("UserService", RoleService)
	assert.Equal(t, "Provides user business logic", desc)
}

func TestGenerateComponentDescription_Repository(t *testing.T) {
	desc := generateComponentDescription("UserRepository", RoleRepository)
	assert.Equal(t, "Provides data access for user records", desc)
}

func TestGenerateComponentDescription_Gateway(t *testing.T) {
	desc := generateComponentDescription("PaymentGateway", RoleGateway)
	assert.Equal(t, "Communicates with external payment services", desc)
}

func TestGenerateComponentDescription_Model(t *testing.T) {
	desc := generateComponentDescription("User", RoleModel)
	assert.Equal(t, "user data model", desc)
}

func TestGenerateRelationshipDescription_DB(t *testing.T) {
	desc := generateRelationshipDescription(analyzer.ExtKindDatabase)
	assert.Equal(t, "Reads and writes data", desc)
}

func TestGenerateRelationshipDescription_Kafka(t *testing.T) {
	desc := generateRelationshipDescription(analyzer.ExtKindKafkaProducer)
	assert.Equal(t, "Publishes events", desc)
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserHandler", "user"},
		{"UserService", "user"},
		{"UserRepository", "user"},
		{"PaymentGateway", "payment"},
		{"OrderClient", "order"},
		{"UserOrderService", "user order"},
		{"Handler", "handler"},
		{"User", "user"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractDomain(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
