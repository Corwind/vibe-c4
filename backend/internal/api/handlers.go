package api

import (
	"encoding/json"
	"net/http"

	"github.com/Corwind/vibe-c4/backend/internal/project"
	"github.com/go-chi/chi/v5"
)

// Handlers holds the HTTP handler dependencies.
type Handlers struct {
	projectService *project.Service
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(ps *project.Service) *Handlers {
	return &Handlers{projectService: ps}
}

func (h *Handlers) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path is required"})
		return
	}

	if req.Name == "" {
		req.Name = req.Path
	}

	p, err := h.projectService.AnalyzeFromPath(r.Context(), req.Path, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
			"id":    p.ID,
		})
		return
	}

	writeJSON(w, http.StatusOK, AnalyzeResponse{
		ID:     p.ID,
		Name:   p.Name,
		Status: string(p.Status),
	})
}

func (h *Handlers) handleGetDiagram(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, ok := h.projectService.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}

	if p.Model == nil {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": string(p.Status)})
		return
	}

	writeJSON(w, http.StatusOK, p.Model)
}

func (h *Handlers) handleGetContext(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, ok := h.projectService.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}

	if p.Model == nil {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": string(p.Status)})
		return
	}

	writeJSON(w, http.StatusOK, buildContextDiagram(p.Model))
}

func (h *Handlers) handleGetContainers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, ok := h.projectService.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}

	if p.Model == nil {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": string(p.Status)})
		return
	}

	writeJSON(w, http.StatusOK, buildContainerDiagram(p.Model))
}

func (h *Handlers) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	containerID := chi.URLParam(r, "containerID")

	p, ok := h.projectService.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}

	if p.Model == nil {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": string(p.Status)})
		return
	}

	// Verify container exists
	if p.Model.FindContainer(containerID) == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "container not found"})
		return
	}

	writeJSON(w, http.StatusOK, buildComponentDiagram(p.Model, containerID))
}

func (h *Handlers) handleGetCode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	componentID := chi.URLParam(r, "componentID")

	p, ok := h.projectService.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}

	if p.Model == nil {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": string(p.Status)})
		return
	}

	if p.Model.FindComponent(componentID) == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "component not found"})
		return
	}

	writeJSON(w, http.StatusOK, buildCodeDiagram(p.Model, componentID))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
