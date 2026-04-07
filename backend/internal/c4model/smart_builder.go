package c4model

import (
	"fmt"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
)

// SmartModelBuilder implements ModelBuilder using deterministic heuristics
// to produce proper C4 diagrams from static analysis data.
type SmartModelBuilder struct{}

// NewSmartModelBuilder creates a new SmartModelBuilder.
func NewSmartModelBuilder() *SmartModelBuilder {
	return &SmartModelBuilder{}
}

// BuildFromAnalysis transforms analyzer output into a C4 model using smart heuristics.
func (b *SmartModelBuilder) BuildFromAnalysis(result *analyzer.AnalysisResult) (*C4Model, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}

	model := &C4Model{}

	b.buildSystemContext(result, model)
	pkgToContainer := b.buildContainers(result, model)
	b.buildComponents(result, model, pkgToContainer)
	b.buildCodeElements(result, model)
	b.enrichRelationships(result, model)

	return model, nil
}

// buildSystemContext creates Level 1: main system, actors, and external systems.
func (b *SmartModelBuilder) buildSystemContext(result *analyzer.AnalysisResult, model *C4Model) {
	modulePath := result.Module.ModulePath

	mainSystemID := sanitizeID("system", modulePath)
	description := generateSystemDescription(result.Entrypoints, result.ExternalInteractions)

	model.Systems = append(model.Systems, System{
		ID:          mainSystemID,
		Name:        moduleName(modulePath),
		Description: description,
		ModulePath:  modulePath,
		External:    false,
	})

	b.inferActors(result.Entrypoints, mainSystemID, model)
	b.addExternalSystems(result.ExternalInteractions, mainSystemID, model)
}

// inferActors creates actor systems from entrypoint kinds.
func (b *SmartModelBuilder) inferActors(entrypoints []analyzer.Entrypoint, mainSystemID string, model *C4Model) {
	seenActors := make(map[string]bool)

	for _, ep := range entrypoints {
		var actorName string
		switch ep.Kind {
		case "http_handler":
			actorName = "User"
		case "grpc_handler", "grpc_server":
			actorName = "Client Service"
		default:
			continue
		}

		if seenActors[actorName] {
			continue
		}
		seenActors[actorName] = true

		actorID := sanitizeID("system", "actor-"+actorName)
		model.Systems = append(model.Systems, System{
			ID:         actorID,
			Name:       actorName,
			External:   true,
			SystemKind: "actor",
		})

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    actorID,
			TargetID:    mainSystemID,
			Description: "uses",
			Level:       "system",
		})
	}
}

// addExternalSystems creates infrastructure external systems from ExternalInteractions.
func (b *SmartModelBuilder) addExternalSystems(interactions []analyzer.ExternalInteraction, mainSystemID string, model *C4Model) {
	infraGroups := buildInfraContainers(interactions)

	for _, g := range infraGroups {
		extSystemID := sanitizeID("system", "external-"+g.SystemKind)

		model.Systems = append(model.Systems, System{
			ID:         extSystemID,
			Name:       g.Name,
			External:   true,
			SystemKind: g.SystemKind,
		})

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    mainSystemID,
			TargetID:    extSystemID,
			Description: generateContainerDescription(g),
			Technology:  g.Technology,
			Level:       "system",
		})
	}
}

// buildContainers creates Level 2 containers and returns a package→containerID lookup.
func (b *SmartModelBuilder) buildContainers(result *analyzer.AnalysisResult, model *C4Model) map[string]string {
	mainSystemID := sanitizeID("system", result.Module.ModulePath)
	groups := groupPackagesIntoContainers(result.Packages, result.ImportGraph, result.ExternalInteractions, result.Module.ModulePath)

	pkgToContainer := make(map[string]string)

	for _, g := range groups {
		if g.IsInfra {
			continue
		}

		containerID := sanitizeID("container", g.Name)
		model.Containers = append(model.Containers, Container{
			ID:          containerID,
			Name:        g.Name,
			Description: generateContainerDescription(g),
			Technology:  g.Technology,
			SystemID:    mainSystemID,
		})

		model.Systems[0].ContainerIDs = append(model.Systems[0].ContainerIDs, containerID)

		for _, pkg := range g.Packages {
			pkgToContainer[pkg] = containerID
		}
	}

	// Container-to-container relationships from import graph
	b.buildContainerRelationships(result, model, pkgToContainer)

	return pkgToContainer
}

// buildContainerRelationships creates relationships between containers based on the import graph.
func (b *SmartModelBuilder) buildContainerRelationships(result *analyzer.AnalysisResult, model *C4Model, pkgToContainer map[string]string) {
	seen := make(map[string]bool)
	for source, targets := range result.ImportGraph {
		sourceContainer := pkgToContainer[source]
		if sourceContainer == "" {
			continue
		}
		for _, target := range targets {
			targetContainer := pkgToContainer[target]
			if targetContainer == "" || targetContainer == sourceContainer {
				continue
			}
			pairKey := sourceContainer + "->" + targetContainer
			if seen[pairKey] {
				continue
			}
			seen[pairKey] = true

			model.Relationships = append(model.Relationships, Relationship{
				SourceID:    sourceContainer,
				TargetID:    targetContainer,
				Description: "uses",
				Technology:  "Go",
				Level:       "container",
			})
		}
	}
}

// buildComponents creates Level 3 components: classified structs and interfaces.
func (b *SmartModelBuilder) buildComponents(result *analyzer.AnalysisResult, model *C4Model, pkgToContainer map[string]string) {
	modulePath := result.Module.ModulePath

	// Build a lookup: external system kind → system ID
	kindToExtSystemID := make(map[string]string)
	for _, sys := range model.Systems {
		if sys.External && sys.SystemKind != "" && sys.SystemKind != "actor" {
			kindToExtSystemID[sys.SystemKind] = sys.ID
		}
	}

	for _, pkg := range result.Packages {
		containerID := pkgToContainer[pkg.ImportPath]
		if containerID == "" {
			continue
		}

		for _, s := range pkg.Structs {
			ctx := classificationContext{
				structInfo:           s,
				pkgPath:              pkg.ImportPath,
				entrypoints:          result.Entrypoints,
				externalInteractions: result.ExternalInteractions,
				callGraph:            result.CallGraph,
				modulePath:           modulePath,
			}
			role := classifyStructRole(ctx)
			if role == RoleSkip {
				continue
			}

			compID := sanitizeID("component", pkg.ImportPath+"/"+s.Name)
			model.Components = append(model.Components, Component{
				ID:          compID,
				Name:        s.Name,
				Description: generateComponentDescription(s.Name, role),
				Type:        ComponentTypeStruct,
				Technology:  "Go " + string(role),
				ContainerID: containerID,
				Role:        string(role),
				PackagePath: pkg.ImportPath,
				Methods:     FormatMethodSignatures(s.Methods),
				Fields:      FormatFieldDescriptors(s.Fields),
			})

			b.addComponentToContainer(model, containerID, compID)
			b.markComponentEntrypoints(s.Name, pkg.ImportPath, result.Entrypoints, compID, model)
			b.addComponentInfraRelationships(s.Name, pkg.ImportPath, result.ExternalInteractions, compID, kindToExtSystemID, model)
		}

		for _, iface := range pkg.Interfaces {
			compID := sanitizeID("component", pkg.ImportPath+"/"+iface.Name)
			model.Components = append(model.Components, Component{
				ID:          compID,
				Name:        iface.Name,
				Type:        ComponentTypeInterface,
				Technology:  "Go interface",
				ContainerID: containerID,
				PackagePath: pkg.ImportPath,
				Methods:     FormatMethodSignatures(iface.Methods),
			})
			b.addComponentToContainer(model, containerID, compID)
		}
	}
}

// markComponentEntrypoints marks a component as an entrypoint if its struct has methods in entrypoints.
func (b *SmartModelBuilder) markComponentEntrypoints(structName, pkgPath string, entrypoints []analyzer.Entrypoint, compID string, model *C4Model) {
	for _, ep := range entrypoints {
		if ep.PkgPath != pkgPath {
			continue
		}
		prefix := structName + "."
		if strings.Contains(ep.FuncName, prefix) {
			for i := range model.Components {
				if model.Components[i].ID == compID {
					model.Components[i].IsEntrypoint = true
					model.Components[i].EntrypointKind = ep.Kind
					model.Components[i].EntrypointRoute = ep.Route
					break
				}
			}
			break
		}
	}
}

// addComponentInfraRelationships creates relationships from components to infrastructure external systems.
func (b *SmartModelBuilder) addComponentInfraRelationships(structName, pkgPath string, interactions []analyzer.ExternalInteraction, compID string, kindToExtSystemID map[string]string, model *C4Model) {
	seen := make(map[string]bool)

	for _, ei := range interactions {
		if ei.PkgPath != pkgPath || ei.TypeName != structName {
			continue
		}

		var systemKind string
		switch ei.Kind {
		case analyzer.ExtKindDatabase:
			systemKind = "database"
		case analyzer.ExtKindKafkaProducer, analyzer.ExtKindKafkaConsumer:
			systemKind = "message_queue"
		case analyzer.ExtKindHTTPClient:
			systemKind = "http_api"
		case analyzer.ExtKindGRPCClient:
			systemKind = "grpc_service"
		default:
			continue
		}

		extSystemID := kindToExtSystemID[systemKind]
		if extSystemID == "" {
			continue
		}

		pairKey := compID + "->" + extSystemID
		if seen[pairKey] {
			continue
		}
		seen[pairKey] = true

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    compID,
			TargetID:    extSystemID,
			Description: generateRelationshipDescription(ei.Kind),
			Technology:  ei.Technology,
			Level:       "component",
		})
	}
}

// buildComponentCallRelationships creates component-to-component relationships from the call graph.
// It collects callee function names per source→target pair to produce descriptions like "calls GetUser, FindByID".
func (b *SmartModelBuilder) buildComponentCallRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	modulePath := result.Module.ModulePath

	// Collect callee function names per source→target pair.
	type pairKey struct{ source, target string }
	pairFuncs := make(map[pairKey][]string)
	pairOrder := make([]pairKey, 0)

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

		if model.FindComponent(sourceID) == nil || model.FindComponent(targetID) == nil {
			continue
		}

		key := pairKey{sourceID, targetID}
		if _, exists := pairFuncs[key]; !exists {
			pairOrder = append(pairOrder, key)
		}
		pairFuncs[key] = append(pairFuncs[key], call.CalleeFunc)
	}

	for _, key := range pairOrder {
		funcs := dedupStrings(pairFuncs[key])
		desc := "calls " + strings.Join(funcs, ", ")

		model.Relationships = append(model.Relationships, Relationship{
			SourceID:    key.source,
			TargetID:    key.target,
			Description: desc,
			Technology:  "Go",
			Level:       "component",
		})
	}
}

// buildInterfaceImplRelationships creates interface implementation relationships.
func (b *SmartModelBuilder) buildInterfaceImplRelationships(result *analyzer.AnalysisResult, model *C4Model) {
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

// buildCodeElements creates Level 4 code elements: methods, fields, standalone functions.
func (b *SmartModelBuilder) buildCodeElements(result *analyzer.AnalysisResult, model *C4Model) {
	for _, pkg := range result.Packages {
		for _, s := range pkg.Structs {
			compID := sanitizeID("component", pkg.ImportPath+"/"+s.Name)

			// Only add code elements for components that exist in the model
			if model.FindComponent(compID) == nil {
				continue
			}

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

	}
}

func (b *SmartModelBuilder) addComponentToContainer(model *C4Model, containerID, componentID string) {
	for i := range model.Containers {
		if model.Containers[i].ID == containerID {
			model.Containers[i].ComponentIDs = append(model.Containers[i].ComponentIDs, componentID)
			return
		}
	}
}

// enrichRelationships centralizes all relationship building after components exist.
func (b *SmartModelBuilder) enrichRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	b.buildComponentCallRelationships(result, model)
	b.buildInterfaceImplRelationships(result, model)
	b.buildFieldDependencyRelationships(result, model)
	b.enrichFromConstructors(result, model)
}

// buildFieldDependencyRelationships scans each component's field types,
// resolves internal struct references to component IDs, and creates "depends on"
// (or "extends" for embedded) relationships.
func (b *SmartModelBuilder) buildFieldDependencyRelationships(result *analyzer.AnalysisResult, model *C4Model) {
	// Build lookup maps.
	pkgNameToPath := make(map[string]string)
	for _, pkg := range result.Packages {
		pkgNameToPath[pkg.Name] = pkg.ImportPath
	}

	structToCompID := make(map[string]string)
	for _, comp := range model.Components {
		structToCompID[comp.PackagePath+"/"+comp.Name] = comp.ID
	}

	seen := make(map[string]bool)

	for _, pkg := range result.Packages {
		for _, s := range pkg.Structs {
			sourceID := structToCompID[pkg.ImportPath+"/"+s.Name]
			if sourceID == "" {
				continue
			}

			for _, f := range s.Fields {
				targetID := ResolveFieldTypeToComponentID(f.Type, pkg.ImportPath, pkgNameToPath, structToCompID)
				if targetID == "" || targetID == sourceID {
					continue
				}

				pairKey := sourceID + "->" + targetID
				if seen[pairKey] {
					continue
				}
				seen[pairKey] = true

				desc := "depends on"
				if f.Embedded {
					desc = "extends"
				}

				model.Relationships = append(model.Relationships, Relationship{
					SourceID:    sourceID,
					TargetID:    targetID,
					Description: desc,
					Technology:  "Go",
					Level:       "component",
				})
			}
		}
	}
}

// enrichFromConstructors detects New* functions, resolves their parameter types
// to components, and adds "depends on" relationships not already covered.
func (b *SmartModelBuilder) enrichFromConstructors(result *analyzer.AnalysisResult, model *C4Model) {
	pkgNameToPath := make(map[string]string)
	for _, pkg := range result.Packages {
		pkgNameToPath[pkg.Name] = pkg.ImportPath
	}

	structToCompID := make(map[string]string)
	for _, comp := range model.Components {
		structToCompID[comp.PackagePath+"/"+comp.Name] = comp.ID
	}

	// Build set of existing relationships for dedup.
	existingRels := make(map[string]bool)
	for _, r := range model.Relationships {
		existingRels[r.SourceID+"->"+r.TargetID] = true
	}

	for _, pkg := range result.Packages {
		for _, fn := range pkg.Functions {
			if !strings.HasPrefix(fn.Name, "New") {
				continue
			}

			// The return type of a New* function tells us which component it constructs.
			// Try to find the component in this package.
			structName := strings.TrimPrefix(fn.Name, "New")
			sourceID := structToCompID[pkg.ImportPath+"/"+structName]
			if sourceID == "" {
				continue
			}

			// Resolve parameter types to components.
			for _, param := range fn.Params {
				targetID := ResolveFieldTypeToComponentID(param, pkg.ImportPath, pkgNameToPath, structToCompID)
				if targetID == "" || targetID == sourceID {
					continue
				}

				pairKey := sourceID + "->" + targetID
				if existingRels[pairKey] {
					continue
				}
				existingRels[pairKey] = true

				model.Relationships = append(model.Relationships, Relationship{
					SourceID:    sourceID,
					TargetID:    targetID,
					Description: "depends on",
					Technology:  "Go",
					Level:       "component",
				})
			}
		}
	}
}

// ResolveFieldTypeToComponentID strips pointer/slice prefixes, resolves
// qualified (pkg.Type) and unqualified type names to component IDs.
func ResolveFieldTypeToComponentID(typeName, currentPkgPath string, pkgNameToPath map[string]string, structToCompID map[string]string) string {
	// Strip pointer and slice prefixes.
	t := typeName
	for strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") {
		t = strings.TrimPrefix(t, "*")
		t = strings.TrimPrefix(t, "[]")
	}

	// Qualified type: "pkg.Type"
	if strings.Contains(t, ".") {
		parts := strings.SplitN(t, ".", 2)
		pkgName := parts[0]
		structName := parts[1]

		pkgPath, ok := pkgNameToPath[pkgName]
		if !ok {
			return "" // external package, not in our model
		}
		return structToCompID[pkgPath+"/"+structName]
	}

	// Unqualified type: look in current package.
	if id := structToCompID[currentPkgPath+"/"+t]; id != "" {
		return id
	}

	return ""
}

// FormatMethodSignatures formats method info into human-readable signatures.
func FormatMethodSignatures(methods []analyzer.MethodInfo) []string {
	if len(methods) == 0 {
		return nil
	}
	sigs := make([]string, 0, len(methods))
	for _, m := range methods {
		sig := m.Name + "(" + strings.Join(m.Params, ", ") + ")"
		if len(m.Returns) > 0 {
			if len(m.Returns) == 1 {
				sig += " " + m.Returns[0]
			} else {
				sig += " (" + strings.Join(m.Returns, ", ") + ")"
			}
		}
		sigs = append(sigs, sig)
	}
	return sigs
}

// FormatFieldDescriptors formats field info into human-readable descriptors.
func FormatFieldDescriptors(fields []analyzer.FieldInfo) []string {
	if len(fields) == 0 {
		return nil
	}
	descs := make([]string, 0, len(fields))
	for _, f := range fields {
		if f.Embedded {
			descs = append(descs, f.Type+" (embedded)")
		} else {
			descs = append(descs, f.Name+" "+f.Type)
		}
	}
	return descs
}

// dedupStrings removes duplicates from a string slice while preserving order.
func dedupStrings(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// moduleName extracts a readable name from a module path.
func moduleName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	if len(parts) == 0 {
		return modulePath
	}
	return parts[len(parts)-1]
}
