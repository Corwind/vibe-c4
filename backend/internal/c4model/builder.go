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

// sanitizeID creates a URL-safe, deterministic ID from a prefix and path.
func sanitizeID(prefix, path string) string {
	id := strings.ReplaceAll(path, "/", "-")
	id = strings.ReplaceAll(id, ".", "-")
	return prefix + "-" + id
}
