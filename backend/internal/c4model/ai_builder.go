package c4model

import (
	"fmt"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
)

// AIModelBuilder builds a C4 model using AI-interpreted architecture on top of raw analysis data.
type AIModelBuilder struct {
	interpretation *llm.InterpretationResult
}

// NewAIModelBuilder creates a new AIModelBuilder from an LLM interpretation result.
func NewAIModelBuilder(interpretation *llm.InterpretationResult) *AIModelBuilder {
	return &AIModelBuilder{interpretation: interpretation}
}

// BuildFromAnalysis transforms analyzer output into a C4 model using AI interpretation
// for levels 1-3 and raw analysis data for level 4 code elements.
func (b *AIModelBuilder) BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}
	if b.interpretation == nil {
		return nil, fmt.Errorf("interpretation result is nil")
	}

	model := &C4Model{}

	b.buildSystemContext(model)
	b.buildContainers(model)
	b.buildComponents(model)
	b.buildRelationships(model)
	b.buildCodeElements(result, model)

	return model, nil
}

// buildSystemContext creates the Level 1 system context from AI interpretation.
func (b *AIModelBuilder) buildSystemContext(model *C4Model) {
	sc := b.interpretation.SystemContext

	// Main system
	mainSystemID := sanitizeID("system", sc.Name)
	model.Systems = append(model.Systems, System{
		ID:          mainSystemID,
		Name:        sc.Name,
		Description: sc.Description,
		External:    false,
	})

	// Actors as external system entries with actor kind
	for _, actor := range sc.Actors {
		actorID := sanitizeID("system", "actor-"+actor.Name)
		model.Systems = append(model.Systems, System{
			ID:          actorID,
			Name:        actor.Name,
			Description: actor.Description,
			External:    true,
			SystemKind:  "actor",
		})

		// Actor uses the main system
		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    actorID,
			TargetID:    mainSystemID,
			Description: "uses",
			Level:       "system",
		})
	}

	// External systems
	for _, ext := range sc.ExternalSystems {
		extID := sanitizeID("system", "external-"+ext.Name)
		model.Systems = append(model.Systems, System{
			ID:          extID,
			Name:        ext.Name,
			Description: ext.Description,
			External:    true,
			SystemKind:  ext.Kind,
		})

		// Main system uses external system
		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    mainSystemID,
			TargetID:    extID,
			Description: "uses",
			Technology:  ext.Technology,
			Level:       "system",
		})
	}
}

// buildContainers creates Level 2 containers from AI interpretation.
func (b *AIModelBuilder) buildContainers(model *C4Model) {
	mainSystemID := model.Systems[0].ID

	for _, c := range b.interpretation.Containers {
		container := Container{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Technology:  c.Technology,
			SystemID:    mainSystemID,
		}
		if len(c.PackagePaths) > 0 {
			container.PackagePath = strings.Join(c.PackagePaths, ",")
		}
		model.Containers = append(model.Containers, container)

		// Track container ID on the main system
		model.Systems[0].ContainerIDs = append(model.Systems[0].ContainerIDs, c.ID)
	}
}

// buildComponents creates Level 3 components from AI interpretation.
func (b *AIModelBuilder) buildComponents(model *C4Model) {
	for _, comp := range b.interpretation.Components {
		component := Component{
			ID:          comp.ID,
			Name:        comp.Name,
			Description: comp.Description,
			Type:        roleToComponentType(comp.Role),
			ContainerID: comp.ContainerID,
		}
		model.Components = append(model.Components, component)

		// Track component ID on its container
		for i := range model.Containers {
			if model.Containers[i].ID == comp.ContainerID {
				model.Containers[i].ComponentIDs = append(model.Containers[i].ComponentIDs, comp.ID)
				break
			}
		}
	}
}

// buildRelationships creates relationships from AI interpretation.
func (b *AIModelBuilder) buildRelationships(model *C4Model) {
	for _, rel := range b.interpretation.Relationships {
		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    rel.SourceID,
			TargetID:    rel.TargetID,
			Description: rel.Description,
			Level:       rel.Level,
		})
	}
}

// buildCodeElements creates Level 4 code elements from raw analysis data,
// matching structs to AI-assigned components via StructNames.
func (b *AIModelBuilder) buildCodeElements(result *analyzer.AnalysisResult, model *C4Model) {
	// Build a lookup: struct name -> component ID from AI interpretation
	structToComponent := make(map[string]string)
	for _, comp := range b.interpretation.Components {
		for _, sn := range comp.StructNames {
			structToComponent[sn] = comp.ID
		}
	}

	for _, pkg := range result.Packages {
		for _, s := range pkg.Structs {
			// Resolve component ID: prefer AI match, fall back to sanitizeID
			compID, ok := structToComponent[s.Name]
			if !ok {
				compID = sanitizeID("component", pkg.ImportPath+"/"+s.Name)
			}

			// Methods
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

			// Fields
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

		// Standalone functions go to their container
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

// roleToComponentType maps an AI role string to a ComponentType.
func roleToComponentType(role string) ComponentType {
	switch role {
	case "controller", "service", "repository", "gateway", "domain":
		return ComponentTypeStruct
	default:
		return ComponentTypeStruct
	}
}
