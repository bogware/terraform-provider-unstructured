// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *UnstructuredClient) {
	t.Helper()
	ts := httptest.NewServer(handler)
	client := NewUnstructuredClient("test-api-key", ts.URL, "test")
	return ts, client
}

func TestNewUnstructuredClient(t *testing.T) {
	c := NewUnstructuredClient("key", "", "1.0.0")
	if c.apiURL != defaultAPIURL {
		t.Errorf("expected default URL %s, got %s", defaultAPIURL, c.apiURL)
	}
	if c.apiKey != "key" {
		t.Errorf("expected api key 'key', got %s", c.apiKey)
	}
	if c.userAgent != "terraform-provider-unstructured/1.0.0" {
		t.Errorf("expected user agent terraform-provider-unstructured/1.0.0, got %s", c.userAgent)
	}
}

func TestNewUnstructuredClientTrimsTrailingSlash(t *testing.T) {
	c := NewUnstructuredClient("key", "https://example.com/api/v1/", "1.0.0")
	if c.apiURL != "https://example.com/api/v1" {
		t.Errorf("expected trailing slash stripped, got %s", c.apiURL)
	}
}

func TestDoRequestSetsHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("unstructured-api-key"); got != "test-key" {
			t.Errorf("expected api key header 'test-key', got %q", got)
		}
		if got := r.Header.Get("accept"); got != "application/json" {
			t.Errorf("expected accept header 'application/json', got %q", got)
		}
		if got := r.Header.Get("user-agent"); got != "terraform-provider-unstructured/test" {
			t.Errorf("expected user-agent header, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewUnstructuredClient("test-key", ts.URL, "test")
	_, _, err := client.doRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRequestSetsContentTypeForBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("content-type"); got != "application/json" {
			t.Errorf("expected content-type 'application/json', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewUnstructuredClient("key", ts.URL, "test")
	_, _, err := client.doRequest(context.Background(), http.MethodPost, "/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Source CRUD Tests ---

func TestCreateSource(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/sources/" {
			t.Errorf("expected /sources/, got %s", r.URL.Path)
		}

		var body CreateSourceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Name != "test-source" || body.Type != "s3" {
			t.Errorf("unexpected body: %+v", body)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SourceConnector{
			ID:        "src-123",
			Name:      body.Name,
			Type:      body.Type,
			Config:    body.Config,
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	})
	defer ts.Close()

	result, err := client.CreateSource(context.Background(), CreateSourceRequest{
		Name:   "test-source",
		Type:   "s3",
		Config: map[string]interface{}{"remote_url": "s3://bucket"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "src-123" {
		t.Errorf("expected ID src-123, got %s", result.ID)
	}
}

func TestGetSource(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/sources/src-123" {
			t.Errorf("expected /sources/src-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SourceConnector{
			ID:        "src-123",
			Name:      "test",
			Type:      "s3",
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	})
	defer ts.Close()

	result, err := client.GetSource(context.Background(), "src-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "src-123" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetSourceNotFound(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"not found"}`))
	})
	defer ts.Close()

	result, err := client.GetSource(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for 404, got %+v", result)
	}
}

func TestUpdateSource(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		updatedAt := "2025-06-01T00:00:00Z"
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SourceConnector{
			ID:        "src-123",
			Name:      "updated",
			Type:      "s3",
			CreatedAt: "2025-01-01T00:00:00Z",
			UpdatedAt: &updatedAt,
		})
	})
	defer ts.Close()

	result, err := client.UpdateSource(context.Background(), "src-123", UpdateSourceRequest{
		Config: map[string]interface{}{"remote_url": "s3://updated-bucket"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "updated" {
		t.Errorf("expected name 'updated', got %s", result.Name)
	}
	if result.UpdatedAt == nil {
		t.Error("expected updated_at to be set")
	}
}

func TestDeleteSource(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := client.DeleteSource(context.Background(), "src-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSourceNotFoundIsIdempotent(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer ts.Close()

	err := client.DeleteSource(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for 404 delete, got: %v", err)
	}
}

func TestListSources(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]SourceConnector{
			{ID: "src-1", Name: "one", Type: "s3", CreatedAt: "2025-01-01T00:00:00Z"},
			{ID: "src-2", Name: "two", Type: "azure", CreatedAt: "2025-01-01T00:00:00Z"},
		})
	})
	defer ts.Close()

	results, err := client.ListSources(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 sources, got %d", len(results))
	}
}

// --- Destination CRUD Tests ---

func TestCreateDestination(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/destinations/" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(DestinationConnector{
			ID:        "dst-123",
			Name:      "test-dest",
			Type:      "elasticsearch",
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	})
	defer ts.Close()

	result, err := client.CreateDestination(context.Background(), CreateDestinationRequest{
		Name:   "test-dest",
		Type:   "elasticsearch",
		Config: map[string]interface{}{"index_name": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "dst-123" {
		t.Errorf("expected ID dst-123, got %s", result.ID)
	}
}

func TestGetDestinationNotFound(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer ts.Close()

	result, err := client.GetDestination(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for 404, got %+v", result)
	}
}

func TestDeleteDestinationIdempotent(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer ts.Close()

	err := client.DeleteDestination(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for 404 delete, got: %v", err)
	}
}

// --- Workflow CRUD Tests ---

func TestCreateWorkflow(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Workflow{
			ID:            "wf-123",
			Name:          "test-wf",
			Sources:       []string{"src-1"},
			Destinations:  []string{"dst-1"},
			WorkflowType:  "custom",
			WorkflowNodes: []WorkflowNode{{Name: "partitioner", Type: "partition", Subtype: "vlm"}},
			Status:        "active",
			CreatedAt:     "2025-01-01T00:00:00Z",
			ReprocessAll:  true,
		})
	})
	defer ts.Close()

	reprocess := true
	result, err := client.CreateWorkflow(context.Background(), CreateWorkflowRequest{
		Name:          "test-wf",
		SourceID:      "src-1",
		DestinationID: "dst-1",
		WorkflowType:  "custom",
		WorkflowNodes: []WorkflowNode{{Name: "partitioner", Type: "partition", Subtype: "vlm"}},
		ReprocessAll:  &reprocess,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "wf-123" {
		t.Errorf("expected ID wf-123, got %s", result.ID)
	}
	if result.Status != "active" {
		t.Errorf("expected status active, got %s", result.Status)
	}
}

func TestGetWorkflowNotFound(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer ts.Close()

	result, err := client.GetWorkflow(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for 404, got %+v", result)
	}
}

func TestDeleteWorkflowIdempotent(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer ts.Close()

	err := client.DeleteWorkflow(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for 404 delete, got: %v", err)
	}
}

func TestRunWorkflow(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf-123/run" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	defer ts.Close()

	err := client.RunWorkflow(context.Background(), "wf-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Job Tests ---

func TestGetJob(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/job-123" {
			t.Errorf("expected /jobs/job-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Job{
			ID:           "job-123",
			WorkflowID:   "wf-123",
			WorkflowName: "test-workflow",
			Status:       "COMPLETED",
			CreatedAt:    "2025-01-01T00:00:00Z",
			JobType:      "ephemeral",
		})
	})
	defer ts.Close()

	result, err := client.GetJob(context.Background(), "job-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %s", result.Status)
	}
}

func TestListJobsWithFilter(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("workflow_id"); got != "wf-123" {
			t.Errorf("expected workflow_id=wf-123, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]Job{
			{ID: "job-1", WorkflowID: "wf-123", Status: "COMPLETED", CreatedAt: "2025-01-01T00:00:00Z"},
		})
	})
	defer ts.Close()

	results, err := client.ListJobs(context.Background(), "wf-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 job, got %d", len(results))
	}
}

func TestListJobsWithoutFilter(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})
	defer ts.Close()

	results, err := client.ListJobs(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(results))
	}
}

// --- Error Handling Tests ---

func TestAPIErrorReturnsStatusAndBody(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"invalid request"}`))
	})
	defer ts.Close()

	_, err := client.CreateSource(context.Background(), CreateSourceRequest{Name: "bad"})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAPIServerError(t *testing.T) {
	ts, client := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail":"internal error"}`))
	})
	defer ts.Close()

	_, err := client.GetSource(context.Background(), "src-123")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
