package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckSubnetDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := testAccRegionalClient(ctx)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "seca_subnet" {
			continue
		}

		nref := secapi.NetworkReference{
			Tenant:    secapi.TenantID(testAccTenant),
			Workspace: secapi.WorkspaceID(rs.Primary.Attributes["workspace_id"]),
			Network:   secapi.NetworkID(rs.Primary.Attributes["network_id"]),
			Name:      rs.Primary.Attributes["name"],
		}

		_, err := client.NetworkV1.GetSubnet(ctx, nref)
		if err == secapi.ErrResourceNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking subnet %q was destroyed: %w", nref.Name, err)
		}
		return fmt.Errorf("subnet %q still exists after destroy", nref.Name)
	}

	return nil
}

func testAccSubnetResourceConfig(labels map[string]string) string {
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
  labels         = %s
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
`, testAccNetworkSku, testAccWorkspaceName, testAccNetworkName, testAccInternetGatewayName, testAccRouteTableName, testAccSubnetName, formatLabels(labels))
}

func testAccSubnetUpdateConfig(labels map[string]string) string {
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
  labels         = %s
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
`, testAccNetworkSku, testAccWorkspaceName, testAccNetworkName, testAccInternetGatewayName, testAccRouteTableName, testAccSubnetName, formatLabels(labels))
}

func testAccSubnetDataSourceConfig(labels map[string]string) string {
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
  labels         = %s
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

data "seca_subnet" "test" {
  name         = %q
  workspace_id = seca_workspace.test.id
  network_id   = seca_network.test.id
}`, testAccNetworkSku, testAccWorkspaceName, testAccNetworkName, testAccInternetGatewayName, testAccRouteTableName, testAccSubnetName, formatLabels(labels), testAccSubnetName)
}

func TestAccSubnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetResourceConfig(map[string]string{"env": "dev"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_subnet.test", "name", testAccSubnetName),
					resource.TestCheckResourceAttr("seca_subnet.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
					resource.TestCheckResourceAttr("seca_subnet.test", "network_id", urnNetwork(testAccWorkspaceName, testAccNetworkName)),
					resource.TestCheckResourceAttr("seca_subnet.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("seca_subnet.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("seca_subnet.test", "cidr.ipv4", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("seca_subnet.test", "route_table_id", urnRouteTable(testAccWorkspaceName, testAccNetworkName, testAccRouteTableName)),
					resource.TestCheckResourceAttr("seca_subnet.test", "labels.env", "dev"),
				),
			},
			{
				Config: testAccSubnetUpdateConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_subnet.test", "name", testAccSubnetName),
					resource.TestCheckResourceAttr("seca_subnet.test", "labels.env", "prod"),
				),
			},
			{
				ResourceName:            "seca_subnet.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           urnWorkspace(testAccWorkspaceName) + "/" + urnNetwork(testAccWorkspaceName, testAccNetworkName) + "/" + testAccSubnetName,
				ImportStateVerifyIgnore: []string{"retry"},
			},
			{
				Config: testAccSubnetDataSourceConfig(map[string]string{"env": "prod"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seca_subnet.test", "name", testAccSubnetName),
					resource.TestCheckResourceAttr("seca_subnet.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
					resource.TestCheckResourceAttr("seca_subnet.test", "network_id", urnNetwork(testAccWorkspaceName, testAccNetworkName)),
					resource.TestCheckResourceAttr("seca_subnet.test", "cidr.ipv4", "10.0.1.0/24"),

					resource.TestCheckResourceAttr("data.seca_subnet.test", "name", testAccSubnetName),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "network_id", urnNetwork(testAccWorkspaceName, testAccNetworkName)),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "region", testAccRegion),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "cidr.ipv4", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("data.seca_subnet.test", "state", "active"),
				),
			},
		},
	})
}
