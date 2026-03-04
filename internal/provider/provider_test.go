// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build acceptance

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unstructured": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("UNSTRUCTURED_API_KEY") == "" {
		t.Fatal("UNSTRUCTURED_API_KEY must be set for acceptance tests")
	}
}
