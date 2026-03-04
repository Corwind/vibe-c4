package handler

import (
	"net/http"

	"github.com/example/sample-project/internal/service"
	"github.com/example/sample-project/pkg/utils"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.Log("health check")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DoWork(w http.ResponseWriter, r *http.Request) {
	result := h.svc.DoWork()
	utils.Log("work done: " + result)
	w.Write([]byte(result))
}

func (h *Handler) ServeHTTP() {
	utils.Log("serving request")
}
