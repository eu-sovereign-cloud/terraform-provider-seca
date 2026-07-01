package acctest

import (
	"testing"

	"github.com/eu-sovereign-cloud/conformance/pkg/mock"
	pkgscenarios "github.com/eu-sovereign-cloud/conformance/pkg/mock/scenarios"
	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/mock/scenarios"
	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seca": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccProviderConfig() string {
	return `
provider "seca" {
  token  = "test"
  tenant = "seca"
  region = "region"
  global_providers = {
    region_v1        = "http://localhost:7070/providers/seca.region",
    authorization_v1 = "http://localhost:7070/providers/seca.authorization"
  }
}`
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Configure the mock
	mockScenario := pkgscenarios.NewScenario("clients.init",
		mock.MockParams{
			ServerURL: "http://localhost:7070",
			AuthToken: "test-token",
		},
	)
	params := scenarios.ClientsInitParams{
		Region: "region-1",
	}
	err := scenarios.ConfigureInitScenarioV1(mockScenario, params)
	if err != nil {
		t.Fatalf("failed to configure mock scenario: %e", err)
	}
}
