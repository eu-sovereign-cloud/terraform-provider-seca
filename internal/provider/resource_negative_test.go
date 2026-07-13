package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Plan-time validation tests (resource.UnitTest — no TF_ACC required)
// ---------------------------------------------------------------------------

// TestAccNetwork_InvalidCIDR verifies that an invalid IPv4 CIDR on seca_network
// produces a plan-time diagnostic rather than an API error.
func TestAccNetwork_InvalidCIDR(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUnitProviderConfig + `
resource "seca_network" "invalid" {
  name         = "test-network"
  workspace_id = "fake-workspace"
  sku_id       = "fake-sku"
  cidr = {
    ipv4 = "not-a-valid-cidr"
  }
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid (IPv4 )?CIDR`),
			},
		},
	})
}

// TestAccSecurityGroup_InvalidDirection verifies that an unrecognised direction
// value triggers a plan-time enum validation error.
func TestAccSecurityGroup_InvalidDirection(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUnitProviderConfig + `
resource "seca_security_group" "invalid" {
  name         = "test-sg"
  workspace_id = "fake-workspace"
  rules = [
    {
      direction = "forward"
      protocol  = "tcp"
    }
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid enum value`),
			},
		},
	})
}

// TestAccSecurityGroup_InvalidPortRange verifies that port 0 fails plan-time
// validation with a clear port-range error.
func TestAccSecurityGroup_InvalidPortRange(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUnitProviderConfig + `
resource "seca_security_group" "invalid" {
  name         = "test-sg"
  workspace_id = "fake-workspace"
  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 0
      }
    }
  ]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid port number`),
			},
		},
	})
}

// TestAccPublicIp_InvalidVersion verifies that an unrecognised IP version
// (e.g. "IPv5") fails plan-time validation.
func TestAccPublicIp_InvalidVersion(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testUnitProviderConfig + `
resource "seca_public_ip" "invalid" {
  name         = "test-pip"
  workspace_id = "fake-workspace"
  version      = "IPv5"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid enum value`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Apply-time API error tests (resource.Test — requires TF_ACC=1)
// ---------------------------------------------------------------------------

// TestAccBlockStorage_InvalidSKU verifies that creating a block-storage volume
// with a non-existent SKU fails with a clear API error at apply time.
func TestAccBlockStorage_InvalidSKU(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "seca_workspace" "test" {
  name = "neg-test-workspace"
}

resource "seca_block_storage" "invalid_sku" {
  name         = "neg-test-storage"
  workspace_id = seca_workspace.test.id
  size_gb      = 10
  sku_id       = "nonexistent-sku-000"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(error creating block storage|invalid.*sku|not found)`),
			},
		},
	})
}

// TestAccBlockStorage_MissingWorkspace verifies that referencing a workspace
// that does not exist produces a clear error diagnostic at apply time.
func TestAccBlockStorage_MissingWorkspace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "seca_block_storage" "missing_ws" {
  name         = "neg-test-storage"
  workspace_id = "00000000-0000-0000-0000-000000000000"
  size_gb      = 10
  sku_id       = "some-sku"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(error creating block storage|workspace.*not found|not found)`),
			},
		},
	})
}
