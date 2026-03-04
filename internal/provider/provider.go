// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure UnstructuredProvider satisfies various provider interfaces.
var _ provider.Provider = &UnstructuredProvider{}

// UnstructuredProvider defines the provider implementation.
type UnstructuredProvider struct {
	version string
}

// UnstructuredProviderModel describes the provider data model.
type UnstructuredProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}

func (p *UnstructuredProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unstructured"
	resp.Version = p.version
}

func (p *UnstructuredProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Unstructured provider manages resources in the [Unstructured](https://unstructured.io/) platform API. " +
			"It supports managing source connectors, destination connectors, and workflows for document processing pipelines.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authenticating with the Unstructured API. " +
					"Can also be set via the `UNSTRUCTURED_API_KEY` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Unstructured API. " +
					"Defaults to `https://platform.unstructuredapp.io/api/v1`. " +
					"Can also be set via the `UNSTRUCTURED_API_URL` environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *UnstructuredProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UnstructuredProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Defer to env vars when config values are unknown (e.g. during plan
	// with variables that depend on other resources).
	apiKey := os.Getenv("UNSTRUCTURED_API_KEY")
	if !data.APIKey.IsNull() && !data.APIKey.IsUnknown() {
		apiKey = data.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Unstructured API key must be set in the provider configuration or "+
				"via the UNSTRUCTURED_API_KEY environment variable.",
		)
		return
	}

	apiURL := os.Getenv("UNSTRUCTURED_API_URL")
	if !data.APIURL.IsNull() && !data.APIURL.IsUnknown() {
		apiURL = data.APIURL.ValueString()
	}

	client := NewUnstructuredClient(apiKey, apiURL, p.version)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *UnstructuredProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSourceResource,
		NewDestinationResource,
		NewWorkflowResource,
	}
}

func (p *UnstructuredProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSourceDataSource,
		NewDestinationDataSource,
		NewWorkflowDataSource,
		NewJobDataSource,
	}
}

// New returns a function that creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnstructuredProvider{
			version: version,
		}
	}
}
