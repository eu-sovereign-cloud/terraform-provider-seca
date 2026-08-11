package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccInstanceSkuConfig() string {
	return testAccProviderConfig() + `
data "seca_instance_sku" "test" {
  name = "compute-sku-1"
}`
}

func TestAccInstanceSku(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceSkuConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_instance_sku.test", "name", "compute-sku-1"),
					resource.TestCheckResourceAttr("data.seca_instance_sku.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_instance_sku.test", "region", testAccRegion),
					resource.TestCheckResourceAttrSet("data.seca_instance_sku.test", "id"),
					resource.TestCheckResourceAttrSet("data.seca_instance_sku.test", "vcpus"),
					resource.TestCheckResourceAttrSet("data.seca_instance_sku.test", "ram"),
				),
			},
		},
	})
}
