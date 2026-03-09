// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderMetadata(t *testing.T) {
	p := &UnstructuredProvider{version: "1.2.3"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "unstructured" {
		t.Errorf("expected type name 'unstructured', got %q", resp.TypeName)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	p := &UnstructuredProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", resp.Diagnostics)
	}

	s := resp.Schema
	validateAttribute(t, s, "api_key", true, false, true)
	validateAttribute(t, s, "api_url", true, false, false)
}

func validateAttribute(t *testing.T, s schema.Schema, name string, optional, required, sensitive bool) {
	t.Helper()
	attr, ok := s.Attributes[name]
	if !ok {
		t.Errorf("expected attribute %q to exist", name)
		return
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Errorf("expected attribute %q to be StringAttribute", name)
		return
	}
	if strAttr.Optional != optional {
		t.Errorf("attribute %q: expected Optional=%v, got %v", name, optional, strAttr.Optional)
	}
	if strAttr.Required != required {
		t.Errorf("attribute %q: expected Required=%v, got %v", name, required, strAttr.Required)
	}
	if strAttr.Sensitive != sensitive {
		t.Errorf("attribute %q: expected Sensitive=%v, got %v", name, sensitive, strAttr.Sensitive)
	}
}

func TestProviderResources(t *testing.T) {
	p := &UnstructuredProvider{}
	resources := p.Resources(context.Background())

	expected := 3 // source, destination, workflow
	if len(resources) != expected {
		t.Errorf("expected %d resources, got %d", expected, len(resources))
	}
}

func TestProviderDataSources(t *testing.T) {
	p := &UnstructuredProvider{}
	dataSources := p.DataSources(context.Background())

	expected := 5 // source, destination, workflow, job, template
	if len(dataSources) != expected {
		t.Errorf("expected %d data sources, got %d", expected, len(dataSources))
	}
}

func TestNewReturnsProviderFactory(t *testing.T) {
	factory := New("test-version")
	if factory == nil {
		t.Fatal("expected factory function, got nil")
	}

	p := factory()
	if p == nil {
		t.Fatal("expected provider instance, got nil")
	}
}

// --- Resource Schema Tests ---

func TestSourceResourceSchema(t *testing.T) {
	r := NewSourceResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	requiredAttrs := []string{"name", "type", "config"}
	computedAttrs := []string{"id", "created_at", "updated_at"}

	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected computed attribute %q to exist", attr)
		}
	}
}

func TestDestinationResourceSchema(t *testing.T) {
	r := NewDestinationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	for _, attr := range []string{"id", "name", "type", "config", "created_at", "updated_at"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

func TestWorkflowResourceSchema(t *testing.T) {
	r := NewWorkflowResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	expectedAttrs := []string{
		"id", "name", "source_id", "destination_id", "workflow_type",
		"workflow_nodes", "template_id", "schedule", "reprocess_all",
		"status", "created_at", "updated_at",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

// --- Data Source Schema Tests ---

func TestSourceDataSourceSchema(t *testing.T) {
	d := NewSourceDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	for _, attr := range []string{"id", "name", "type", "config", "created_at", "updated_at"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

func TestDestinationDataSourceSchema(t *testing.T) {
	d := NewDestinationDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	for _, attr := range []string{"id", "name", "type", "config", "created_at", "updated_at"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

func TestWorkflowDataSourceSchema(t *testing.T) {
	d := NewWorkflowDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	expectedAttrs := []string{
		"id", "name", "source_id", "destination_id", "workflow_type",
		"workflow_nodes", "schedule", "reprocess_all", "status",
		"created_at", "updated_at",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

func TestJobDataSourceSchema(t *testing.T) {
	d := NewJobDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	for _, attr := range []string{"id", "workflow_id", "workflow_name", "status", "created_at", "runtime", "job_type"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

func TestTemplateDataSourceSchema(t *testing.T) {
	d := NewTemplateDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	for _, attr := range []string{"id", "name", "description", "workflow_type", "workflow_nodes"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
	}
}

// --- Resource/Data Source Metadata Tests ---

func TestSourceResourceMetadata(t *testing.T) {
	r := NewSourceResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_source" {
		t.Errorf("expected type name 'unstructured_source', got %q", resp.TypeName)
	}
}

func TestDestinationResourceMetadata(t *testing.T) {
	r := NewDestinationResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_destination" {
		t.Errorf("expected type name 'unstructured_destination', got %q", resp.TypeName)
	}
}

func TestWorkflowResourceMetadata(t *testing.T) {
	r := NewWorkflowResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_workflow" {
		t.Errorf("expected type name 'unstructured_workflow', got %q", resp.TypeName)
	}
}

func TestSourceDataSourceMetadata(t *testing.T) {
	d := NewSourceDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_source" {
		t.Errorf("expected type name 'unstructured_source', got %q", resp.TypeName)
	}
}

func TestDestinationDataSourceMetadata(t *testing.T) {
	d := NewDestinationDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_destination" {
		t.Errorf("expected type name 'unstructured_destination', got %q", resp.TypeName)
	}
}

func TestWorkflowDataSourceMetadata(t *testing.T) {
	d := NewWorkflowDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_workflow" {
		t.Errorf("expected type name 'unstructured_workflow', got %q", resp.TypeName)
	}
}

func TestJobDataSourceMetadata(t *testing.T) {
	d := NewJobDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_job" {
		t.Errorf("expected type name 'unstructured_job', got %q", resp.TypeName)
	}
}

func TestTemplateDataSourceMetadata(t *testing.T) {
	d := NewTemplateDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unstructured"}, resp)

	if resp.TypeName != "unstructured_template" {
		t.Errorf("expected type name 'unstructured_template', got %q", resp.TypeName)
	}
}

// --- mapWorkflowToState Tests ---

func TestMapWorkflowToStateBasic(t *testing.T) {
	r := &WorkflowResource{}
	data := &WorkflowResourceModel{}
	updatedAt := "2025-06-01T00:00:00Z"

	wf := &Workflow{
		ID:            "wf-123",
		Name:          "test",
		Sources:       []string{"src-1"},
		Destinations:  []string{"dst-1"},
		WorkflowType:  "custom",
		WorkflowNodes: []WorkflowNode{{Name: "p", Type: "partition", Subtype: "vlm"}},
		Schedule: &WorkflowSchedule{
			CrontabEntries: []CrontabEntry{{CronExpression: "0 * * * *"}},
		},
		Status:       "active",
		CreatedAt:    "2025-01-01T00:00:00Z",
		UpdatedAt:    &updatedAt,
		ReprocessAll: true,
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	if data.ID.ValueString() != "wf-123" {
		t.Errorf("expected ID wf-123, got %s", data.ID.ValueString())
	}
	if data.SourceID.ValueString() != "src-1" {
		t.Errorf("expected source_id src-1, got %s", data.SourceID.ValueString())
	}
	if data.DestinationID.ValueString() != "dst-1" {
		t.Errorf("expected destination_id dst-1, got %s", data.DestinationID.ValueString())
	}
	if data.Schedule.ValueString() != "0 * * * *" {
		t.Errorf("expected schedule '0 * * * *', got %s", data.Schedule.ValueString())
	}
	if !data.ReprocessAll.ValueBool() {
		t.Error("expected reprocess_all true")
	}
	if data.UpdatedAt.ValueString() != "2025-06-01T00:00:00Z" {
		t.Errorf("expected updated_at, got %s", data.UpdatedAt.ValueString())
	}

	// Verify workflow_nodes is valid JSON.
	var nodes []WorkflowNode
	if err := json.Unmarshal([]byte(data.WorkflowNodes.ValueString()), &nodes); err != nil {
		t.Fatalf("workflow_nodes is not valid JSON: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "p" {
		t.Errorf("unexpected nodes: %+v", nodes)
	}
}

func TestMapWorkflowToStateEmptyFields(t *testing.T) {
	r := &WorkflowResource{}
	data := &WorkflowResourceModel{}

	wf := &Workflow{
		ID:           "wf-456",
		Name:         "empty",
		WorkflowType: "template",
		Status:       "inactive",
		CreatedAt:    "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	if !data.SourceID.IsNull() {
		t.Error("expected source_id to be null")
	}
	if !data.DestinationID.IsNull() {
		t.Error("expected destination_id to be null")
	}
	if !data.WorkflowNodes.IsNull() {
		t.Error("expected workflow_nodes to be null")
	}
	if !data.Schedule.IsNull() {
		t.Error("expected schedule to be null")
	}
	if !data.UpdatedAt.IsNull() {
		t.Error("expected updated_at to be null")
	}
}

func TestMapWorkflowToStatePreservesTemplateID(t *testing.T) {
	r := &WorkflowResource{}
	data := &WorkflowResourceModel{
		TemplateID: types.StringValue("tmpl-existing"),
	}

	wf := &Workflow{
		ID:           "wf-789",
		Name:         "tmpl-wf",
		WorkflowType: "template",
		Status:       "active",
		CreatedAt:    "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	// TemplateID should be preserved from existing state.
	if data.TemplateID.ValueString() != "tmpl-existing" {
		t.Errorf("expected template_id preserved as 'tmpl-existing', got %s", data.TemplateID.ValueString())
	}
}

func TestMapWorkflowToStateUnknownTemplateIDBecomesNull(t *testing.T) {
	r := &WorkflowResource{}
	data := &WorkflowResourceModel{
		TemplateID: types.StringUnknown(),
	}

	wf := &Workflow{
		ID:           "wf-789",
		Name:         "tmpl-wf",
		WorkflowType: "template",
		Status:       "active",
		CreatedAt:    "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	if !data.TemplateID.IsNull() {
		t.Error("expected unknown template_id to become null")
	}
}

func TestMapWorkflowToStatePreservesSchedule(t *testing.T) {
	r := &WorkflowResource{}
	// Simulate Create/Update: schedule already has the user's human-readable value.
	data := &WorkflowResourceModel{
		Schedule: types.StringValue("daily"),
	}

	wf := &Workflow{
		ID:           "wf-123",
		Name:         "scheduled",
		WorkflowType: "custom",
		Schedule: &WorkflowSchedule{
			CrontabEntries: []CrontabEntry{{CronExpression: "0 0 * * *"}},
		},
		Status:    "active",
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	// Schedule should be preserved as user-provided "daily", not overwritten with cron.
	if data.Schedule.ValueString() != "daily" {
		t.Errorf("expected schedule preserved as 'daily', got %s", data.Schedule.ValueString())
	}
}

func TestMapWorkflowToStateScheduleNullUsesCron(t *testing.T) {
	r := &WorkflowResource{}
	// Simulate Read/Import: schedule is null (unknown state).
	data := &WorkflowResourceModel{}

	wf := &Workflow{
		ID:           "wf-123",
		Name:         "imported",
		WorkflowType: "custom",
		Schedule: &WorkflowSchedule{
			CrontabEntries: []CrontabEntry{{CronExpression: "0 0 * * *"}},
		},
		Status:    "active",
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	// When schedule was null, it should be populated from API response.
	if data.Schedule.ValueString() != "0 0 * * *" {
		t.Errorf("expected schedule '0 0 * * *', got %s", data.Schedule.ValueString())
	}
}

func TestMapWorkflowToStateStripsNodeIDs(t *testing.T) {
	r := &WorkflowResource{}
	data := &WorkflowResourceModel{}

	nodeID := "node-server-generated-id"
	wf := &Workflow{
		ID:           "wf-123",
		Name:         "test",
		WorkflowType: "custom",
		WorkflowNodes: []WorkflowNode{
			{ID: &nodeID, Name: "partitioner", Type: "partition", Subtype: "vlm"},
		},
		Status:    "active",
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	r.mapWorkflowToState(context.Background(), wf, data)

	// Verify node IDs are stripped from state.
	var nodes []WorkflowNode
	if err := json.Unmarshal([]byte(data.WorkflowNodes.ValueString()), &nodes); err != nil {
		t.Fatalf("workflow_nodes is not valid JSON: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].ID != nil {
		t.Error("expected node ID to be stripped (nil)")
	}
	if nodes[0].Name != "partitioner" {
		t.Errorf("expected node name 'partitioner', got %s", nodes[0].Name)
	}
}
