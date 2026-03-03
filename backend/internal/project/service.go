package project

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
)

// Status represents the analysis status of a project.
type Status string

const (
	StatusPending  Status = "pending"
	StatusRunning  Status = "running"
	StatusComplete Status = "complete"
	StatusFailed   Status = "failed"
)

// Project holds project metadata and its C4 model.
type Project struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Path   string        `json:"path"`
	Status Status        `json:"status"`
	Model  *c4model.C4Model `json:"model,omitempty"`
	Error  string        `json:"error,omitempty"`
}

// Service manages projects and their analysis.
type Service struct {
	mu       sync.RWMutex
	projects map[string]*Project
	analyzer analyzer.Analyzer
	builder  c4model.ModelBuilder
}

// NewService creates a new project service.
func NewService(a analyzer.Analyzer, b c4model.ModelBuilder) *Service {
	return &Service{
		projects: make(map[string]*Project),
		analyzer: a,
		builder:  b,
	}
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

	model, err := s.builder.BuildFromAnalysis(result)
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

// Get returns a project by ID.
func (s *Service) Get(id string) (*Project, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	return p, ok
}

func generateID(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h[:8])
}
