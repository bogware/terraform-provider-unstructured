// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build acceptance

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccWorkflowResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowResourceConfig("test-workflow"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("unstructured_workflow.test", tfjsonpath.New("name"), knownvalue.StringExact("test-workflow")),
					statecheck.ExpectKnownValue("unstructured_workflow.test", tfjsonpath.New("workflow_type"), knownvalue.StringExact("custom")),
				},
			},
			{
				ResourceName:      "unstructured_workflow.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkflowResourceConfig(name string) string {
	return `
resource "unstructured_source" "test" {
  name = "test-wf-source"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://test-bucket/input"
    key        = "test-key"
    secret     = "test-secret"
  })
}

resource "unstructured_destination" "test" {
  name = "test-wf-destination"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://test-bucket/output"
    key        = "test-key"
    secret     = "test-secret"
  })
}

resource "unstructured_workflow" "test" {
  name           = "` + name + `"
  source_id      = unstructured_source.test.id
  destination_id = unstructured_destination.test.id
  workflow_type  = "custom"

  workflow_nodes = jsonencode([
    {
      name    = "partitioner"
      type    = "partition"
      subtype = "vlm"
      settings = {
        provider  = "anthropic"
        model     = "claude-sonnet-4-20250514"
        strategy  = "auto"
      }
    },
    {
      name    = "chunker"
      type    = "chunk"
      subtype = "chunk_by_title"
      settings = {
        max_characters     = 1024
        overlap            = 128
        multipage_sections = true
      }
    }
  ])

  reprocess_all = true
}
`
}
