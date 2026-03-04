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

var _ datasource.DataSource = &DestinationDataSource{}

// NewDestinationDataSource returns a new destination connector data source.
func NewDestinationDataSource() datasource.DataSource {
	return &DestinationDataSource{}
}

// DestinationDataSource defines the data source implementation.
type DestinationDataSource struct {
	client *UnstructuredClient
}

// DestinationDataSourceModel describes the data source data model.
type DestinationDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *DestinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (d *DestinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing destination connector.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the destination connector.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the destination connector.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of destination connector.",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The JSON-encoded connector configuration.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the destination connector was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the destination connector was last updated.",
			},
		},
	}
}

func (d *DestinationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DestinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DestinationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := d.client.GetDestination(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read destination connector: %s", err))
		return
	}
	if dest == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Destination connector %s not found.", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(dest.Name)
	data.Type = types.StringValue(dest.Type)

	configJSON, err := json.Marshal(dest.Config)
	if err != nil {
		resp.Diagnostics.AddError("Config Error", fmt.Sprintf("Unable to marshal config: %s", err))
		return
	}
	data.Config = types.StringValue(string(configJSON))

	data.CreatedAt = types.StringValue(dest.CreatedAt)
	if dest.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*dest.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
