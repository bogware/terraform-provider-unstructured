data "unstructured_template" "example" {
  id = "existing-template-uuid"
}

# Use a template to create a workflow
resource "unstructured_workflow" "from_template" {
  name           = "template-based-pipeline"
  source_id      = unstructured_source.s3.id
  destination_id = unstructured_destination.pinecone.id
  workflow_type  = "template"
  template_id    = data.unstructured_template.example.id
  reprocess_all  = false
}
