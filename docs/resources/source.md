---
page_title: "unstructured_source Resource - Unstructured"
subcategory: "Connectors"
description: |-
  Manages a source connector in the Unstructured platform.
---

# unstructured_source (Resource)

Manages a source connector in the Unstructured platform. Source connectors define where data is ingested from (e.g., S3, Azure Blob Storage, Google Drive).

## Example Usage

{{tffile "examples/resources/unstructured_source/resource.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the resource ID.

{{codefile "shell" "examples/resources/unstructured_source/import.sh"}}
