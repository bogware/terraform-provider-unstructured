---
page_title: "unstructured_destination Data Source - Unstructured"
subcategory: "Connectors"
description: |-
  Retrieves information about an existing destination connector by ID or name.
---

# unstructured_destination (Data Source)

Use this data source to retrieve information about an existing destination connector. Look up by `id` or `name` (exactly one must be specified).

## Example Usage

{{tffile "examples/data-sources/unstructured_destination/data-source.tf"}}

{{ .SchemaMarkdown | trimspace }}
