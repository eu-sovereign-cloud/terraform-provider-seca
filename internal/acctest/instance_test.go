package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckInstanceDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_instance" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.ComputeV1.GetInstance(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking instance %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("instance %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccInstanceResourceConfig(sshKey string, labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name = "workspace-1"
}

data "seca_instance_sku" "test" {
  name = "DXS"
}

resource "seca_block_storage" "boot" {
  name         = "boot-vol-1"
  workspace_id = seca_workspace.test.name

  size_gb = 10
  sku_id  = "storage-skus/RD500"
}

resource "seca_instance" "test" {
  name         = "instance-1"
  workspace_id = seca_workspace.test.name

  sku_id   = data.seca_instance_sku.test.id
  ssh_keys = [%q]
  zone     = "zone-a"
  labels   = %s

  boot_volume = {
    device_id = seca_block_storage.boot.id
  }
}
`, sshKey, formatLabels(labels))
}

func testAccInstanceDataSourceConfig(sshKey string, labels map[string]string) string {
	return testAccInstanceResourceConfig(sshKey, labels) + `
data "seca_instance" "test" {
  name         = seca_instance.test.name
  workspace_id = seca_workspace.test.name
}`
}

func TestAccInstance(t *testing.T) {
	sshKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl test@acctest"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceResourceConfig(sshKey, map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_instance.test", "name", "instance-1"),
					resource.TestCheckResourceAttr("seca_instance.test", "workspace_id", "workspace-1"),
					resource.TestCheckResourceAttr("seca_instance.test", "zone", "zone-a"),
					resource.TestCheckResourceAttrSet("seca_instance.test", "id"),
					resource.TestCheckResourceAttrSet("seca_instance.test", "tenant"),
					resource.TestCheckResourceAttrSet("seca_instance.test", "region"),
					resource.TestCheckResourceAttrSet("seca_instance.test", "power_state"),
					resource.TestCheckResourceAttr("seca_instance.test", "boot_volume.device_id", "boot-vol-1"),
					resource.TestCheckResourceAttr("seca_instance.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccInstanceResourceConfig(sshKey, map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_instance.test", "name", "instance-1"),
					resource.TestCheckResourceAttr("seca_instance.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:      "seca_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "workspace-1/instance-1",
			},
			{
				Config: testAccInstanceDataSourceConfig(sshKey, map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_instance.test", "name", "instance-1"),
					resource.TestCheckResourceAttr("seca_instance.test", "labels.env", "prod"),
					resource.TestCheckResourceAttr("data.seca_instance.test", "name", "instance-1"),
					resource.TestCheckResourceAttrSet("data.seca_instance.test", "id"),
					resource.TestCheckResourceAttrSet("data.seca_instance.test", "power_state"),
					resource.TestCheckResourceAttrSet("data.seca_instance.test", "state"),
				),
			},
		},
	})
}
