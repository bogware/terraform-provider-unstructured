# S3 destination connector
resource "unstructured_destination" "s3" {
  name = "my-s3-destination"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://my-bucket/processed"
    key        = var.aws_access_key
    secret     = var.aws_secret_key
  })
}

# Elasticsearch destination connector
resource "unstructured_destination" "elasticsearch" {
  name = "my-elasticsearch-destination"
  type = "elasticsearch"
  config = jsonencode({
    hosts      = ["https://my-cluster.es.cloud.io"]
    es_api_key = var.es_api_key
    index_name = "my-index"
  })
}

# Pinecone destination connector
resource "unstructured_destination" "pinecone" {
  name = "my-pinecone-destination"
  type = "pinecone"
  config = jsonencode({
    api_key    = var.pinecone_api_key
    index_name = "my-index"
  })
}
