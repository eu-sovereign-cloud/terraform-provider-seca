package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used by acceptance tests that call
// resource.Test() (requires TF_ACC=1).
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seca": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck skips the test when required env vars are absent.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, env := range []string{"SECA_TOKEN", "SECA_TENANT", "SECA_REGION", "SECA_REGION_V1_URL"} {
		if os.Getenv(env) == "" {
			t.Skipf("%s environment variable not set; skipping acceptance test", env)
		}
	}
}

// testAccProviderConfig returns a provider block using values from environment
// variables; used by apply-time acceptance tests.
func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "seca" {
  token  = %q
  tenant = %q
  region = %q
  global_providers = {
    region_v1        = %q
    authorization_v1 = %q
  }
}
`,
		os.Getenv("SECA_TOKEN"),
		os.Getenv("SECA_TENANT"),
		os.Getenv("SECA_REGION"),
		os.Getenv("SECA_REGION_V1_URL"),
		os.Getenv("SECA_AUTH_V1_URL"),
	)
}

// testUnitProviderConfig returns a provider block with dummy values for use in
// unit-style acceptance tests (resource.UnitTest) that only exercise plan-time
// validation and never reach the API.
const testUnitProviderConfig = `
provider "seca" {
  token  = "test-token"
  tenant = "test-tenant"
  region = "test-region"
  global_providers = {
    region_v1 = "http://localhost:19999"
  }
}
`
