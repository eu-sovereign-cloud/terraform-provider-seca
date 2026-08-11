package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccInstanceSkuConfig() string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_instance_sku" "test" {
  name = %q
}`, testAccInstanceSku)
}

func TestAccInstanceSku(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceSkuConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_instance_sku.test", "name", testAccInstanceSku),
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
