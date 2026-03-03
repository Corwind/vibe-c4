package service

import "github.com/example/sample-project/pkg/utils"

// Worker is the interface for doing work.
type Worker interface {
	DoWork() string
	Stop() error
}

// BaseService provides shared service functionality.
type BaseService struct {
	Name string
}

// Service implements Worker.
type Service struct {
	BaseService
	running bool
}

func New() *Service {
	return &Service{
		BaseService: BaseService{Name: "default"},
	}
}

func (s *Service) DoWork() string {
	utils.Log("doing work")
	return "done"
}

func (s *Service) Stop() error {
	s.running = false
	return nil
}

func HelperFunc(msg string) string {
	return "helper: " + msg
}
