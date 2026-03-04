package c4model

// DiagramLevel represents the C4 diagram abstraction level.
type DiagramLevel int

const (
	LevelContext   DiagramLevel = iota + 1 // Level 1 - System Context
	LevelContainer                         // Level 2 - Container
	LevelComponent                         // Level 3 - Component
	LevelCode                              // Level 4 - Code
)

// String returns the human-readable name of the diagram level.
func (d DiagramLevel) String() string {
	switch d {
	case LevelContext:
		return "Context"
	case LevelContainer:
		return "Container"
	case LevelComponent:
		return "Component"
	case LevelCode:
		return "Code"
	default:
		return "Unknown"
	}
}

// Relationship represents a directed relationship between two C4 elements.
type Relationship struct {
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
	Description string `json:"description"`
	Technology  string `json:"technology,omitempty"`
	Level       string `json:"level,omitempty"`
}

// System represents a C4 Level 1 system.
type System struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	External      bool   `json:"external"`
	ModulePath    string `json:"module_path,omitempty"`
	SystemKind    string `json:"system_kind,omitempty"`
	ContainerIDs  []string `json:"container_ids,omitempty"`
}

// Container represents a C4 Level 2 container (a major package/directory).
type Container struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Technology   string `json:"technology,omitempty"`
	PackagePath  string `json:"package_path"`
	SystemID     string `json:"system_id"`
	ComponentIDs []string `json:"component_ids,omitempty"`
}

// Component represents a C4 Level 3 component (an interface, struct, or key function).
type Component struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	Technology      string        `json:"technology,omitempty"`
	Type            ComponentType `json:"type"`
	ContainerID     string        `json:"container_id"`
	CodeElements    []string      `json:"code_elements,omitempty"`
	IsEntrypoint    bool          `json:"is_entrypoint,omitempty"`
	EntrypointKind  string        `json:"entrypoint_kind,omitempty"`
	EntrypointRoute string        `json:"entrypoint_route,omitempty"`
}

// ComponentType categorizes what kind of Go construct a component represents.
type ComponentType string

const (
	ComponentTypeStruct    ComponentType = "struct"
	ComponentTypeInterface ComponentType = "interface"
	ComponentTypeFunction  ComponentType = "function"
)

// CodeElement represents a C4 Level 4 code element (a function, method, or field).
type CodeElement struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        CodeElementType `json:"type"`
	ComponentID string          `json:"component_id"`
	FilePath    string          `json:"file_path,omitempty"`
	Line        int             `json:"line,omitempty"`
}

// CodeElementType categorizes what kind of code element this is.
type CodeElementType string

const (
	CodeElementTypeFunction CodeElementType = "function"
	CodeElementTypeMethod   CodeElementType = "method"
	CodeElementTypeField    CodeElementType = "field"
)

// C4Model is the aggregate root that holds the full C4 model for a project.
type C4Model struct {
	ProjectID     string         `json:"project_id"`
	Systems       []System       `json:"systems"`
	Containers    []Container    `json:"containers"`
	Components    []Component    `json:"components"`
	CodeElements  []CodeElement  `json:"code_elements"`
	Relationships []Relationship `json:"relationships"`
}

// FindSystem returns the system with the given ID, or nil if not found.
func (m *C4Model) FindSystem(id string) *System {
	for i := range m.Systems {
		if m.Systems[i].ID == id {
			return &m.Systems[i]
		}
	}
	return nil
}

// FindContainer returns the container with the given ID, or nil if not found.
func (m *C4Model) FindContainer(id string) *Container {
	for i := range m.Containers {
		if m.Containers[i].ID == id {
			return &m.Containers[i]
		}
	}
	return nil
}

// FindComponent returns the component with the given ID, or nil if not found.
func (m *C4Model) FindComponent(id string) *Component {
	for i := range m.Components {
		if m.Components[i].ID == id {
			return &m.Components[i]
		}
	}
	return nil
}

// ContainersForSystem returns all containers belonging to the given system.
func (m *C4Model) ContainersForSystem(systemID string) []Container {
	var result []Container
	for _, c := range m.Containers {
		if c.SystemID == systemID {
			result = append(result, c)
		}
	}
	return result
}

// ComponentsForContainer returns all components belonging to the given container.
func (m *C4Model) ComponentsForContainer(containerID string) []Component {
	var result []Component
	for _, c := range m.Components {
		if c.ContainerID == containerID {
			result = append(result, c)
		}
	}
	return result
}

// RelationshipsFrom returns all relationships originating from the given source ID.
func (m *C4Model) RelationshipsFrom(sourceID string) []Relationship {
	var result []Relationship
	for _, r := range m.Relationships {
		if r.SourceID == sourceID {
			result = append(result, r)
		}
	}
	return result
}

// RelationshipsTo returns all relationships targeting the given ID.
func (m *C4Model) RelationshipsTo(targetID string) []Relationship {
	var result []Relationship
	for _, r := range m.Relationships {
		if r.TargetID == targetID {
			result = append(result, r)
		}
	}
	return result
}
