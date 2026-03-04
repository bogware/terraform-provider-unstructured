// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceResourceConfig("test-s3-source"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("unstructured_source.test", tfjsonpath.New("name"), knownvalue.StringExact("test-s3-source")),
					statecheck.ExpectKnownValue("unstructured_source.test", tfjsonpath.New("type"), knownvalue.StringExact("s3")),
				},
			},
			{
				ResourceName:      "unstructured_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSourceResourceConfig(name string) string {
	return `
resource "unstructured_source" "test" {
  name = "` + name + `"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://test-bucket/input"
    key        = "test-key"
    secret     = "test-secret"
  })
}
`
}
