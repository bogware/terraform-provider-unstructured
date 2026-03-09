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
	ID           types.String `tfsdk:"id"`
	WorkflowID   types.String `tfsdk:"workflow_id"`
	WorkflowName types.String `tfsdk:"workflow_name"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	Runtime      types.String `tfsdk:"runtime"`
	JobType      types.String `tfsdk:"job_type"`
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
			"workflow_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the workflow that owns this job.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The status of the job. " +
					"Possible values: `SCHEDULED`, `IN_PROGRESS`, `COMPLETED`, `STOPPED`, `FAILED`.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the job was created.",
			},
			"runtime": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The runtime duration of the job.",
			},
			"job_type": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The type of the job. " +
					"Possible values: `ephemeral`, `persistent`, `scheduled`, `template`.",
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
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Job with ID %q not found.", data.ID.ValueString()))
		return
	}

	data.WorkflowID = types.StringValue(job.WorkflowID)
	data.WorkflowName = types.StringValue(job.WorkflowName)
	data.Status = types.StringValue(job.Status)
	data.CreatedAt = types.StringValue(job.CreatedAt)

	if job.Runtime != nil {
		data.Runtime = types.StringValue(*job.Runtime)
	} else {
		data.Runtime = types.StringNull()
	}

	if job.JobType != "" {
		data.JobType = types.StringValue(job.JobType)
	} else {
		data.JobType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
