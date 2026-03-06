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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SourceResource{}
var _ resource.ResourceWithImportState = &SourceResource{}

// NewSourceResource returns a new source connector resource.
func NewSourceResource() resource.Resource {
	return &SourceResource{}
}

// SourceResource defines the resource implementation.
type SourceResource struct {
	client *UnstructuredClient
}

// SourceResourceModel describes the resource data model.
type SourceResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *SourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (r *SourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a source connector in the Unstructured platform. " +
			"Source connectors define where data is ingested from (e.g., S3, Azure Blob Storage, Google Drive).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the source connector.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A unique name for this source connector. Changing this forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The type of source connector. Valid values include: " +
					"`azure`, `box`, `confluence`, `couchbase`, `databricks_volumes`, `dropbox`, " +
					"`elasticsearch`, `gcs`, `google_drive`, `kafka-cloud`, `mongodb`, `onedrive`, " +
					"`opensearch`, `outlook`, `postgres`, `s3`, `salesforce`, `sharepoint`, " +
					"`slack`, `snowflake`, `teradata`, `jira`, `zendesk`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				MarkdownDescription: "A JSON-encoded string containing the connector-specific configuration. " +
					"The structure depends on the connector type. For example, an S3 source might use: " +
					"`jsonencode({remote_url = \"s3://my-bucket/path\", key = \"...\", secret = \"...\"})`.",
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

func (r *SourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
		resp.Diagnostics.AddError("Invalid Config", fmt.Sprintf("Unable to parse config JSON: %s", err))
		return
	}

	apiReq := CreateSourceRequest{
		Name:   data.Name.ValueString(),
		Type:   data.Type.ValueString(),
		Config: config,
	}

	source, err := r.client.CreateSource(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create source connector: %s", err))
		return
	}

	data.ID = types.StringValue(source.ID)
	data.CreatedAt = types.StringValue(source.CreatedAt)
	if source.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*source.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	tflog.Trace(ctx, "created source connector", map[string]interface{}{"id": source.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.GetSource(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source connector: %s", err))
		return
	}
	if source == nil {
		resp.State.RemoveResource(ctx)
		return
	}

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

func (r *SourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
		resp.Diagnostics.AddError("Invalid Config", fmt.Sprintf("Unable to parse config JSON: %s", err))
		return
	}

	apiReq := UpdateSourceRequest{
		Config: config,
	}

	source, err := r.client.UpdateSource(ctx, data.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update source connector: %s", err))
		return
	}

	data.CreatedAt = types.StringValue(source.CreatedAt)
	if source.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*source.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	tflog.Trace(ctx, "updated source connector", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSource(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete source connector: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted source connector", map[string]interface{}{"id": data.ID.ValueString()})
}

func (r *SourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
