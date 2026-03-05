---
page_title: "unstructured_source Data Source - Unstructured"
subcategory: "Connectors"
description: |-
  Retrieves information about an existing source connector by ID or name.
---

# unstructured_source (Data Source)

Use this data source to retrieve information about an existing source connector. Look up by `id` or `name` (exactly one must be specified).

## Example Usage

{{tffile "examples/data-sources/unstructured_source/data-source.tf"}}

{{ .SchemaMarkdown | trimspace }}
