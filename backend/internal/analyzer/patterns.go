package analyzer

import "strings"

// DetectExternalPattern checks if a function call matches a known external system pattern.
// It takes the resolved full package path, the type name (if method call), and the function name.
// Returns an ExternalInteraction if matched, nil otherwise.
func DetectExternalPattern(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if ei := detectChiRouter(fullPkgPath, typeName, funcName); ei != nil {
		return ei
	}
	if ei := detectNetHTTP(fullPkgPath, typeName, funcName); ei != nil {
		return ei
	}
	if ei := detectDatabase(fullPkgPath, typeName, funcName); ei != nil {
		return ei
	}
	if ei := detectKafka(fullPkgPath, typeName, funcName); ei != nil {
		return ei
	}
	if ei := detectGRPC(fullPkgPath, typeName, funcName); ei != nil {
		return ei
	}
	return nil
}

func detectChiRouter(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if !strings.Contains(fullPkgPath, "go-chi/chi") {
		return nil
	}
	switch funcName {
	case "Get", "Post", "Put", "Delete", "Patch", "Handle", "HandleFunc", "Route", "Group", "Mount":
		return &ExternalInteraction{
			Kind:       ExtKindHTTPHandler,
			PkgPath:    fullPkgPath,
			TypeName:   typeName,
			FuncName:   funcName,
			Technology: "chi",
		}
	}
	return nil
}

func detectNetHTTP(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if fullPkgPath != "net/http" {
		return nil
	}
	switch funcName {
	case "HandleFunc", "Handle", "ListenAndServe", "ListenAndServeTLS":
		return &ExternalInteraction{
			Kind:       ExtKindHTTPHandler,
			PkgPath:    fullPkgPath,
			TypeName:   typeName,
			FuncName:   funcName,
			Technology: "net/http",
		}
	case "Get", "Post", "Head", "PostForm", "Do", "NewRequest":
		return &ExternalInteraction{
			Kind:       ExtKindHTTPClient,
			PkgPath:    fullPkgPath,
			TypeName:   typeName,
			FuncName:   funcName,
			Technology: "net/http",
		}
	}
	return nil
}

func detectDatabase(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if fullPkgPath == "database/sql" {
		switch funcName {
		case "Query", "QueryRow", "QueryContext", "QueryRowContext", "Exec", "ExecContext", "Prepare", "PrepareContext", "Begin", "BeginTx", "Open":
			return &ExternalInteraction{
				Kind:       ExtKindDatabase,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "database/sql",
			}
		}
		return nil
	}
	if strings.Contains(fullPkgPath, "gorm.io/gorm") {
		switch funcName {
		case "Create", "Find", "First", "Where", "Save", "Delete", "Updates", "Update", "Raw", "Exec":
			return &ExternalInteraction{
				Kind:       ExtKindDatabase,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "GORM",
			}
		}
		return nil
	}
	if strings.Contains(fullPkgPath, "jmoiron/sqlx") {
		switch funcName {
		case "Query", "QueryRow", "Queryx", "QueryRowx", "Get", "Select", "Exec", "Prepare", "Preparex", "NamedExec", "NamedQuery":
			return &ExternalInteraction{
				Kind:       ExtKindDatabase,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "sqlx",
			}
		}
		return nil
	}
	return nil
}

func detectKafka(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if strings.Contains(fullPkgPath, "Shopify/sarama") {
		switch funcName {
		case "SendMessage", "SendMessages":
			return &ExternalInteraction{
				Kind:       ExtKindKafkaProducer,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "sarama",
			}
		case "ConsumePartition", "ConsumeClaim":
			return &ExternalInteraction{
				Kind:       ExtKindKafkaConsumer,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "sarama",
			}
		}
		return nil
	}
	if strings.Contains(fullPkgPath, "segmentio/kafka-go") {
		switch funcName {
		case "WriteMessages":
			return &ExternalInteraction{
				Kind:       ExtKindKafkaProducer,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "kafka-go",
			}
		case "ReadMessage", "FetchMessage":
			return &ExternalInteraction{
				Kind:       ExtKindKafkaConsumer,
				PkgPath:    fullPkgPath,
				TypeName:   typeName,
				FuncName:   funcName,
				Technology: "kafka-go",
			}
		}
		return nil
	}
	return nil
}

func detectGRPC(fullPkgPath, typeName, funcName string) *ExternalInteraction {
	if fullPkgPath != "google.golang.org/grpc" {
		return nil
	}
	switch funcName {
	case "NewServer", "RegisterService":
		return &ExternalInteraction{
			Kind:       ExtKindGRPCServer,
			PkgPath:    fullPkgPath,
			TypeName:   typeName,
			FuncName:   funcName,
			Technology: "gRPC",
		}
	case "Dial", "DialContext", "NewClient":
		return &ExternalInteraction{
			Kind:       ExtKindGRPCClient,
			PkgPath:    fullPkgPath,
			TypeName:   typeName,
			FuncName:   funcName,
			Technology: "gRPC",
		}
	}
	return nil
}

// technologyFromPkg returns a human-readable technology name from a package path.
func technologyFromPkg(fullPkgPath string) string {
	switch {
	case strings.Contains(fullPkgPath, "go-chi/chi"):
		return "chi"
	case fullPkgPath == "net/http":
		return "net/http"
	case fullPkgPath == "database/sql":
		return "database/sql"
	case strings.Contains(fullPkgPath, "gorm.io/gorm"):
		return "GORM"
	case strings.Contains(fullPkgPath, "jmoiron/sqlx"):
		return "sqlx"
	case strings.Contains(fullPkgPath, "Shopify/sarama"):
		return "sarama"
	case strings.Contains(fullPkgPath, "segmentio/kafka-go"):
		return "kafka-go"
	case fullPkgPath == "google.golang.org/grpc":
		return "gRPC"
	default:
		return fullPkgPath
	}
}
