# Look up a workflow template by ID
data "unstructured_template" "example" {
  id = "existing-template-uuid"
}

# Use the template to create a workflow
resource "unstructured_workflow" "from_template" {
  name           = "template-based-pipeline"
  source_id      = "existing-source-uuid"
  destination_id = "existing-destination-uuid"
  workflow_type  = "template"
  template_id    = data.unstructured_template.example.id
}
