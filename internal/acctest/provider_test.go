package acctest

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/provider"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	testAccToken        = getEnvDefault("SECA_TEST_TOKEN", "test-token")
	testAccTenant       = getEnvDefault("SECA_TEST_TENANT", "tenant-1")
	testAccRegion       = getEnvDefault("SECA_TEST_REGION", "region-1")
	testAccEndpointReg  = getEnvDefault("SECA_TEST_REGION_ENDPOINT", "http://localhost:8080/providers/seca.region")
	testAccEndpointAuth = getEnvDefault("SECA_TEST_AUTH_ENDPOINT", "http://localhost:8080/providers/seca.authorization")
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seca": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "seca" {
  token  = %q
  tenant = %q
  region = %q
  global_providers = {
    region_v1        = %q,
    authorization_v1 = %q
  }
}`, testAccToken, testAccTenant, testAccRegion, testAccEndpointReg, testAccEndpointAuth)
}

func testAccRegionalClient(ctx context.Context) (*secapi.RegionalClient, error) {
	globalClient, err := secapi.NewGlobalClient(&secapi.GlobalConfig{
		AuthToken: testAccToken,
		Endpoints: secapi.GlobalEndpoints{
			RegionV1:        testAccEndpointReg,
			AuthorizationV1: testAccEndpointAuth,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create global client: %w", err)
	}

	regionalClient, err := globalClient.NewRegionalClient(ctx, testAccRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to create regional client: %w", err)
	}

	return regionalClient, nil
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("{\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "    %s = %q\n", k, labels[k])
	}
	b.WriteString("  }")
	return b.String()
}

// The `id` attribute, and every reference attribute that carries it, always holds the
// full URN — `{provider}/{version}/tenants/{tenant}[/workspaces/{workspace}[/…]]/{type}/{name}`

func urnTenantScoped(provider, kind, name string) string {
	return fmt.Sprintf("%s/tenants/%s/%s/%s", provider, testAccTenant, kind, name)
}

func urnWorkspaceScoped(provider, workspace, kind, name string) string {
	return fmt.Sprintf("%s/tenants/%s/workspaces/%s/%s/%s", provider, testAccTenant, workspace, kind, name)
}

func urnWorkspace(name string) string {
	return urnTenantScoped("seca.workspace/v1", "workspaces", name)
}

func urnNetwork(workspace, name string) string {
	return urnWorkspaceScoped("seca.network/v1", workspace, "networks", name)
}

func urnSubnet(workspace, network, name string) string {
	return urnNetwork(workspace, network) + "/subnets/" + name
}

func urnRouteTable(workspace, network, name string) string {
	return urnNetwork(workspace, network) + "/route-tables/" + name
}

func urnBlockStorage(workspace, name string) string {
	return urnWorkspaceScoped("seca.storage/v1", workspace, "block-storages", name)
}

func urnNetworkSku(name string) string {
	return urnTenantScoped("seca.network/v1", "skus", name)
}
