---
page_title: "Unstructured Provider"
description: |-
  The Unstructured provider manages source connectors, destination connectors, and document processing workflows on the Unstructured platform.
---

# Unstructured Provider

The Unstructured provider manages resources on the [Unstructured](https://unstructured.io/) platform API. Use it to declaratively configure document processing pipelines that ingest data from cloud storage, partition and enrich documents, generate embeddings, and deliver results to vector databases and other destinations.

## Authentication

The provider requires an API key from the [Unstructured dashboard](https://platform.unstructuredapp.io/).

Set it via environment variable (recommended):

```bash
export UNSTRUCTURED_API_KEY="your-api-key"
```

Or configure it in the provider block:

```hcl
provider "unstructured" {
  api_key = var.unstructured_api_key
}
```

## Example Usage

{{tffile "examples/provider/provider.tf"}}

### Complete Pipeline Example

```hcl
# Source: where documents come from
resource "unstructured_source" "s3" {
  name = "company-docs"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://my-bucket/documents"
    key        = var.aws_access_key
    secret     = var.aws_secret_key
  })
}

# Destination: where processed data goes
resource "unstructured_destination" "pinecone" {
  name = "vector-index"
  type = "pinecone"
  config = jsonencode({
    api_key    = var.pinecone_api_key
    index_name = "documents"
  })
}

# Workflow: the processing pipeline
resource "unstructured_workflow" "pipeline" {
  name            = "doc-pipeline"
  source_id       = unstructured_source.s3.id
  destination_id  = unstructured_destination.pinecone.id
  workflow_type   = "custom"
  schedule        = "daily"
  reprocess_all   = false

  workflow_nodes = jsonencode([
    {
      name    = "partitioner"
      type    = "partition"
      subtype = "vlm"
    },
    {
      name    = "chunker"
      type    = "chunk"
      subtype = "chunk_by_title"
    },
    {
      name    = "embedder"
      type    = "embed"
      subtype = "openai"
      settings = {
        model_name = "text-embedding-3-small"
        api_key    = var.openai_api_key
      }
    }
  ])
}
```

{{ .SchemaMarkdown | trimspace }}
