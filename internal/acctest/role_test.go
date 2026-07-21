package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccGlobalClient() (*secapi.GlobalClient, error) {
	return secapi.NewGlobalClient(&secapi.GlobalConfig{
		AuthToken: testAccToken,
		Endpoints: secapi.GlobalEndpoints{
			RegionV1:        testAccEndpointReg,
			AuthorizationV1: testAccEndpointAuth,
		},
	})
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccGlobalClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_role" {
			continue
		}

		tref := secapi.TenantReference{
			Tenant: secapi.TenantID(testAccTenant),
			Name:   rs.Primary.Attributes["name"],
		}

		_, err := client.AuthorizationV1.GetRole(ctx, tref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking role %q was destroyed: %w", tref.Name, err)
		}
		return fmt.Errorf("role %q still exists after destroy", tref.Name)
	}

	return nil
}

func testAccRoleResourceConfig(permissions string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_role" "test" {
  name = "role-1"

  permissions = [%s]

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
`, permissions)
}

func testAccRoleDataSourceConfig() string {
	return testAccProviderConfig() + `
resource "seca_role" "test" {
  name = "role-1"

  permissions = [
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }
  ]

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
data "seca_role" "test" {
  name = seca_role.test.name
}`
}

func TestAccRole(t *testing.T) {
	permCreate := `
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }`

	permUpdate := `
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list", "put", "delete"]
    }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResourceConfig(permCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role.test", "name", "role-1"),
					resource.TestCheckResourceAttr("seca_role.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("seca_role.test", "permissions.0.provider", "seca.storage/v1"),
					resource.TestCheckResourceAttr("seca_role.test", "permissions.0.verb.#", "2"),
				),
			},
			{
				Config: testAccRoleResourceConfig(permUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role.test", "name", "role-1"),
					resource.TestCheckResourceAttr("seca_role.test", "permissions.0.verb.#", "4"),
				),
			},
			{
				ResourceName:            "seca_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "role-1",
				ImportStateVerifyIgnore: []string{"retry"},
			},
			{
				Config: testAccRoleDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role.test", "name", "role-1"),

					resource.TestCheckResourceAttr("data.seca_role.test", "name", "role-1"),
					resource.TestCheckResourceAttr("data.seca_role.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_role.test", "state", "active"),
					resource.TestCheckResourceAttr("data.seca_role.test", "permissions.#", "1"),
				),
			},
		},
	})
}
