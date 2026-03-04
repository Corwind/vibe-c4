package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectExternalPattern(t *testing.T) {
	tests := []struct {
		name       string
		pkgPath    string
		typeName   string
		funcName   string
		wantNil    bool
		wantKind   ExternalDependencyKind
		wantTech   string
	}{
		{
			name:     "chi router Get",
			pkgPath:  "github.com/go-chi/chi/v5",
			funcName: "Get",
			wantKind: ExtKindHTTPHandler,
			wantTech: "chi",
		},
		{
			name:     "chi router Post",
			pkgPath:  "github.com/go-chi/chi/v5",
			funcName: "Post",
			wantKind: ExtKindHTTPHandler,
			wantTech: "chi",
		},
		{
			name:     "chi router Mount",
			pkgPath:  "github.com/go-chi/chi/v5",
			funcName: "Mount",
			wantKind: ExtKindHTTPHandler,
			wantTech: "chi",
		},
		{
			name:     "net/http HandleFunc",
			pkgPath:  "net/http",
			funcName: "HandleFunc",
			wantKind: ExtKindHTTPHandler,
			wantTech: "net/http",
		},
		{
			name:     "net/http ListenAndServe",
			pkgPath:  "net/http",
			funcName: "ListenAndServe",
			wantKind: ExtKindHTTPHandler,
			wantTech: "net/http",
		},
		{
			name:     "net/http client Get",
			pkgPath:  "net/http",
			funcName: "Get",
			wantKind: ExtKindHTTPClient,
			wantTech: "net/http",
		},
		{
			name:     "net/http client Do",
			pkgPath:  "net/http",
			funcName: "Do",
			wantKind: ExtKindHTTPClient,
			wantTech: "net/http",
		},
		{
			name:     "database/sql Query",
			pkgPath:  "database/sql",
			funcName: "Query",
			wantKind: ExtKindDatabase,
			wantTech: "database/sql",
		},
		{
			name:     "database/sql Exec",
			pkgPath:  "database/sql",
			funcName: "Exec",
			wantKind: ExtKindDatabase,
			wantTech: "database/sql",
		},
		{
			name:     "GORM Create",
			pkgPath:  "gorm.io/gorm",
			funcName: "Create",
			wantKind: ExtKindDatabase,
			wantTech: "GORM",
		},
		{
			name:     "GORM Find",
			pkgPath:  "gorm.io/gorm",
			funcName: "Find",
			wantKind: ExtKindDatabase,
			wantTech: "GORM",
		},
		{
			name:     "sqlx Select",
			pkgPath:  "github.com/jmoiron/sqlx",
			funcName: "Select",
			wantKind: ExtKindDatabase,
			wantTech: "sqlx",
		},
		{
			name:     "sarama producer SendMessage",
			pkgPath:  "github.com/Shopify/sarama",
			funcName: "SendMessage",
			wantKind: ExtKindKafkaProducer,
			wantTech: "sarama",
		},
		{
			name:     "sarama consumer ConsumePartition",
			pkgPath:  "github.com/Shopify/sarama",
			funcName: "ConsumePartition",
			wantKind: ExtKindKafkaConsumer,
			wantTech: "sarama",
		},
		{
			name:     "kafka-go producer WriteMessages",
			pkgPath:  "github.com/segmentio/kafka-go",
			funcName: "WriteMessages",
			wantKind: ExtKindKafkaProducer,
			wantTech: "kafka-go",
		},
		{
			name:     "kafka-go consumer ReadMessage",
			pkgPath:  "github.com/segmentio/kafka-go",
			funcName: "ReadMessage",
			wantKind: ExtKindKafkaConsumer,
			wantTech: "kafka-go",
		},
		{
			name:     "grpc server NewServer",
			pkgPath:  "google.golang.org/grpc",
			funcName: "NewServer",
			wantKind: ExtKindGRPCServer,
			wantTech: "gRPC",
		},
		{
			name:     "grpc server RegisterService",
			pkgPath:  "google.golang.org/grpc",
			funcName: "RegisterService",
			wantKind: ExtKindGRPCServer,
			wantTech: "gRPC",
		},
		{
			name:     "grpc client Dial",
			pkgPath:  "google.golang.org/grpc",
			funcName: "Dial",
			wantKind: ExtKindGRPCClient,
			wantTech: "gRPC",
		},
		{
			name:     "grpc client NewClient",
			pkgPath:  "google.golang.org/grpc",
			funcName: "NewClient",
			wantKind: ExtKindGRPCClient,
			wantTech: "gRPC",
		},
		{
			name:    "unknown package returns nil",
			pkgPath: "github.com/unknown/pkg",
			funcName: "DoSomething",
			wantNil: true,
		},
		{
			name:    "known package but unknown func returns nil",
			pkgPath: "net/http",
			funcName: "UnknownFunc",
			wantNil: true,
		},
		{
			name:    "chi with unknown method returns nil",
			pkgPath: "github.com/go-chi/chi/v5",
			funcName: "UnknownMethod",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectExternalPattern(tt.pkgPath, tt.typeName, tt.funcName)
			if tt.wantNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result, "expected non-nil ExternalInteraction")
			assert.Equal(t, tt.wantKind, result.Kind)
			assert.Equal(t, tt.wantTech, result.Technology)
			assert.Equal(t, tt.funcName, result.FuncName)
			assert.Equal(t, tt.pkgPath, result.PkgPath)
		})
	}
}

func TestTechnologyFromPkg(t *testing.T) {
	tests := []struct {
		pkgPath  string
		wantTech string
	}{
		{"github.com/go-chi/chi/v5", "chi"},
		{"net/http", "net/http"},
		{"database/sql", "database/sql"},
		{"gorm.io/gorm", "GORM"},
		{"github.com/jmoiron/sqlx", "sqlx"},
		{"github.com/Shopify/sarama", "sarama"},
		{"github.com/segmentio/kafka-go", "kafka-go"},
		{"google.golang.org/grpc", "gRPC"},
		{"github.com/unknown/pkg", "github.com/unknown/pkg"},
		{"fmt", "fmt"},
	}

	for _, tt := range tests {
		t.Run(tt.pkgPath, func(t *testing.T) {
			assert.Equal(t, tt.wantTech, technologyFromPkg(tt.pkgPath))
		})
	}
}
