package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckWorkspaceDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_workspace" {
			continue
		}

		tref := secapi.TenantReference{
			Tenant: secapi.TenantID(testAccTenant),
			Name:   rs.Primary.Attributes["name"],
		}

		_, err := client.WorkspaceV1.GetWorkspace(ctx, tref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking workspace %q was destroyed: %w", tref.Name, err)
		}
		return fmt.Errorf("workspace %q still exists after destroy", tref.Name)
	}

	return nil
}

func testAccWorkspaceResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name   = "workspace-1"
  labels = %s
  retry = {
    delay        = 10
    interval     = 10
    max_attempts = 3
  }
  timeouts {
    create = "1m"
	update = "1m"
	read =   "30s"
    delete = "1m"
  }
}
`, formatLabels(labels))
}

func testAccWorkspaceDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_workspace" "test" {
  name   = "workspace-1"
  labels = %s
  retry = {
    delay        = 10
    interval     = 10
    max_attempts = 3
  }
  timeouts {
    create = "1m"
	update = "1m"
	read =   "30s"
    delete = "1m"
  }  
}
data "seca_workspace" "test" {
  name = "workspace-1"
}`, formatLabels(labels))
}

func TestAccWorkspace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("seca_workspace.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccWorkspaceResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:            "seca_workspace.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "workspace-1",
				ImportStateVerifyIgnore: []string{"retry"},
			},
			{
				Config: testAccWorkspaceDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),

					resource.TestCheckResourceAttr("data.seca_workspace.test", "name", "workspace-1"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "tenant", "seca"),
					resource.TestCheckResourceAttr("seca_workspace.test", "region", "region"),
					resource.TestCheckResourceAttr("data.seca_workspace.test", "state", "active"),
				),
			},
		},
	})
}
