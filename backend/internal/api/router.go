package api

import (
	"encoding/json"
	"net/http"

	"github.com/Corwind/vibe-c4/backend/internal/project"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates and configures the chi router with middleware and routes.
func NewRouter(handlers ...*Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/v1/health", handleHealth)

	if len(handlers) > 0 {
		h := handlers[0]
		r.Get("/api/v1/projects", h.handleListProjects)
		r.Post("/api/v1/projects/analyze", h.handleAnalyze)
		r.Get("/api/v1/projects/{id}/diagram", h.handleGetDiagram)
		r.Get("/api/v1/projects/{id}/diagram/context", h.handleGetContext)
		r.Get("/api/v1/projects/{id}/diagram/containers", h.handleGetContainers)
		r.Get("/api/v1/projects/{id}/diagram/containers/{containerID}/components", h.handleGetComponents)
		r.Get("/api/v1/projects/{id}/diagram/components/{componentID}/code", h.handleGetCode)
	}

	return r
}

// NewDefaultRouter creates a fully-wired router with default dependencies.
func NewDefaultRouter(ps *project.Service) *chi.Mux {
	h := NewHandlers(ps)
	return NewRouter(h)
}

type healthResponse struct {
	Status string `json:"status"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
