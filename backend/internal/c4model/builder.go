package c4model

import (
	"fmt"
	"path"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// ModelBuilder defines the interface for building C4 models from analysis results.
type ModelBuilder interface {
	BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error)
}

// DefaultModelBuilder implements ModelBuilder.
type DefaultModelBuilder struct{}

// NewModelBuilder creates a new DefaultModelBuilder.
func NewModelBuilder() *DefaultModelBuilder {
	return &DefaultModelBuilder{}
}

// BuildFromAnalysis transforms analyzer output into a C4 model hierarchy.
func (b *DefaultModelBuilder) BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}

	model := &C4Model{}

	b.buildSystems(result, model)
	b.buildContainers(result, model)
	b.buildComponents(result, model)
	b.buildCodeElements(result, model)
	b.buildRelationships(result, model)
	b.buildExternalSystems(result, model)
	b.buildComponentRelationships(result, model)
	b.buildInterfaceImplRelationships(result, model)
	b.markEntrypoints(result, model)

	return model, nil
}

func (b *DefaultModelBuilder) buildSystems(result *analyzer.AnalysisResult, model *C4Model) {
	// Main system from module
	mainSystemID := sanitizeID("system", result.Module.ModulePath)
	moduleName := path.Base(result.Module.ModulePath)

	model.Systems = append(model.Systems, System{
		ID:         mainSystemID,
		Name:       moduleName,
		ModulePath: result.Module.ModulePath,
		External:   false,
	})

	// External systems from direct dependencies
	for _, dep := range result.Module.DirectDeps {
		model.Systems = append(model.Systems, System{
			ID:         sanitizeID("system", dep.Path),
			Name:       shortModuleName(dep.Path),
			ModulePath: dep.Path,
			External:   true,
		})
	}
}

func (b *DefaultModelBuilder) buildContainers(result *analyzer.AnalysisResult, model *C4Model) {
	mainSystemID := sanitizeID("system", result.Module.ModulePath)

	for _, pkg := range result.Packages {
		containerID := sanitizeID("container", pkg.ImportPath)

		container := Container{
			ID:          containerID,
			Name:        pkg.Dir,
			Technology:  "Go",
			PackagePath: pkg.ImportPath,
			SystemID:    mainSystemID,
		}

		model.Containers = append(model.Containers, container)
	}

	// Update main system's container IDs
	for i := range model.Systems {
		if model.Systems[i].ID == mainSystemID {
			for _, c := range model.Containers {
				model.Systems[i].ContainerIDs = append(model.Systems[i].ContainerIDs, c.ID)
			}
			break
		}
	}
}

func (b *DefaultModelBuilder) buildComponents(result *analyzer.AnalysisResult, model *C4Model) {
	for _, pkg := range result.Packages {
		containerID := sanitizeID("container", pkg.ImportPath)

		// Structs become components
		for _, s := range pkg.Structs {
			compID := sanitizeID("component", pkg.ImportPath+"/"+s.Name)
			model.Components = append(model.Components, Component{
				ID:          compID,
				Name:        s.Name,
				Type:        ComponentTypeStruct,
				Technology:  "Go struct",
				ContainerID: containerID,
			})

			// Update container's component IDs
			b.addComponentToContainer(model, containerID, compID)
		}

		// Interfaces become components
		for _, iface := range pkg.Interfaces {
			compID := sanitizeID("component", pkg.ImportPath+"/"+iface.Name)
			model.Components = append(model.Components, Component{
				ID:          compID,
				Name:        iface.Name,
				Type:        ComponentTypeInterface,
				Technology:  "Go interface",
				ContainerID: containerID,
			})
			b.addComponentToContainer(model, containerID, compID)
		}
	}
}

func (b *DefaultModelBuilder) buildCodeElements(result *analyzer.AnalysisResult, model *C4Model) {
	for _, pkg := range result.Packages {
		// Methods on structs
		for _, s := range pkg.Structs {
			compID := sanitizeID("component", pkg.ImportPath+"/"+s.Name)
			for _, m := range s.Methods {
				ceID := sanitizeID("code", pkg.ImportPath+"/"+s.Name+"."+m.Name)
				model.CodeElements = append(model.CodeElements, CodeElement{
					ID:          ceID,
					Name:        m.Name,
					Type:        CodeElementTypeMethod,
					ComponentID: compID,
					FilePath:    m.FilePath,
					Line:        m.Line,
				})
			}

			// Fields as code elements
			for _, f := range s.Fields {
				if f.Embedded {
					continue
				}
				ceID := sanitizeID("code", pkg.ImportPath+"/"+s.Name+"."+f.Name)
				model.CodeElements = append(model.CodeElements, CodeElement{
					ID:          ceID,
					Name:        f.Name,
					Type:        CodeElementTypeField,
					ComponentID: compID,
				})
			}
		}

		// Standalone functions as code elements in their container
		for _, fn := range pkg.Functions {
			containerID := sanitizeID("container", pkg.ImportPath)
			ceID := sanitizeID("code", pkg.ImportPath+"/"+fn.Name)
			model.CodeElements = append(model.CodeElements, CodeElement{
				ID:          ceID,
				Name:        fn.Name,
				Type:        CodeElementTypeFunction,
				ComponentID: containerID,
				FilePath:    fn.FilePath,
				Line:        fn.Line,
			})
		}
	}
}

func (b *DefaultModelBuilder) buildRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	// Container-level relationships from import graph
	for source, targets := range result.ImportGraph {
		sourceID := sanitizeID("container", source)
		for _, target := range targets {
			targetID := sanitizeID("container", target)
			model.Relationships = append(model.Relationships, Relationship{
				SourceID:    sourceID,
				TargetID:    targetID,
				Description: "imports",
				Technology:  "Go import",
				Level:       "container",
			})
		}
	}

	// System-level relationships: main system uses external systems
	mainSystemID := sanitizeID("system", result.Module.ModulePath)
	for _, dep := range result.Module.DirectDeps {
		extSystemID := sanitizeID("system", dep.Path)
		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    mainSystemID,
			TargetID:    extSystemID,
			Description: "depends on",
			Technology:  "Go module",
			Level:       "system",
		})
	}
}

func (b *DefaultModelBuilder) addComponentToContainer(model *C4Model, containerID, componentID string) {
	for i := range model.Containers {
		if model.Containers[i].ID == containerID {
			model.Containers[i].ComponentIDs = append(model.Containers[i].ComponentIDs, componentID)
			return
		}
	}
}

// shortModuleName extracts a readable short name from a Go module path.
// For "github.com/go-chi/chi/v5" it returns "chi/v5".
// For "github.com/lib/pq" it returns "pq".
func shortModuleName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	if len(parts) <= 2 {
		return path.Base(modulePath)
	}
	// Skip domain (github.com) and org, take the rest
	// e.g. github.com/go-chi/chi/v5 -> chi/v5
	return strings.Join(parts[2:], "/")
}

func (b *DefaultModelBuilder) buildExternalSystems(result *analyzer.AnalysisResult, model *C4Model) {
	mainSystemID := sanitizeID("system", result.Module.ModulePath)

	// Group by kind, deduplicating
	seenKinds := make(map[string]bool)
	for _, interaction := range result.ExternalInteractions {
		kind := interaction.Kind

		// Skip inbound handlers — they are not external systems
		if kind == analyzer.ExtKindHTTPHandler || kind == analyzer.ExtKindGRPCServer {
			continue
		}

		// Normalize kafka producer/consumer to single "kafka" key
		kindKey := string(kind)
		if kind == analyzer.ExtKindKafkaProducer || kind == analyzer.ExtKindKafkaConsumer {
			kindKey = "kafka"
		}

		if seenKinds[kindKey] {
			continue
		}
		seenKinds[kindKey] = true

		var name, systemKind, description string
		switch kind {
		case analyzer.ExtKindDatabase:
			name = "Database"
			systemKind = "database"
			description = "reads from / writes to"
		case analyzer.ExtKindKafkaProducer, analyzer.ExtKindKafkaConsumer:
			name = "Message Queue"
			systemKind = "message_queue"
			description = "produces to / consumes from"
		case analyzer.ExtKindHTTPClient:
			name = "External HTTP API"
			systemKind = "http_api"
			description = "calls"
		case analyzer.ExtKindGRPCClient:
			name = "gRPC Service"
			systemKind = "grpc_service"
			description = "calls"
		default:
			continue
		}

		extSystemID := sanitizeID("system", "external-"+kindKey)

		model.Systems = append(model.Systems, System{
			ID:         extSystemID,
			Name:       name,
			SystemKind: systemKind,
			External:   true,
		})

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    mainSystemID,
			TargetID:    extSystemID,
			Description: description,
			Technology:  interaction.Technology,
			Level:       "system",
		})
	}
}

func (b *DefaultModelBuilder) buildComponentRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	seen := make(map[string]bool)
	modulePath := result.Module.ModulePath

	for _, call := range result.CallGraph {
		if !strings.HasPrefix(call.CallerPkg, modulePath) || !strings.HasPrefix(call.CalleePkg, modulePath) {
			continue
		}
		if call.CallerType == "" || call.CalleeType == "" {
			continue
		}

		sourceID := sanitizeID("component", call.CallerPkg+"/"+call.CallerType)
		targetID := sanitizeID("component", call.CalleePkg+"/"+call.CalleeType)

		if sourceID == targetID {
			continue
		}

		pairKey := sourceID + "->" + targetID
		if seen[pairKey] {
			continue
		}
		seen[pairKey] = true

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    sourceID,
			TargetID:    targetID,
			Description: "calls",
			Technology:  "Go",
			Level:       "component",
		})
	}
}

func (b *DefaultModelBuilder) buildInterfaceImplRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	for _, impl := range result.InterfaceImpls {
		sourceID := sanitizeID("component", impl.StructPkg+"/"+impl.StructName)
		targetID := sanitizeID("component", impl.InterfacePkg+"/"+impl.InterfaceName)

		if model.FindComponent(sourceID) == nil || model.FindComponent(targetID) == nil {
			continue
		}

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    sourceID,
			TargetID:    targetID,
			Description: "implements",
			Level:       "component",
		})
	}
}

func (b *DefaultModelBuilder) markEntrypoints(result *analyzer.AnalysisResult, model *C4Model) {
	for _, ep := range result.Entrypoints {
		switch ep.Kind {
		case "http_handler":
			// Find component by matching container's PackagePath to ep.PkgPath
			// If FuncName contains ".", split to get type.method and match by type
			typeName := ""
			if strings.Contains(ep.FuncName, ".") {
				parts := strings.SplitN(ep.FuncName, ".", 2)
				typeName = parts[0]
			}

			for i := range model.Components {
				comp := &model.Components[i]
				container := model.FindContainer(comp.ContainerID)
				if container == nil || container.PackagePath != ep.PkgPath {
					continue
				}
				if typeName != "" && comp.Name != typeName {
					continue
				}
				comp.IsEntrypoint = true
				comp.EntrypointKind = ep.Kind
				comp.EntrypointRoute = ep.Route
			}

		case "main":
			for i := range model.Components {
				comp := &model.Components[i]
				container := model.FindContainer(comp.ContainerID)
				if container == nil || container.PackagePath != ep.PkgPath {
					continue
				}
				comp.IsEntrypoint = true
				comp.EntrypointKind = ep.Kind
			}
		}
	}
}

// sanitizeID creates a URL-safe, deterministic ID from a prefix and path.
func sanitizeID(prefix, path string) string {
	id := strings.ReplaceAll(path, "/", "-")
	id = strings.ReplaceAll(id, ".", "-")
	return prefix + "-" + id
}
