package service

import "github.com/example/sample-project/pkg/utils"

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) DoWork() string {
	utils.Log("doing work")
	return "done"
}
