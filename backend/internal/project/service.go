package project

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/llm"
)

// Status represents the analysis status of a project.
type Status string

const (
	StatusPending  Status = "pending"
	StatusRunning  Status = "running"
	StatusComplete Status = "completed"
	StatusFailed   Status = "failed"
)

// Project holds project metadata and its C4 model.
type Project struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Path   string           `json:"path"`
	Status Status           `json:"status"`
	Model  *c4model.C4Model `json:"model,omitempty"`
	Error  string           `json:"error,omitempty"`
}

// ServiceOption configures the project Service.
type ServiceOption func(*Service)

// WithInterpreter adds an AI interpreter to the service pipeline.
func WithInterpreter(i llm.Interpreter) ServiceOption {
	return func(s *Service) {
		s.interpreter = i
	}
}

// Service manages projects and their analysis.
type Service struct {
	mu          sync.RWMutex
	projects    map[string]*Project
	analyzer    analyzer.Analyzer
	builder     c4model.ModelBuilder
	interpreter llm.Interpreter
}

// NewService creates a new project service.
func NewService(a analyzer.Analyzer, b c4model.ModelBuilder, opts ...ServiceOption) *Service {
	s := &Service{
		projects: make(map[string]*Project),
		analyzer: a,
		builder:  b,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AnalyzeFromPath triggers analysis of a Go project at the given path.
func (s *Service) AnalyzeFromPath(ctx context.Context, projectPath string, name string) (*Project, error) {
	id := generateID(projectPath)

	p := &Project{
		ID:     id,
		Name:   name,
		Path:   projectPath,
		Status: StatusRunning,
	}

	s.mu.Lock()
	s.projects[id] = p
	s.mu.Unlock()

	result, err := s.analyzer.AnalyzeProject(ctx, projectPath)
	if err != nil {
		s.mu.Lock()
		p.Status = StatusFailed
		p.Error = err.Error()
		s.mu.Unlock()
		return p, fmt.Errorf("analysis failed: %w", err)
	}

	builder := s.builderForResult(ctx, result, projectPath)

	model, err := builder.BuildFromAnalysis(result)
	if err != nil {
		s.mu.Lock()
		p.Status = StatusFailed
		p.Error = err.Error()
		s.mu.Unlock()
		return p, fmt.Errorf("model building failed: %w", err)
	}

	model.ProjectID = id

	s.mu.Lock()
	p.Model = model
	p.Status = StatusComplete
	s.mu.Unlock()

	return p, nil
}

// builderForResult returns an AI-enhanced builder if an interpreter is configured and succeeds,
// otherwise falls back to the default static builder.
func (s *Service) builderForResult(ctx context.Context, result *analyzer.AnalysisResult, projectPath string) c4model.ModelBuilder {
	if s.interpreter == nil {
		return s.builder
	}

	preparer := &llm.ContextPreparer{}
	input := preparer.Prepare(result, projectPath)

	interpretation, err := s.interpreter.Interpret(ctx, input)
	if err != nil {
		log.Printf("AI interpretation failed, falling back to static builder: %v", err)
		return s.builder
	}

	return c4model.NewAIModelBuilder(interpretation)
}

// Get returns a project by ID.
func (s *Service) Get(id string) (*Project, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	return p, ok
}

// List returns all projects.
func (s *Service) List() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projects := make([]*Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	return projects
}

func generateID(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h[:8])
}
