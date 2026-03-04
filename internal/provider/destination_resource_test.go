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

func TestAccDestinationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationResourceConfig("test-s3-destination"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("unstructured_destination.test", tfjsonpath.New("name"), knownvalue.StringExact("test-s3-destination")),
					statecheck.ExpectKnownValue("unstructured_destination.test", tfjsonpath.New("type"), knownvalue.StringExact("s3")),
				},
			},
			{
				ResourceName:      "unstructured_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDestinationResourceConfig(name string) string {
	return `
resource "unstructured_destination" "test" {
  name = "` + name + `"
  type = "s3"
  config = jsonencode({
    remote_url = "s3://test-bucket/output"
    key        = "test-key"
    secret     = "test-secret"
  })
}
`
}
