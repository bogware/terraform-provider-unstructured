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

var _ resource.Resource = &DestinationResource{}
var _ resource.ResourceWithImportState = &DestinationResource{}

// NewDestinationResource returns a new destination connector resource.
func NewDestinationResource() resource.Resource {
	return &DestinationResource{}
}

// DestinationResource defines the resource implementation.
type DestinationResource struct {
	client *UnstructuredClient
}

// DestinationResourceModel describes the resource data model.
type DestinationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *DestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a destination connector in the Unstructured platform. " +
			"Destination connectors define where processed data is delivered (e.g., S3, Pinecone, Elasticsearch).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the destination connector.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A unique name for this destination connector.",
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The type of destination connector. Valid values include: " +
					"`azure`, `astradb`, `azure_ai_search`, `couchbase`, `databricks_volumes`, " +
					"`databricks_volume_delta_tables`, `delta_table`, `elasticsearch`, `gcs`, " +
					"`kafka-cloud`, `milvus`, `mongodb`, `motherduck`, `neo4j`, `onedrive`, " +
					"`opensearch`, `pinecone`, `postgres`, `redis`, `qdrant-cloud`, `s3`, " +
					"`snowflake`, `teradata`, `weaviate-cloud`, `ibm_watsonx_s3`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				MarkdownDescription: "A JSON-encoded string containing the connector-specific configuration. " +
					"The structure depends on the connector type. For example, an S3 destination might use: " +
					"`jsonencode({remote_url = \"s3://my-bucket/output\", key = \"...\", secret = \"...\"})`.",
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

func (r *DestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DestinationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
		resp.Diagnostics.AddError("Invalid Config", fmt.Sprintf("Unable to parse config JSON: %s", err))
		return
	}

	apiReq := CreateDestinationRequest{
		Name:   data.Name.ValueString(),
		Type:   data.Type.ValueString(),
		Config: config,
	}

	dest, err := r.client.CreateDestination(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create destination connector: %s", err))
		return
	}

	data.ID = types.StringValue(dest.ID)
	data.CreatedAt = types.StringValue(dest.CreatedAt)
	if dest.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*dest.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	tflog.Trace(ctx, "created destination connector", map[string]interface{}{"id": dest.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DestinationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := r.client.GetDestination(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read destination connector: %s", err))
		return
	}
	if dest == nil {
		resp.State.RemoveResource(ctx)
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

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DestinationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
		resp.Diagnostics.AddError("Invalid Config", fmt.Sprintf("Unable to parse config JSON: %s", err))
		return
	}

	apiReq := CreateDestinationRequest{
		Name:   data.Name.ValueString(),
		Type:   data.Type.ValueString(),
		Config: config,
	}

	dest, err := r.client.UpdateDestination(ctx, data.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update destination connector: %s", err))
		return
	}

	data.CreatedAt = types.StringValue(dest.CreatedAt)
	if dest.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*dest.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	tflog.Trace(ctx, "updated destination connector", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DestinationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDestination(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete destination connector: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted destination connector", map[string]interface{}{"id": data.ID.ValueString()})
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
