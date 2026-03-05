---
page_title: "unstructured_destination Resource - Unstructured"
subcategory: "Connectors"
description: |-
  Manages a destination connector in the Unstructured platform.
---

# unstructured_destination (Resource)

Manages a destination connector in the Unstructured platform. Destination connectors define where processed data is delivered (e.g., S3, Pinecone, Elasticsearch).

## Example Usage

{{tffile "examples/resources/unstructured_destination/resource.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the resource ID.

{{codefile "shell" "examples/resources/unstructured_destination/import.sh"}}
