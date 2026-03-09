// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &WorkflowResource{}
var _ resource.ResourceWithImportState = &WorkflowResource{}

// NewWorkflowResource returns a new workflow resource.
func NewWorkflowResource() resource.Resource {
	return &WorkflowResource{}
}

// WorkflowResource defines the resource implementation.
type WorkflowResource struct {
	client *UnstructuredClient
}

// WorkflowResourceModel describes the resource data model.
type WorkflowResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	SourceID      types.String `tfsdk:"source_id"`
	DestinationID types.String `tfsdk:"destination_id"`
	WorkflowType  types.String `tfsdk:"workflow_type"`
	WorkflowNodes types.String `tfsdk:"workflow_nodes"`
	TemplateID    types.String `tfsdk:"template_id"`
	Schedule      types.String `tfsdk:"schedule"`
	ReprocessAll  types.Bool   `tfsdk:"reprocess_all"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (r *WorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow in the Unstructured platform. " +
			"Workflows define how data is processed, connecting source connectors to destination connectors " +
			"through a configurable pipeline of partitioning, chunking, enrichment, and embedding steps.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the workflow.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A unique name for this workflow.",
			},
			"source_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the source connector. Required for remote sources. Omit for local sources.",
			},
			"destination_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the destination connector. Required for remote destinations. Omit for local destinations.",
			},
			"workflow_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The type of workflow. Use `custom` for workflows with manually specified nodes, " +
					"or `template` for template-based workflows.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workflow_nodes": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "A JSON-encoded array of workflow node definitions for `custom` type workflows. " +
					"Each node has `name`, `type`, `subtype`, and optional `settings` fields. " +
					"Example: `jsonencode([{name = \"partitioner\", type = \"partition\", subtype = \"vlm\", settings = {...}}])`.",
			},
			"template_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The template ID to use for `template` type workflows. " +
					"Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The schedule for recurring workflow runs. Supported values include: " +
					"`every 15 minutes`, `every hour`, `every 2 hours`, `every 4 hours`, `every 6 hours`, " +
					"`every 8 hours`, `every 10 hours`, `every 12 hours`, `daily`, `weekly`, `monthly`.",
			},
			"reprocess_all": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to reprocess all documents on each run. Defaults to `true`.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the workflow (`active` or `inactive`).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the workflow was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the workflow was last updated.",
			},
		},
	}
}

func (r *WorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*UnstructuredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *UnstructuredClient, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := CreateWorkflowRequest{
		Name:         data.Name.ValueString(),
		WorkflowType: data.WorkflowType.ValueString(),
	}

	if !data.SourceID.IsNull() && !data.SourceID.IsUnknown() {
		apiReq.SourceID = data.SourceID.ValueString()
	}
	if !data.DestinationID.IsNull() && !data.DestinationID.IsUnknown() {
		apiReq.DestinationID = data.DestinationID.ValueString()
	}
	if !data.TemplateID.IsNull() && !data.TemplateID.IsUnknown() {
		apiReq.TemplateID = data.TemplateID.ValueString()
	}
	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		apiReq.Schedule = data.Schedule.ValueString()
	}
	if !data.ReprocessAll.IsNull() && !data.ReprocessAll.IsUnknown() {
		v := data.ReprocessAll.ValueBool()
		apiReq.ReprocessAll = &v
	}

	if !data.WorkflowNodes.IsNull() && !data.WorkflowNodes.IsUnknown() {
		var nodes []WorkflowNode
		if err := json.Unmarshal([]byte(data.WorkflowNodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddError("Invalid Workflow Nodes", fmt.Sprintf("Unable to parse workflow_nodes JSON: %s", err))
			return
		}
		apiReq.WorkflowNodes = nodes
	}

	wf, err := r.client.CreateWorkflow(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create workflow: %s", err))
		return
	}

	r.mapWorkflowToState(ctx, wf, &data)
	tflog.Trace(ctx, "created workflow", map[string]interface{}{"id": wf.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wf, err := r.client.GetWorkflow(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflow: %s", err))
		return
	}
	if wf == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapWorkflowToState(ctx, wf, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := UpdateWorkflowRequest{
		Name: data.Name.ValueString(),
	}

	if !data.SourceID.IsNull() && !data.SourceID.IsUnknown() {
		apiReq.SourceID = data.SourceID.ValueString()
	}
	if !data.DestinationID.IsNull() && !data.DestinationID.IsUnknown() {
		apiReq.DestinationID = data.DestinationID.ValueString()
	}
	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		apiReq.Schedule = data.Schedule.ValueString()
	}
	if !data.ReprocessAll.IsNull() && !data.ReprocessAll.IsUnknown() {
		v := data.ReprocessAll.ValueBool()
		apiReq.ReprocessAll = &v
	}

	if !data.WorkflowNodes.IsNull() && !data.WorkflowNodes.IsUnknown() {
		var nodes []WorkflowNode
		if err := json.Unmarshal([]byte(data.WorkflowNodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddError("Invalid Workflow Nodes", fmt.Sprintf("Unable to parse workflow_nodes JSON: %s", err))
			return
		}
		apiReq.WorkflowNodes = nodes
	}

	wf, err := r.client.UpdateWorkflow(ctx, data.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update workflow: %s", err))
		return
	}

	r.mapWorkflowToState(ctx, wf, &data)
	tflog.Trace(ctx, "updated workflow", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorkflow(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete workflow: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted workflow", map[string]interface{}{"id": data.ID.ValueString()})
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *WorkflowResource) mapWorkflowToState(ctx context.Context, wf *Workflow, data *WorkflowResourceModel) {
	data.ID = types.StringValue(wf.ID)
	data.Name = types.StringValue(wf.Name)
	data.WorkflowType = types.StringValue(wf.WorkflowType)
	data.Status = types.StringValue(wf.Status)
	data.ReprocessAll = types.BoolValue(wf.ReprocessAll)
	data.CreatedAt = types.StringValue(wf.CreatedAt)

	if wf.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*wf.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	if len(wf.Sources) > 0 {
		data.SourceID = types.StringValue(wf.Sources[0])
	} else {
		data.SourceID = types.StringNull()
	}

	if len(wf.Destinations) > 0 {
		data.DestinationID = types.StringValue(wf.Destinations[0])
	} else {
		data.DestinationID = types.StringNull()
	}

	if len(wf.WorkflowNodes) > 0 {
		// Strip server-generated node IDs before storing in state to
		// avoid perpetual diffs with the user's input JSON.
		cleaned := make([]WorkflowNode, len(wf.WorkflowNodes))
		copy(cleaned, wf.WorkflowNodes)
		for i := range cleaned {
			cleaned[i].ID = nil
		}
		nodesJSON, err := json.Marshal(cleaned)
		if err != nil {
			tflog.Error(ctx, "failed to marshal workflow nodes", map[string]interface{}{"error": err.Error()})
			data.WorkflowNodes = types.StringNull()
		} else {
			data.WorkflowNodes = types.StringValue(string(nodesJSON))
		}
	} else {
		data.WorkflowNodes = types.StringNull()
	}

	// Schedule: the API accepts human-readable values ("daily", "every hour")
	// but returns cron expressions. Preserve the user's plan value during
	// Create/Update to avoid perpetual diffs. Only use the API cron value
	// during Read (when data.Schedule is null/unknown, e.g. import).
	if data.Schedule.IsNull() || data.Schedule.IsUnknown() {
		if wf.Schedule != nil && len(wf.Schedule.CrontabEntries) > 0 {
			data.Schedule = types.StringValue(wf.Schedule.CrontabEntries[0].CronExpression)
		} else {
			data.Schedule = types.StringNull()
		}
	}

	// TemplateID is write-only (used at creation time); preserve the
	// current state value. On import it will be null.
	if data.TemplateID.IsUnknown() {
		data.TemplateID = types.StringNull()
	}
}
