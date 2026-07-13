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

func testAccRoleAssignmentResourceConfig(subs string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_role" "test_ra" {
  name = "role-for-assignment-test"

  permissions = [
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }
  ]
}

resource "seca_role_assignment" "test" {
  name = "role-assignment-acctest-1"

  subs = [%s]
  scopes = [
    {
      tenants = [%q]
    }
  ]
  roles = [seca_role.test_ra.name]
}
`, subs, testAccTenant)
}

func testAccRoleAssignmentDataSourceConfig(subs string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "seca_role" "test_ra" {
  name = "role-for-assignment-test"

  permissions = [
    {
      provider  = "seca.storage/v1"
      resources = ["block-storages/*"]
      verb      = ["get", "list"]
    }
  ]
}

resource "seca_role_assignment" "test" {
  name = "role-assignment-acctest-1"

  subs = [%s]
  scopes = [
    {
      tenants = [%q]
    }
  ]
  roles = [seca_role.test_ra.name]
}

data "seca_role_assignment" "test" {
  name = "role-assignment-acctest-1"
}`, subs, testAccTenant)
}

func TestAccRoleAssignment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAssignmentResourceConfig(`"user-test-1"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "name", "role-assignment-acctest-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.0", "user-test-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "scopes.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "scopes.0.tenants.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "scopes.0.tenants.0", testAccTenant),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "roles.0", "role-for-assignment-test"),
				),
			},
			{
				Config: testAccRoleAssignmentResourceConfig(`"user-test-1", "user-test-2"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "name", "role-assignment-acctest-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.#", "2"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "subs.1", "user-test-2"),
				),
			},
			{
				ResourceName:      "seca_role_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "role-assignment-acctest-1",
			},
			{
				Config: testAccRoleAssignmentDataSourceConfig(`"user-test-1"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_role_assignment.test", "name", "role-assignment-acctest-1"),
					resource.TestCheckResourceAttr("seca_role_assignment.test", "tenant", testAccTenant),

					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "name", "role-assignment-acctest-1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "state", "active"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "subs.#", "1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "subs.0", "user-test-1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("data.seca_role_assignment.test", "roles.0", "role-for-assignment-test"),
				),
			},
		},
	})
}
