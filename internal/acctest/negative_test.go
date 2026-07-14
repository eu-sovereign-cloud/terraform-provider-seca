package acctest

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccBlockStorageInvalidSKUConfig returns a config that attempts to create
// block storage with a non-existent SKU.
func testAccBlockStorageInvalidSKUConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-neg-1"
}
resource "seca_block_storage" "invalid_sku" {
  name         = "bs-invalid-sku"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/DOES-NOT-EXIST-XYZ"
}`
}

// testAccBlockStorageMissingWorkspaceConfig returns a config that attempts to
// create block storage referencing a workspace that does not exist.
func testAccBlockStorageMissingWorkspaceConfig() string {
	return testAccProviderConfig() + `
resource "seca_block_storage" "missing_ws" {
  name         = "bs-missing-ws"
  workspace_id = "workspace-does-not-exist"

  size_gb = 10
  sku_id  = "storage-skus/RD500"
}`
}

// TestAccBlockStorage_InvalidSKU verifies that a non-existent SKU produces
// a clear error diagnostic rather than a panic or silent failure.
func TestAccBlockStorage_InvalidSKU(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBlockStorageInvalidSKUConfig(),
				ExpectError: regexp.MustCompile(`Error creating block storage`),
			},
		},
	})
}

// TestAccBlockStorage_MissingWorkspace verifies that referencing a
// non-existent workspace produces a clear error diagnostic.
func TestAccBlockStorage_MissingWorkspace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBlockStorageMissingWorkspaceConfig(),
				ExpectError: regexp.MustCompile(`Error creating block storage`),
			},
		},
	})
}
