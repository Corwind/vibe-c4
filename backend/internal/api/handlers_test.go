package api_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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

type wrappedAnalyzeResponse struct {
	Data api.AnalyzeResponse `json:"data"`
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

	var resp wrappedAnalyzeResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp.Data.ID
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

	var resp wrappedAnalyzeResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Data.ID)
	assert.Equal(t, "sample-project", resp.Data.Name)
	assert.Equal(t, "complete", resp.Data.Status)
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

func TestListProjects_Empty(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ListProjectsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Data)
}

func TestListProjects_WithProjects(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	analyzeProject(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ListProjectsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "sample-project", resp.Data[0].Name)
	assert.Equal(t, "complete", resp.Data[0].Status)
}

func TestAnalyzeUpload_Success(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	// Create a zip file containing the sample project
	zipBuf := createTestZip(t, testdataPath(t))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "sample-project.zip")
	require.NoError(t, err)
	_, err = io.Copy(part, zipBuf)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "response: %s", rec.Body.String())

	var resp wrappedAnalyzeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Data.ID)
	assert.Equal(t, "sample-project", resp.Data.Name)
	assert.Equal(t, "complete", resp.Data.Status)
}

func TestAnalyzeUpload_MissingFile(t *testing.T) {
	h, _ := setupFullRouter(t)
	router := api.NewRouter(h)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// createTestZip creates a zip archive of the given directory and returns it as a bytes.Reader.
func createTestZip(t *testing.T, srcDir string) *bytes.Reader {
	t.Helper()
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			_, err := w.Create(relPath + "/")
			return err
		}

		f, err := w.Create(relPath)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = f.Write(content)
		return err
	})
	require.NoError(t, err)
	require.NoError(t, w.Close())

	return bytes.NewReader(buf.Bytes())
}
