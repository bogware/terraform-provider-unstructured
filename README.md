# terraform-provider-unstructured

A Terraform provider for managing resources on the [Unstructured](https://unstructured.io/) platform. This provider enables you to declaratively manage source connectors, destination connectors, and document processing workflows through Infrastructure as Code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.13
- [Go](https://golang.org/doc/install) >= 1.24 (to build the provider plugin)
- An [Unstructured Platform](https://platform.unstructuredapp.io/) account and API key

## Installation

### From the Terraform Registry

```hcl
terraform {
  required_providers {
    unstructured = {
      source  = "bogware/unstructured"
      version = "~> 0.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/bogware/terraform-provider-unstructured.git
cd terraform-provider-unstructured
make install
```

## Authentication

The provider requires an API key to authenticate with the Unstructured platform. You can obtain one from the [Unstructured dashboard](https://platform.unstructuredapp.io/).

Configure via the provider block:

```hcl
provider "unstructured" {
  api_key = var.unstructured_api_key
}
```

Or via environment variables (recommended for CI/CD):

```bash
export UNSTRUCTURED_API_KEY="your-api-key"
export UNSTRUCTURED_API_URL="https://platform.unstructuredapp.io/api/v1"  # optional
```

> **Note:** `UNSTRUCTURED_API_URL` only needs to be set if you are using a non-default API endpoint.

## Usage

### Source Connector

Source connectors define where your documents are ingested from.

```hcl
resource "unstructured_source" "s3_docs" {
  name = "s3-documents"
  type = "s3"
  config = jsonencode({
    remote_url    = "s3://my-bucket/documents/"
    key           = var.aws_access_key
    secret        = var.aws_secret_key
    recursive     = true
  })
}
```

### Destination Connector

Destination connectors define where processed data is delivered.

```hcl
resource "unstructured_destination" "pinecone" {
  name = "pinecone-index"
  type = "pinecone"
  config = jsonencode({
    api_key    = var.pinecone_api_key
    index_name = "documents"
    environment = "us-east-1-aws"
  })
}
```

### Workflow

Workflows connect sources to destinations through a configurable processing pipeline.

```hcl
resource "unstructured_workflow" "pipeline" {
  name            = "document-pipeline"
  source_id       = unstructured_source.s3_docs.id
  destination_id  = unstructured_destination.pinecone.id
  workflow_type   = "custom"
  schedule        = "daily"
  reprocess_all   = false

  workflow_nodes = jsonencode([
    {
      name    = "partitioner"
      type    = "partition"
      subtype = "vlm"
      settings = {
        provider   = "anthropic"
        model      = "claude-sonnet-4-20250514"
      }
    },
    {
      name    = "chunker"
      type    = "chunk"
      subtype = "chunk_by_title"
      settings = {
        multipage_sections    = true
        combine_text_under_n_chars = 500
        max_characters        = 1500
      }
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

### Data Sources

Read existing resources for reference or cross-stack usage:

```hcl
data "unstructured_source" "existing" {
  id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

data "unstructured_destination" "existing" {
  id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

data "unstructured_workflow" "existing" {
  id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

data "unstructured_job" "latest" {
  id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

### Import

All managed resources support `terraform import`:

```bash
terraform import unstructured_source.s3_docs <source-id>
terraform import unstructured_destination.pinecone <destination-id>
terraform import unstructured_workflow.pipeline <workflow-id>
```

## Resources Reference

| Resource | Description |
|----------|-------------|
| `unstructured_source` | Manages a source connector (S3, Azure Blob, GCS, OneDrive, Sharepoint, SFTP, and 15+ more) |
| `unstructured_destination` | Manages a destination connector (Pinecone, Elasticsearch, Weaviate, S3, PostgreSQL, and 20+ more) |
| `unstructured_workflow` | Manages a processing workflow connecting sources to destinations |

## Data Sources Reference

| Data Source | Description |
|-------------|-------------|
| `unstructured_source` | Reads an existing source connector by ID |
| `unstructured_destination` | Reads an existing destination connector by ID |
| `unstructured_workflow` | Reads an existing workflow by ID |
| `unstructured_job` | Reads a job execution by ID |

## Supported Connector Types

### Sources

`azure`, `confluence`, `couchbase`, `databricks_volumes`, `delta_table`, `elasticsearch`, `gcs`, `google_drive`, `kafka-cloud`, `mongodb`, `motherduck`, `onedrive`, `opensearch`, `outlook`, `postgres`, `s3`, `salesforce`, `sftp`, `sharepoint`, `snowflake`, `teradata`, `ibm_watsonx_s3`

### Destinations

`azure`, `astradb`, `azure_ai_search`, `couchbase`, `databricks_volumes`, `databricks_volume_delta_tables`, `delta_table`, `elasticsearch`, `gcs`, `kafka-cloud`, `milvus`, `mongodb`, `motherduck`, `neo4j`, `onedrive`, `opensearch`, `pinecone`, `postgres`, `redis`, `qdrant-cloud`, `s3`, `snowflake`, `teradata`, `weaviate-cloud`, `ibm_watsonx_s3`

## Development

### Prerequisites

- Go >= 1.24
- GNU Make
- [golangci-lint](https://golangci-lint.run/)

### Build

```bash
make build
```

### Run Unit Tests

```bash
make test
```

### Run Acceptance Tests

Acceptance tests create real resources against the Unstructured API:

```bash
export UNSTRUCTURED_API_KEY="your-api-key"
make testacc
```

### Generate Documentation

```bash
make generate
```

### Full Validation (format, lint, build, generate)

```bash
make
```

## License

MPL-2.0 - See [LICENSE](LICENSE) for details.
