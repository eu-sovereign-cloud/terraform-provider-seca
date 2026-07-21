package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckRoleAssignmentDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccGlobalClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_role_assignment" {
			continue
		}

		tref := secapi.TenantReference{
			Tenant: secapi.TenantID(testAccTenant),
			Name:   rs.Primary.Attributes["name"],
		}

		_, err := client.AuthorizationV1.GetRoleAssignment(ctx, tref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking role assignment %q was destroyed: %w", tref.Name, err)
		}
		return fmt.Errorf("role assignment %q still exists after destroy", tref.Name)
	}

	return nil
}

func testAccRoleAssignmentResourceConfig(subs, roles string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_role" "test" {
  name = "role-1"

  permissions = [
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }
  ]
}

resource "seca_role_assignment" "test" {
  name  = "ra-1"
  subs  = [%s]
  roles = [seca_role.test.name]

  scopes = [
    {
      tenants = [%s]
    }
  ]
}
`, subs, roles)
}

func testAccRoleAssignmentDataSourceConfig() string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_role" "test" {
  name = "role-1"

  permissions = [
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }
  ]
}

resource "seca_role_assignment" "test" {
  name  = "ra-1"
  subs  = [%q]
  roles = [seca_role.test.name]

  scopes = [
    {
      tenants = [%q]
    }
  ]
}

data "seca_role_assignment" "test" {
  name = seca_role_assignment.test.name
}`, "sa-1", testAccTenant)
}

func TestAccRoleAssignment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAssignmentResourceConfig(fmt.Sprintf("%q", "sa-1"), fmt.Sprintf("%q", testAccTenant)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "name", "ra-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.0", "sa-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "scopes.#", "1"),
				),
			},
			{
				Config: testAccRoleAssignmentResourceConfig(fmt.Sprintf("%q, %q", "sa-1", "sa-2"), fmt.Sprintf("%q", testAccTenant)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.#", "2"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.0", "sa-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.1", "sa-2"),
				),
			},
			{
				ResourceName:      "seca_role_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "ra-1",
			},
			{
				Config: testAccRoleAssignmentDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "name", "ra-1"),

					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "name", "ra-1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "state", "active"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "subs.#", "1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "scopes.#", "1"),
				),
			},
		},
	})
}
