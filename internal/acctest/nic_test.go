package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckNicDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_nic" {
			continue
		}

		wref := secapi.WorkspaceReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetNic(ctx, wref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking NIC %q was destroyed: %w", wref.Name, err)
		}
		return fmt.Errorf("NIC %q still exists after destroy", wref.Name)
	}

	return nil
}

func testAccNicResourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_network_sku" "test" {
  name = %q
}

resource "seca_workspace" "test" {
  name = %q
}
resource "seca_network" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id

  sku_id = data.seca_network_sku.test.id
  cidr = {
    ipv4 = "10.0.0.0/16"
  }
}
resource "seca_internet_gateway" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
}
resource "seca_route_table" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  network_id   = seca_network.test.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.test.id
    }
  ]
}
resource "seca_subnet" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  network_id   = seca_network.test.id

  cidr = {
    ipv4 = "10.0.1.0/24"
  }
  route_table_id = seca_route_table.test.id
  zone           = "zone-a"
}
resource "seca_nic" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  subnet_id    = seca_subnet.test.id

  addresses = ["10.0.1.10"]

  labels = %s
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
`, testAccNetworkSku, testAccWorkspaceName, testAccNetworkName, testAccInternetGatewayName, testAccRouteTableName, testAccSubnetName, testAccNicName, formatLabels(labels))
}

func testAccNicDataSourceConfig(labels map[string]string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_network_sku" "test" {
  name = %q
}

resource "seca_workspace" "test" {
  name = %q
}
resource "seca_network" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id

  sku_id = data.seca_network_sku.test.id
  cidr = {
    ipv4 = "10.0.0.0/16"
  }
}
resource "seca_internet_gateway" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
}
resource "seca_route_table" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  network_id   = seca_network.test.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.test.id
    }
  ]
}
resource "seca_subnet" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  network_id   = seca_network.test.id

  cidr = {
    ipv4 = "10.0.1.0/24"
  }
  route_table_id = seca_route_table.test.id
  zone           = "zone-a"
}
resource "seca_nic" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  subnet_id    = seca_subnet.test.id

  addresses = ["10.0.1.10"]

  labels = %s
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
data "seca_nic" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
}`, testAccNetworkSku, testAccWorkspaceName, testAccNetworkName, testAccInternetGatewayName, testAccRouteTableName, testAccSubnetName, testAccNicName, formatLabels(labels), testAccNicName)
}

func TestAccNic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNicResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", testAccNicName),
					resource.TestCheckResourceAttr("seca_nic.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
					resource.TestCheckResourceAttr("seca_nic.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_nic.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_nic.test", "subnet_id", urnSubnet(testAccWorkspaceName, testAccNetworkName, testAccSubnetName)),
					resource.TestCheckResourceAttr("seca_nic.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccNicResourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", testAccNicName),
					resource.TestCheckResourceAttr("seca_nic.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:            "seca_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           urnWorkspace(testAccWorkspaceName) + "/" + testAccNicName,
				ImportStateVerifyIgnore: []string{"retry"},
			},
			{
				Config: testAccNicDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_nic.test", "name", testAccNicName),
					resource.TestCheckResourceAttr("seca_nic.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),

					resource.TestCheckResourceAttr("data.seca_nic.test", "name", testAccNicName),
					resource.TestCheckResourceAttr("data.seca_nic.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
					resource.TestCheckResourceAttr("data.seca_nic.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_nic.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_nic.test", "subnet_id", urnSubnet(testAccWorkspaceName, testAccNetworkName, testAccSubnetName)),
					resource.TestCheckResourceAttr("data.seca_nic.test", "state", "active"),
				),
			},
		},
	})
}
