// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func TestProviderMetadata(t *testing.T) {
	p := &UnstructuredProvider{version: "1.2.3"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "unstructured" {
		t.Errorf("expected type name 'unstructured', got %q", resp.TypeName)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	p := &UnstructuredProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", resp.Diagnostics)
	}

	s := resp.Schema
	validateAttribute(t, s, "api_key", true, false, true)
	validateAttribute(t, s, "api_url", true, false, false)
}

func validateAttribute(t *testing.T, s schema.Schema, name string, optional, required, sensitive bool) {
	t.Helper()
	attr, ok := s.Attributes[name]
	if !ok {
		t.Errorf("expected attribute %q to exist", name)
		return
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Errorf("expected attribute %q to be StringAttribute", name)
		return
	}
	if strAttr.Optional != optional {
		t.Errorf("attribute %q: expected Optional=%v, got %v", name, optional, strAttr.Optional)
	}
	if strAttr.Required != required {
		t.Errorf("attribute %q: expected Required=%v, got %v", name, required, strAttr.Required)
	}
	if strAttr.Sensitive != sensitive {
		t.Errorf("attribute %q: expected Sensitive=%v, got %v", name, sensitive, strAttr.Sensitive)
	}
}

func TestProviderResources(t *testing.T) {
	p := &UnstructuredProvider{}
	resources := p.Resources(context.Background())

	expected := 3 // source, destination, workflow
	if len(resources) != expected {
		t.Errorf("expected %d resources, got %d", expected, len(resources))
	}
}

func TestProviderDataSources(t *testing.T) {
	p := &UnstructuredProvider{}
	dataSources := p.DataSources(context.Background())

	expected := 4 // source, destination, workflow, job
	if len(dataSources) != expected {
		t.Errorf("expected %d data sources, got %d", expected, len(dataSources))
	}
}

func TestNewReturnsProviderFactory(t *testing.T) {
	factory := New("test-version")
	if factory == nil {
		t.Fatal("expected factory function, got nil")
	}

	p := factory()
	if p == nil {
		t.Fatal("expected provider instance, got nil")
	}
}
