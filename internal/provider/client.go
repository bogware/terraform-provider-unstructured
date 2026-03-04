// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultAPIURL = "https://platform.unstructuredapp.io/api/v1"

// UnstructuredClient is the HTTP client for the Unstructured API.
type UnstructuredClient struct {
	httpClient *http.Client
	apiKey     string
	apiURL     string
}

// NewUnstructuredClient creates a new API client.
func NewUnstructuredClient(apiKey, apiURL string) *UnstructuredClient {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &UnstructuredClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
		apiURL:     apiURL,
	}
}

func (c *UnstructuredClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	url := c.apiURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("unstructured-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// --- Source Connector Operations ---

// SourceConnector represents a source connector from the API.
type SourceConnector struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt *string                `json:"updated_at"`
}

// CreateSourceRequest is the request body for creating a source connector.
type CreateSourceRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// CreateSource creates a new source connector.
func (c *UnstructuredClient) CreateSource(ctx context.Context, req CreateSourceRequest) (*SourceConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/sources/", req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var source SourceConnector
	if err := json.Unmarshal(body, &source); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &source, nil
}

// GetSource retrieves a source connector by ID.
func (c *UnstructuredClient) GetSource(ctx context.Context, id string) (*SourceConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/sources/"+id, nil)
	if err != nil {
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var source SourceConnector
	if err := json.Unmarshal(body, &source); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &source, nil
}

// UpdateSource updates an existing source connector.
func (c *UnstructuredClient) UpdateSource(ctx context.Context, id string, req CreateSourceRequest) (*SourceConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/sources/"+id, req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var source SourceConnector
	if err := json.Unmarshal(body, &source); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &source, nil
}

// DeleteSource deletes a source connector by ID.
func (c *UnstructuredClient) DeleteSource(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/sources/"+id, nil)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}
	return nil
}

// ListSources lists all source connectors.
func (c *UnstructuredClient) ListSources(ctx context.Context) ([]SourceConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/sources/", nil)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var sources []SourceConnector
	if err := json.Unmarshal(body, &sources); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return sources, nil
}

// --- Destination Connector Operations ---

// DestinationConnector represents a destination connector from the API.
type DestinationConnector struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt *string                `json:"updated_at"`
}

// CreateDestinationRequest is the request body for creating a destination connector.
type CreateDestinationRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// CreateDestination creates a new destination connector.
func (c *UnstructuredClient) CreateDestination(ctx context.Context, req CreateDestinationRequest) (*DestinationConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/destinations/", req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var dest DestinationConnector
	if err := json.Unmarshal(body, &dest); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &dest, nil
}

// GetDestination retrieves a destination connector by ID.
func (c *UnstructuredClient) GetDestination(ctx context.Context, id string) (*DestinationConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/destinations/"+id, nil)
	if err != nil {
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var dest DestinationConnector
	if err := json.Unmarshal(body, &dest); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &dest, nil
}

// UpdateDestination updates an existing destination connector.
func (c *UnstructuredClient) UpdateDestination(ctx context.Context, id string, req CreateDestinationRequest) (*DestinationConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/destinations/"+id, req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var dest DestinationConnector
	if err := json.Unmarshal(body, &dest); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &dest, nil
}

// DeleteDestination deletes a destination connector by ID.
func (c *UnstructuredClient) DeleteDestination(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/destinations/"+id, nil)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}
	return nil
}

// ListDestinations lists all destination connectors.
func (c *UnstructuredClient) ListDestinations(ctx context.Context) ([]DestinationConnector, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/destinations/", nil)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var dests []DestinationConnector
	if err := json.Unmarshal(body, &dests); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return dests, nil
}

// --- Workflow Operations ---

// WorkflowNode represents a node in a workflow DAG.
type WorkflowNode struct {
	ID       *string                `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Subtype  string                 `json:"subtype"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// CrontabEntry represents a cron schedule entry.
type CrontabEntry struct {
	CronExpression string `json:"cron_expression"`
}

// WorkflowSchedule represents the schedule for a workflow.
type WorkflowSchedule struct {
	CrontabEntries []CrontabEntry `json:"crontab_entries"`
}

// Workflow represents a workflow from the API.
type Workflow struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Sources       []string          `json:"sources"`
	Destinations  []string          `json:"destinations"`
	WorkflowType  string            `json:"workflow_type"`
	WorkflowNodes []WorkflowNode    `json:"workflow_nodes"`
	Schedule      *WorkflowSchedule `json:"schedule,omitempty"`
	Status        string            `json:"status"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     *string           `json:"updated_at"`
	ReprocessAll  bool              `json:"reprocess_all"`
}

// CreateWorkflowRequest is the request body for creating a workflow.
type CreateWorkflowRequest struct {
	Name          string         `json:"name"`
	SourceID      string         `json:"source_id,omitempty"`
	DestinationID string         `json:"destination_id,omitempty"`
	WorkflowType  string         `json:"workflow_type"`
	WorkflowNodes []WorkflowNode `json:"workflow_nodes,omitempty"`
	TemplateID    string         `json:"template_id,omitempty"`
	Schedule      string         `json:"schedule,omitempty"`
	ReprocessAll  *bool          `json:"reprocess_all,omitempty"`
}

// UpdateWorkflowRequest is the request body for updating a workflow.
type UpdateWorkflowRequest struct {
	Name          string         `json:"name,omitempty"`
	SourceID      string         `json:"source_id,omitempty"`
	DestinationID string         `json:"destination_id,omitempty"`
	WorkflowNodes []WorkflowNode `json:"workflow_nodes,omitempty"`
	Schedule      string         `json:"schedule,omitempty"`
	ReprocessAll  *bool          `json:"reprocess_all,omitempty"`
}

// CreateWorkflow creates a new workflow.
func (c *UnstructuredClient) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*Workflow, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/workflows/", req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var wf Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &wf, nil
}

// GetWorkflow retrieves a workflow by ID.
func (c *UnstructuredClient) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/workflows/"+id, nil)
	if err != nil {
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var wf Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &wf, nil
}

// UpdateWorkflow updates an existing workflow.
func (c *UnstructuredClient) UpdateWorkflow(ctx context.Context, id string, req UpdateWorkflowRequest) (*Workflow, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/workflows/"+id, req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var wf Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &wf, nil
}

// DeleteWorkflow deletes a workflow by ID.
func (c *UnstructuredClient) DeleteWorkflow(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/workflows/"+id, nil)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}
	return nil
}

// RunWorkflow triggers a workflow run.
func (c *UnstructuredClient) RunWorkflow(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/workflows/"+id+"/run", nil)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}
	return nil
}

// ListWorkflows lists all workflows.
func (c *UnstructuredClient) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/workflows/", nil)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var wfs []Workflow
	if err := json.Unmarshal(body, &wfs); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return wfs, nil
}

// --- Job Operations ---

// Job represents a job from the API.
type Job struct {
	ID               string  `json:"id"`
	WorkflowID       string  `json:"workflow_id"`
	Status           string  `json:"status"`
	ProcessingStatus string  `json:"processing_status,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        *string `json:"updated_at"`
}

// GetJob retrieves a job by ID.
func (c *UnstructuredClient) GetJob(ctx context.Context, id string) (*Job, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/jobs/"+id, nil)
	if err != nil {
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var job Job
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &job, nil
}

// ListJobs lists jobs, optionally filtered by workflow ID.
func (c *UnstructuredClient) ListJobs(ctx context.Context, workflowID string) ([]Job, error) {
	path := "/jobs/"
	if workflowID != "" {
		path += "?workflow_id=" + workflowID
	}
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(body))
	}

	var jobs []Job
	if err := json.Unmarshal(body, &jobs); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return jobs, nil
}
