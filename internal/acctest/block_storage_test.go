package acctest

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckBlockStorageDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_block_storage" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.StorageV1.GetBlockStorage(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking block storage %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("block storage %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccBlockStorageResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_storage_sku" "test" {
  name = "sku-1"
}
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.test.id
  labels  = %s
  retry = {
    delay        = 10
    interval     = 10
    max_attempts = 3
  }
  timeouts {
    create = "1m"
    update = "1m"
    read   = "30s"
    delete = "1m"
  }
}
`, formatLabels(labels))
}

func testAccBlockStorageDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_storage_sku" "test" {
  name = "sku-1"
}
resource "seca_workspace" "test" {
  name = "workspace-1"
}
resource "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.test.id
  labels  = %s
  retry = {
    delay        = 10
    interval     = 10
    max_attempts = 3
  }
  timeouts {
    create = "1m"
    update = "1m"
    read   = "30s"
    delete = "1m"
  }
}
data "seca_block_storage" "test" {
  name         = "block-storage-1"
  workspace_id = seca_workspace.test.id
}`, formatLabels(labels))
}

func testAccBlockStorageInvalidSKUConfig() string {
	return testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "workspace-neg-1"
}
resource "seca_block_storage" "invalid_sku" {
  name         = "bs-invalid-sku"
  workspace_id = seca_workspace.test.id

  size_gb = 10
  sku_id  = "storage-skus/DOES-NOT-EXIST-XYZ"
}`
}

func testAccBlockStorageMissingWorkspaceConfig() string {
	return testAccProviderConfig() + `
resource "seca_block_storage" "missing_ws" {
  name         = "bs-missing-ws"
  workspace_id = "workspace-does-not-exist"

  size_gb = 10
  sku_id  = "storage-skus/RD500"
}`
}

func TestAccBlockStorage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBlockStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", urnWorkspace("workspace-1")),
					resource.TestCheckResourceAttr("seca_block_storage.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_block_storage.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccBlockStorageResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:            "seca_block_storage.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           urnWorkspace("workspace-1") + "/block-storage-1",
				ImportStateVerifyIgnore: []string{"retry"},
			},
			{
				Config: testAccBlockStorageDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "workspace_id", urnWorkspace("workspace-1")),
					resource.TestCheckResourceAttr("seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("seca_block_storage.test", "labels.env", "prod"),

					resource.TestCheckResourceAttr("data.seca_block_storage.test", "name", "block-storage-1"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "workspace_id", urnWorkspace("workspace-1")),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("data.seca_block_storage.test", "state", "active"),
				),
			},
		},
	})
}

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
