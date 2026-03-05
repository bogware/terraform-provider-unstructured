---
page_title: "unstructured_workflow Resource - Unstructured"
subcategory: "Workflows"
description: |-
  Manages a workflow in the Unstructured platform.
---

# unstructured_workflow (Resource)

Manages a workflow in the Unstructured platform. Workflows define how data is processed, connecting source connectors to destination connectors through a configurable pipeline of partitioning, chunking, enrichment, and embedding steps.

## Example Usage

{{tffile "examples/resources/unstructured_workflow/resource.tf"}}

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the resource ID.

{{codefile "shell" "examples/resources/unstructured_workflow/import.sh"}}
