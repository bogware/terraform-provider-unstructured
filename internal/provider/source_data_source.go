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

var _ datasource.DataSource = &SourceDataSource{}

// NewSourceDataSource returns a new source connector data source.
func NewSourceDataSource() datasource.DataSource {
	return &SourceDataSource{}
}

// SourceDataSource defines the data source implementation.
type SourceDataSource struct {
	client *UnstructuredClient
}

// SourceDataSourceModel describes the data source data model.
type SourceDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// exactlyOneOf returns true if exactly one of the provided bools is true.
func exactlyOneOf(vals ...bool) bool {
	count := 0
	for _, v := range vals {
		if v {
			count++
		}
	}
	return count == 1
}

func (d *SourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (d *SourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing source connector.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the source connector. Exactly one of `id` or `name` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the source connector. Exactly one of `id` or `name` must be set.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of source connector.",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The JSON-encoded connector configuration.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the source connector was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the source connector was last updated.",
			},
		},
	}
}

func (d *SourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && !data.Name.IsUnknown() && data.Name.ValueString() != ""

	if !exactlyOneOf(hasID, hasName) {
		resp.Diagnostics.AddError("Invalid Configuration", "Exactly one of `id` or `name` must be set.")
		return
	}

	var source *SourceConnector

	if hasID {
		s, err := d.client.GetSource(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source connector: %s", err))
			return
		}
		if s == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Source connector with id %q not found.", data.ID.ValueString()))
			return
		}
		source = s
	} else {
		sources, err := d.client.ListSources(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list source connectors: %s", err))
			return
		}
		for i := range sources {
			if sources[i].Name == data.Name.ValueString() {
				source = &sources[i]
				break
			}
		}
		if source == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Source connector with name %q not found.", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.StringValue(source.ID)
	data.Name = types.StringValue(source.Name)
	data.Type = types.StringValue(source.Type)

	configJSON, err := json.Marshal(source.Config)
	if err != nil {
		resp.Diagnostics.AddError("Config Error", fmt.Sprintf("Unable to marshal config: %s", err))
		return
	}
	data.Config = types.StringValue(string(configJSON))

	data.CreatedAt = types.StringValue(source.CreatedAt)
	if source.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*source.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
