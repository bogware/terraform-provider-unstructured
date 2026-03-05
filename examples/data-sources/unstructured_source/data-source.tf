# Look up by ID
data "unstructured_source" "by_id" {
  id = "existing-source-connector-uuid"
}

# Look up by name
data "unstructured_source" "by_name" {
  name = "my-s3-source"
}
