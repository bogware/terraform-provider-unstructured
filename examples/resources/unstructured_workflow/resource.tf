# Custom workflow with VLM partitioning, chunking, and embedding
resource "unstructured_workflow" "custom" {
  name           = "my-document-pipeline"
  source_id      = unstructured_source.s3.id
  destination_id = unstructured_destination.pinecone.id
  workflow_type  = "custom"

  workflow_nodes = jsonencode([
    {
      name    = "partitioner"
      type    = "partition"
      subtype = "vlm"
      settings = {
        provider = "anthropic"
        model    = "claude-sonnet-4-20250514"
        strategy = "auto"
      }
    },
    {
      name    = "chunker"
      type    = "chunk"
      subtype = "chunk_by_title"
      settings = {
        max_characters     = 1024
        overlap            = 128
        multipage_sections = true
      }
    },
    {
      name    = "embedder"
      type    = "embed"
      subtype = "togetherai"
      settings = {
        model_name = "togethercomputer/m2-bert-80M-8k-retrieval"
      }
    }
  ])

  schedule      = "daily"
  reprocess_all = false
}

# Template-based workflow using a pre-configured template
resource "unstructured_workflow" "from_template" {
  name           = "template-pipeline"
  source_id      = unstructured_source.s3.id
  destination_id = unstructured_destination.pinecone.id
  workflow_type  = "platinum"
  template_id    = "template-uuid"
  schedule       = "weekly"
}
