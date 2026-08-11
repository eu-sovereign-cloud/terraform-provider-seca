package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccStorageSkuDataSourceConfig() string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "seca_storage_sku" "test" {
  name = %q
}`, testAccStorageSku)
}

func TestAccStorageSku(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageSkuDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "name", testAccStorageSku),
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "tenant", testAccTenant),
					resource.TestCheckResourceAttr("data.seca_storage_sku.test", "region", testAccRegion),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "iops"),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "type"),
					resource.TestCheckResourceAttrSet("data.seca_storage_sku.test", "min_volume_size"),
				),
			},
		},
	})
}
