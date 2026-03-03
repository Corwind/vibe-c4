package handler

import (
	"github.com/example/sample-project/internal/service"
	"github.com/example/sample-project/pkg/utils"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ServeHTTP() {
	utils.Log("serving request")
}
