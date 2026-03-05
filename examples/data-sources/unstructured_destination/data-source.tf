# Look up by ID
data "unstructured_destination" "by_id" {
  id = "existing-destination-connector-uuid"
}

# Look up by name
data "unstructured_destination" "by_name" {
  name = "my-pinecone-destination"
}
