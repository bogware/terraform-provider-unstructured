// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JobDataSource{}

// NewJobDataSource returns a new job data source.
func NewJobDataSource() datasource.DataSource {
	return &JobDataSource{}
}

// JobDataSource defines the data source implementation.
type JobDataSource struct {
	client *UnstructuredClient
}

// JobDataSourceModel describes the data source data model.
type JobDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	WorkflowID       types.String `tfsdk:"workflow_id"`
	Status           types.String `tfsdk:"status"`
	ProcessingStatus types.String `tfsdk:"processing_status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (d *JobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *JobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a workflow job.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the job.",
			},
			"workflow_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the workflow that owns this job.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The status of the job. " +
					"Possible values: `SCHEDULED`, `IN_PROGRESS`, `COMPLETED`, `STOPPED`, `FAILED`.",
			},
			"processing_status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The detailed processing status of the job. " +
					"Possible values: `SCHEDULED`, `IN_PROGRESS`, `SUCCESS`, `COMPLETED_WITH_ERRORS`, `STOPPED`, `FAILED`.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the job was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the job was last updated.",
			},
		},
	}
}

func (d *JobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *JobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := d.client.GetJob(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read job: %s", err))
		return
	}
	if job == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Job %s not found.", data.ID.ValueString()))
		return
	}

	data.WorkflowID = types.StringValue(job.WorkflowID)
	data.Status = types.StringValue(job.Status)

	if job.ProcessingStatus != "" {
		data.ProcessingStatus = types.StringValue(job.ProcessingStatus)
	} else {
		data.ProcessingStatus = types.StringNull()
	}

	data.CreatedAt = types.StringValue(job.CreatedAt)
	if job.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*job.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
