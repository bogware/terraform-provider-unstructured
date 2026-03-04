# S3 source connector
resource "unstructured_source" "s3" {
  name = "my-s3-source"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://my-bucket/documents"
    key        = var.aws_access_key
    secret     = var.aws_secret_key
  })
}

# Azure Blob Storage source connector
resource "unstructured_source" "azure" {
  name = "my-azure-source"
  type = "azure"
  config = jsonencode({
    remote_url    = "az://my-container/documents"
    account_name  = var.azure_account_name
    account_key   = var.azure_account_key
  })
}

# Google Cloud Storage source connector
resource "unstructured_source" "gcs" {
  name = "my-gcs-source"
  type = "gcs"
  config = jsonencode({
    remote_url          = "gs://my-bucket/documents"
    service_account_key = var.gcp_service_account_key
  })
}
