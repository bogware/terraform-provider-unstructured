// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WorkflowDataSource{}

// NewWorkflowDataSource returns a new workflow data source.
func NewWorkflowDataSource() datasource.DataSource {
	return &WorkflowDataSource{}
}

// WorkflowDataSource defines the data source implementation.
type WorkflowDataSource struct {
	client *UnstructuredClient
}

// WorkflowDataSourceModel describes the data source data model.
type WorkflowDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	SourceID      types.String `tfsdk:"source_id"`
	DestinationID types.String `tfsdk:"destination_id"`
	WorkflowType  types.String `tfsdk:"workflow_type"`
	WorkflowNodes types.String `tfsdk:"workflow_nodes"`
	Schedule      types.String `tfsdk:"schedule"`
	ReprocessAll  types.Bool   `tfsdk:"reprocess_all"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (d *WorkflowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *WorkflowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing workflow.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the workflow.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the workflow.",
			},
			"source_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the source connector.",
			},
			"destination_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the destination connector.",
			},
			"workflow_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of workflow.",
			},
			"workflow_nodes": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A JSON-encoded array of workflow node definitions.",
			},
			"schedule": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The cron schedule expression for the workflow.",
			},
			"reprocess_all": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether all documents are reprocessed on each run.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the workflow.",
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

func (d *WorkflowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*UnstructuredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *UnstructuredClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkflowDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wf, err := d.client.GetWorkflow(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflow: %s", err))
		return
	}
	if wf == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Workflow with ID %q not found.", data.ID.ValueString()))
		return
	}

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
		nodesJSON, err := json.Marshal(wf.WorkflowNodes)
		if err != nil {
			resp.Diagnostics.AddError("Nodes Error", fmt.Sprintf("Unable to marshal workflow nodes: %s", err))
			return
		}
		data.WorkflowNodes = types.StringValue(string(nodesJSON))
	} else {
		data.WorkflowNodes = types.StringNull()
	}

	if wf.Schedule != nil && len(wf.Schedule.CrontabEntries) > 0 {
		data.Schedule = types.StringValue(wf.Schedule.CrontabEntries[0].CronExpression)
	} else {
		data.Schedule = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
