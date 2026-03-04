package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Corwind/vibe-c4/backend/internal/project"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 100 << 20 // 100 MB

// Handlers holds the HTTP handler dependencies.
type Handlers struct {
	projectService *project.Service
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(ps *project.Service) *Handlers {
	return &Handlers{projectService: ps}
}

func (h *Handlers) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects := h.projectService.List()
	resp := ListProjectsResponse{
		Data: make([]ProjectResponse, 0, len(projects)),
	}
	for _, p := range projects {
		resp.Data = append(resp.Data, ProjectResponse{
			ID:     p.ID,
			Name:   p.Name,
			Source: p.Path,
			Status: string(p.Status),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		h.handleAnalyzeUpload(w, r)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.GitURL != "" {
		h.handleAnalyzeGitURL(w, r, req)
		return
	}

	if req.Path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path or gitUrl is required"})
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

	writeJSON(w, http.StatusOK, map[string]any{
		"data": AnalyzeResponse{
			ID:     p.ID,
			Name:   p.Name,
			Status: string(p.Status),
		},
	})
}

func (h *Handlers) handleAnalyzeGitURL(w http.ResponseWriter, r *http.Request, req AnalyzeRequest) {
	cloneDir, err := os.MkdirTemp("", "vibe-c4-git-*")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create temp directory"})
		return
	}

	cmd := exec.CommandContext(r.Context(), "git", "clone", "--depth", "1", req.GitURL, cloneDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(cloneDir)
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("failed to clone repository: %s", string(output)),
		})
		return
	}

	name := req.Name
	if name == "" {
		// Extract repo name from URL (e.g., "https://github.com/user/repo" -> "repo")
		name = filepath.Base(strings.TrimSuffix(req.GitURL, ".git"))
	}

	p, err := h.projectService.AnalyzeFromPath(r.Context(), cloneDir, name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
			"id":    p.ID,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": AnalyzeResponse{
			ID:     p.ID,
			Name:   p.Name,
			Status: string(p.Status),
		},
	})
}

func (h *Handlers) handleAnalyzeUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file field is required"})
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "vibe-c4-upload-*.zip")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create temp file"})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save uploaded file"})
		return
	}
	tmpFile.Close()

	extractDir, err := os.MkdirTemp("", "vibe-c4-project-*")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create extraction directory"})
		return
	}

	if err := extractZip(tmpFile.Name(), extractDir); err != nil {
		os.RemoveAll(extractDir)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("failed to extract zip: %v", err)})
		return
	}

	// Find the project root: if the zip contains a single directory, use it
	projectDir := findProjectRoot(extractDir)

	name := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))

	p, err := h.projectService.AnalyzeFromPath(r.Context(), projectDir, name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
			"id":    p.ID,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": AnalyzeResponse{
			ID:     p.ID,
			Name:   p.Name,
			Status: string(p.Status),
		},
	})
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := filepath.Join(destDir, f.Name)

		// Prevent zip slip attack
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// findProjectRoot checks if extractDir contains a single directory and returns it,
// otherwise returns extractDir itself.
func findProjectRoot(extractDir string) string {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return extractDir
	}

	// Filter out hidden files (like __MACOSX)
	var dirs []os.DirEntry
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") && !strings.HasPrefix(e.Name(), "__") {
			dirs = append(dirs, e)
		}
	}

	if len(dirs) == 1 && dirs[0].IsDir() {
		return filepath.Join(extractDir, dirs[0].Name())
	}
	return extractDir
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
