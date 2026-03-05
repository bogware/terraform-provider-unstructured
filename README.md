# Terraform Provider for Unstructured

There's a lot of territory out there in the world of unstructured data. Documents scattered across cloud buckets, PDFs piling up like tumbleweeds, and no clear trail between raw files and the vector databases that need them. This provider brings law and order to that frontier.

The Unstructured Terraform provider manages resources on the [Unstructured](https://unstructured.io/) platform -- source connectors, destination connectors, and processing workflows -- so you can wrangle your document pipelines the same way you manage the rest of your infrastructure: declaratively, repeatably, and under version control.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.13
- [Go](https://golang.org/doc/install) >= 1.24 (only if building from source)
- An [Unstructured Platform](https://platform.unstructuredapp.io/) account and API key

## Getting Started

### 1. Declare the Provider

Tell Terraform where to find the provider. That's your first move -- like saddling up before you ride.

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

### 2. Configure Authentication

The provider needs an API key. You can set it in the provider block or, better yet, keep it out of your configuration files entirely and use an environment variable. A man doesn't leave his keys lying around in the street.

```bash
export UNSTRUCTURED_API_KEY="your-api-key"
```

Or configure it directly:

```hcl
provider "unstructured" {
  api_key = var.unstructured_api_key
}
```

### 3. Define Your Source

A source connector tells the platform where your documents live. Could be an S3 bucket, Azure Blob Storage, Google Drive -- wherever the trail starts.

```hcl
resource "unstructured_source" "documents" {
  name = "company-documents"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://my-bucket/documents"
    key        = var.aws_access_key
    secret     = var.aws_secret_key
  })
}
```

### 4. Define Your Destination

A destination connector is where the processed data ends up. Pinecone, Elasticsearch, Weaviate, another S3 bucket -- wherever the trail leads.

```hcl
resource "unstructured_destination" "vectors" {
  name = "pinecone-index"
  type = "pinecone"
  config = jsonencode({
    api_key    = var.pinecone_api_key
    index_name = "documents"
  })
}
```

### 5. Wire It Together with a Workflow

A workflow is the heart of the operation. It connects your source to your destination and defines the processing pipeline in between -- partitioning, chunking, embedding, the works.

```hcl
resource "unstructured_workflow" "pipeline" {
  name            = "document-pipeline"
  source_id       = unstructured_source.documents.id
  destination_id  = unstructured_destination.vectors.id
  workflow_type   = "custom"
  schedule        = "daily"
  reprocess_all   = false

  workflow_nodes = jsonencode([
    {
      name    = "partitioner"
      type    = "partition"
      subtype = "vlm"
      settings = {
        provider = "anthropic"
        model    = "claude-sonnet-4-20250514"
      }
    },
    {
      name    = "chunker"
      type    = "chunk"
      subtype = "chunk_by_title"
      settings = {
        max_characters             = 1500
        combine_text_under_n_chars = 500
        multipage_sections         = true
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

### 6. Deploy

```bash
terraform init
terraform plan
terraform apply
```

That's it. Three resources, one pipeline, and your documents are moving from raw storage to a vector database on a schedule. No console clicking. No manual steps. Just Terraform doing what Terraform does.

## Resources

These are the managed resources -- things the provider creates, reads, updates, and deletes on your behalf.

| Resource | Description |
|----------|-------------|
| `unstructured_source` | A source connector. Defines where documents are ingested from. Supports 20+ connector types. |
| `unstructured_destination` | A destination connector. Defines where processed data is delivered. Supports 25+ connector types. |
| `unstructured_workflow` | A processing workflow. Connects a source to a destination through a pipeline of processing nodes. |

All resources support `terraform import` by ID.

## Data Sources

Data sources let you read existing resources without managing them. Useful for referencing connectors or workflows that were created somewhere else -- maybe by hand, maybe by another team. A good marshal knows what's already in town before he starts making changes.

| Data Source | Lookup | Description |
|-------------|--------|-------------|
| `unstructured_source` | By `id` or `name` | Read an existing source connector. |
| `unstructured_destination` | By `id` or `name` | Read an existing destination connector. |
| `unstructured_workflow` | By `id` | Read an existing workflow. |
| `unstructured_job` | By `id` | Read a workflow job execution. |
| `unstructured_template` | By `id` | Read an available workflow template for template-based workflows. |

```hcl
# Look up a source by name -- no need to hunt for the UUID
data "unstructured_source" "existing" {
  name = "my-s3-source"
}

# Look up a template and use it to build a workflow
data "unstructured_template" "basic" {
  id = "template-uuid"
}

resource "unstructured_workflow" "from_template" {
  name            = "template-pipeline"
  source_id       = data.unstructured_source.existing.id
  destination_id  = unstructured_destination.vectors.id
  workflow_type   = "template"
  template_id     = data.unstructured_template.basic.id
}
```

## Supported Connector Types

The Unstructured platform supports a wide range of sources and destinations. Each connector type has its own configuration shape, passed as a JSON-encoded string via the `config` attribute. Consult the [Unstructured documentation](https://docs.unstructured.io/) for the specific fields each connector expects.

### Sources

`azure` `box` `confluence` `couchbase` `databricks_volumes` `dropbox` `elasticsearch` `gcs` `google_drive` `jira` `kafka-cloud` `mongodb` `onedrive` `opensearch` `outlook` `postgres` `s3` `salesforce` `sftp` `sharepoint` `slack` `snowflake` `teradata` `zendesk`

### Destinations

`azure` `astradb` `azure_ai_search` `couchbase` `databricks_volumes` `databricks_volume_delta_tables` `delta_table` `elasticsearch` `gcs` `kafka-cloud` `milvus` `mongodb` `motherduck` `neo4j` `onedrive` `opensearch` `pinecone` `postgres` `qdrant-cloud` `redis` `s3` `snowflake` `teradata` `weaviate-cloud` `ibm_watsonx_s3`

## Workflow Nodes

When building a `custom` workflow, the `workflow_nodes` attribute takes a JSON-encoded array of processing steps. Each node has a `name`, `type`, `subtype`, and optional `settings`. Think of them as stations along the trail -- your documents stop at each one and come out the other side a little more refined.

| Type | Subtypes | What It Does |
|------|----------|--------------|
| `partition` | `auto`, `vlm`, `fast`, `hi_res` | Extracts structured elements from raw documents. The first stop on every trail. |
| `chunk` | `chunk_by_title`, `chunk_by_page`, `chunk_by_similarity`, `chunk_by_character` | Splits partitioned elements into right-sized chunks for downstream use. |
| `embed` | `openai`, `azure_openai`, `togetherai`, `huggingface`, `octoai`, `bedrock` | Generates vector embeddings. The step that makes your documents searchable. |
| `enrich` | `summarize`, `image_summarize` | Enriches elements with AI-generated summaries. |
| `filter` | `filter_element_types` | Filters elements by type before they move further down the line. |

### Schedule Options

Workflows can run on a schedule. Set the `schedule` attribute to one of these values:

`every 15 minutes` `every hour` `every 2 hours` `every 4 hours` `every 6 hours` `every 8 hours` `every 10 hours` `every 12 hours` `daily` `weekly` `monthly`

## Configuration Reference

### Provider

| Attribute | Required | Sensitive | Description |
|-----------|----------|-----------|-------------|
| `api_key` | No | Yes | API key for the Unstructured platform. Falls back to `UNSTRUCTURED_API_KEY` env var. |
| `api_url` | No | No | Base URL for the API. Defaults to `https://platform.unstructuredapp.io/api/v1`. Falls back to `UNSTRUCTURED_API_URL` env var. |

## Importing Existing Resources

If you've got resources already running on the platform -- built before you decided to bring Terraform into town -- you can import them:

```bash
terraform import unstructured_source.my_source "source-uuid"
terraform import unstructured_destination.my_dest "destination-uuid"
terraform import unstructured_workflow.my_workflow "workflow-uuid"
```

Write the matching resource block in your configuration first, then run the import. Terraform will read the current state from the API and sync up. After that, it's under management -- same as if you'd created it from scratch.

## Development

If you're looking to work on the provider itself, here's how things are laid out.

### Build

```bash
git clone https://github.com/bogware/terraform-provider-unstructured.git
cd terraform-provider-unstructured
make build
```

### Unit Tests

Unit tests run against mocked HTTP endpoints. No API key or Terraform binary needed. These are the tests that run on every pull request.

```bash
make test
```

### Acceptance Tests

Acceptance tests run against the live Unstructured API. They create and destroy real resources. Set your API key first.

```bash
export UNSTRUCTURED_API_KEY="your-api-key"
make testacc
```

### Generate Documentation

The docs in `docs/` are generated from the provider schema and the example files in `examples/`. If you change a schema or an example, regenerate:

```bash
make generate
```

### Full Validation

```bash
make        # runs fmt, lint, install, generate
```

## License

MPL-2.0. See [LICENSE](LICENSE) for the full text.
