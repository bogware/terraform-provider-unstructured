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

var _ datasource.DataSource = &TemplateDataSource{}

// NewTemplateDataSource returns a new template data source.
func NewTemplateDataSource() datasource.DataSource {
	return &TemplateDataSource{}
}

// TemplateDataSource defines the data source implementation.
type TemplateDataSource struct {
	client *UnstructuredClient
}

// TemplateDataSourceModel describes the data source data model.
type TemplateDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	WorkflowType  types.String `tfsdk:"workflow_type"`
	WorkflowNodes types.String `tfsdk:"workflow_nodes"`
}

func (d *TemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (d *TemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an available workflow template. " +
			"Templates provide pre-configured workflow node pipelines that can be used when creating " +
			"template-based workflows via the `template_id` attribute on `unstructured_workflow`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the template.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the template.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of what the template does.",
			},
			"workflow_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The workflow type for this template.",
			},
			"workflow_nodes": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A JSON-encoded array of workflow node definitions in this template.",
			},
		},
	}
}

func (d *TemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TemplateDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := d.client.GetTemplate(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read template: %s", err))
		return
	}
	if tmpl == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Template %s not found.", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(tmpl.Name)

	if tmpl.Description != "" {
		data.Description = types.StringValue(tmpl.Description)
	} else {
		data.Description = types.StringNull()
	}

	if tmpl.WorkflowType != "" {
		data.WorkflowType = types.StringValue(tmpl.WorkflowType)
	} else {
		data.WorkflowType = types.StringNull()
	}

	if len(tmpl.WorkflowNodes) > 0 {
		nodesJSON, err := json.Marshal(tmpl.WorkflowNodes)
		if err != nil {
			resp.Diagnostics.AddError("Nodes Error", fmt.Sprintf("Unable to marshal workflow nodes: %s", err))
			return
		}
		data.WorkflowNodes = types.StringValue(string(nodesJSON))
	} else {
		data.WorkflowNodes = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
