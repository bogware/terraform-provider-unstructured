---
page_title: "unstructured_template Data Source - Unstructured"
subcategory: "Workflows"
description: |-
  Retrieves information about an available workflow template.
---

# unstructured_template (Data Source)

Use this data source to retrieve information about an available workflow template. Templates provide pre-configured workflow node pipelines that can be used when creating template-based workflows via the `template_id` attribute on `unstructured_workflow`.

## Example Usage

{{tffile "examples/data-sources/unstructured_template/data-source.tf"}}

{{ .SchemaMarkdown | trimspace }}
