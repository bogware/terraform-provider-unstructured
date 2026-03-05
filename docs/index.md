---
page_title: "Unstructured Provider"
description: |-
  The Unstructured provider manages resources in the Unstructured platform for document processing pipelines.
---

# Unstructured Provider

The Unstructured provider manages resources in the [Unstructured](https://unstructured.io/) platform API. It supports managing source connectors, destination connectors, and workflows for document processing pipelines.

## Authentication

The provider requires an API key to authenticate with the Unstructured platform. You can obtain one from the [Unstructured dashboard](https://platform.unstructuredapp.io/).

The API key can be provided via the provider configuration block or via the `UNSTRUCTURED_API_KEY` environment variable.

## Example Usage

{{tffile "examples/provider/provider.tf"}}

{{ .SchemaMarkdown | trimspace }}
