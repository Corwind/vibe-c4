package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Corwind/vibe-c4/backend/internal/analyzer"
	"github.com/Corwind/vibe-c4/backend/internal/api"
	"github.com/Corwind/vibe-c4/backend/internal/c4model"
	"github.com/Corwind/vibe-c4/backend/internal/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(filename), "..", "analyzer", "testdata", "sample-project")
}

func setupRouter(t *testing.T) *http.ServeMux {
	t.Helper()
	a := analyzer.NewGoAnalyzer()
	b := c4model.NewModelBuilder()
	ps := project.NewService(a, b)
	h := api.NewHandlers(ps)
	_ = api.NewRouter(h)
	return nil
}

func setupFullRouter(t *testing.T) (*api.Handlers, *project.Service) {
	t.Helper()
	a := analyzer.NewGoAnalyzer()
	b := c4model.NewModelBuilder()
	ps := project.NewService(a, b)
	h := api.NewHandlers(ps)
	return h, ps
}

func analyzeProject(t *testing.T, router http.Handler) string {
	t.Helper()
	body := map[string]string{
		"path": testdataPath(t),
		"name": "sample-project",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "analyze should succeed: %s", rec.Body.String())

	var resp api.AnalyzeResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp.ID
}

func TestAnalyze_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	body := map[string]string{
		"path": testdataPath(t),
		"name": "sample-project",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AnalyzeResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "sample-project", resp.Name)
	assert.Equal(t, "complete", resp.Status)
}

func TestAnalyze_MissingPath(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	body := map[string]string{"name": "test"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnalyze_InvalidBody(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetDiagram_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var model c4model.C4Model
	err := json.Unmarshal(rec.Body.Bytes(), &model)
	require.NoError(t, err)
	assert.NotEmpty(t, model.Systems)
	assert.NotEmpty(t, model.Containers)
}

func TestGetDiagram_NotFound(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/nonexistent/diagram", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetContext_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/context", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var diagram api.DiagramResponse
	err := json.Unmarshal(rec.Body.Bytes(), &diagram)
	require.NoError(t, err)
	assert.NotEmpty(t, diagram.Nodes, "context diagram should have nodes")
}

func TestGetContainers_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/containers", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var diagram api.DiagramResponse
	err := json.Unmarshal(rec.Body.Bytes(), &diagram)
	require.NoError(t, err)
	assert.NotEmpty(t, diagram.Nodes, "container diagram should have nodes")
	assert.NotEmpty(t, diagram.Edges, "container diagram should have edges")
}

func getModel(t *testing.T, router http.Handler, projectID string) *c4model.C4Model {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID+"/diagram", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var model c4model.C4Model
	err := json.Unmarshal(rec.Body.Bytes(), &model)
	require.NoError(t, err)
	return &model
}

func TestGetComponents_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)
	model := getModel(t, router, id)

	// Find the container ID for internal/service
	var containerID string
	for _, c := range model.Containers {
		if c.Name == "internal/service" {
			containerID = c.ID
			break
		}
	}
	require.NotEmpty(t, containerID, "should find internal/service container")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/containers/"+containerID+"/components", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var diagram api.DiagramResponse
	err := json.Unmarshal(rec.Body.Bytes(), &diagram)
	require.NoError(t, err)
	assert.NotEmpty(t, diagram.Nodes, "component diagram should have nodes")
}

func TestGetComponents_ContainerNotFound(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/containers/nonexistent/components", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCode_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)
	model := getModel(t, router, id)

	// Find the component ID for Service
	var componentID string
	for _, c := range model.Components {
		if c.Name == "Service" {
			componentID = c.ID
			break
		}
	}
	require.NotEmpty(t, componentID, "should find Service component")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/components/"+componentID+"/code", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var diagram api.DiagramResponse
	err := json.Unmarshal(rec.Body.Bytes(), &diagram)
	require.NoError(t, err)
	assert.NotEmpty(t, diagram.Nodes, "code diagram should have nodes")
}

func TestGetCode_ComponentNotFound(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)
	id := analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+id+"/diagram/components/nonexistent/code", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
