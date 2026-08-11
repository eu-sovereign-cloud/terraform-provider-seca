# Testing Strategy

## Two-Layer Approach

The provider uses two completely separate test layers, each with a different purpose and location.

## Layer 1: Unit Tests (`internal/provider/`)

**Package:** `package provider` (white-box access to unexported functions)

**Purpose:** Verify that mapping functions between SDK types and Terraform model types work correctly — including all edge cases (null, zero, empty, missing optional fields).

**Scope:** Only `types.go` helpers and the `xxxToXxxModel()` / `xxxFromModel()` private functions. No HTTP calls. No mocking framework.

**Pattern:**
```go
func TestBlockStorageToResourceModel(t *testing.T) {
    // Build an SDK object with known fields
    block := &sdk.BlockStorage{
        Metadata: &sdk.RegionalWorkspaceResourceMetadata{...},
        Spec: sdk.BlockStorageSpec{...},
    }

    // Call the mapping function
    model, diags := blockStorageToResourceModel(context.Background(), block)

    // Assert no errors
    require.False(t, diags.HasError())

    // Assert every field
    assert.Equal(t, "expected", model.Field.ValueString())
    assert.True(t, model.NullableField.IsNull())
}
```

**Test naming:** `Test<FunctionName>` e.g. `TestBlockStorageToResourceModel`, `TestFromTime`.

**What to cover in unit tests:**
- All standard metadata fields (id, name, tenant, region, created_at, deleted_at, last_modified_at)
- All labels/annotations/extensions (non-nil and nil/empty)
- All spec fields (required and optional)
- Optional pointer fields: nil → null, non-nil → value
- Status fields (data sources only): each possible state value
- Edge cases: zero time, nil pointer, empty map, null map

**What NOT to cover in unit tests:** API calls, provider lifecycle, Terraform plan/apply — that's Layer 2.

## Layer 2: Acceptance Tests (`internal/acctest/`)

**Package:** `package acctest` (separate package, no white-box access)

**Purpose:** Verify end-to-end behavior against a live SECA cluster. Tests create real resources, verify their state via Terraform, and destroy them on teardown.

**Guard:** Tests only run when `TF_ACC=1` is set. The CI job sets this. Local runs require a live cluster.

**Cluster:** Tests use the provider config in `provider_test.go`, which reads endpoints from environment variables (falling back to `http://localhost:8080/...` defaults). Point the tests at any cluster by setting:
- `SECA_TEST_REGION_ENDPOINT` — the `seca.region` provider endpoint
- `SECA_TEST_AUTH_ENDPOINT` — the `seca.authorization` provider endpoint
- `SECA_TEST_TOKEN`, `SECA_TEST_TENANT`, `SECA_TEST_REGION` — credentials and scope
- `SECA_TEST_STORAGE_SKU`, `SECA_TEST_INSTANCE_SKU`, `SECA_TEST_NETWORK_SKU` — the SKU names the tests reference, exposed as `testAccStorageSku` / `testAccInstanceSku` / `testAccNetworkSku`

No source edits are needed to target a different environment.

**Never hardcode environment-derived values in assertions.** Any check whose expected value comes from the cluster identity — `tenant`, `region`, SKU names, and configs that name a region — must compare against `testAccTenant` / `testAccRegion` / `testAccStorageSku` and friends, not a literal. SKU catalogs differ per environment, so a config that names a SKU must interpolate the variable:

```go
return testAccProviderConfig() + fmt.Sprintf(`
data "seca_storage_sku" "test" {
  name = %q
}
...
`, testAccStorageSku, formatLabels(labels))
```

Literals only belong to values the test itself supplies (sizes, labels) or that are genuinely fixed by the API.

**Never hardcode resource names either.** Every name a test creates carries a per-run suffix (`testAccRunID`, the run's Unix time in milliseconds) so a run cannot collide with resources an earlier run left behind on a shared environment. `provider_test.go` declares one variable per resource kind — `testAccWorkspaceName`, `testAccNetworkName`, `testAccSubnetName`, … — built by `testAccResourceName(prefix)`. Use them in configs, assertions, `ImportStateId`, and the `urn*` helpers:

```go
resource "seca_workspace" "test" {
  name = %q
}
`, testAccWorkspaceName, ...)

resource.TestCheckResourceAttr("seca_network.test", "workspace_id", urnWorkspace(testAccWorkspaceName)),
ImportStateId: urnWorkspace(testAccWorkspaceName) + "/" + testAccNetworkName,
```

The variables are package-level, so a name is stable across every step of a run — never generate a name inside a config function, or step 2 would target a different resource than step 1. A new resource kind needs a new `testAccResourceName` variable rather than a literal.

**Pattern:**
```go
func TestAccBlockStorage(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckBlockStorageDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccBlockStorageResourceConfig(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("seca_block_storage.test", "name", "block-storage-1"),
                    ...
                ),
            },
            {
                Config: testAccBlockStorageDataSourceConfig(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.seca_block_storage.test", "state", "active"),
                    ...
                ),
            },
        },
    })
}
```

**Config builder convention:** Private functions named `testAcc<Resource><Type>Config()` where type is `ResourceConfig`, `DataSourceConfig`, or `UpdateConfig`.

**What each acceptance test must cover:**

| Step | What to verify |
|---|---|
| Step 1: Create resource | All user-specified fields are set correctly |
| Step 2: Create + data source | Data source reads resource state; `state` = "active" |
| Step 3 (if updatable fields exist): Update | Updated fields are reflected |
| `CheckDestroy` | After teardown, each resource returns `secapi.ErrResourceNotFound` from the API |

**Destroy verification (`CheckDestroy`):** Every `TestCase` sets `CheckDestroy: testAccCheck<Resource>Destroy`. The helper builds an SDK client via `testAccRegionalClient(ctx)` (shares credentials/endpoints with the provider config), iterates state resources of its type, and asserts each `Get<Resource>` call returns `secapi.ErrResourceNotFound`. A still-present resource (or any other error) fails the test, catching orphaned API resources that Terraform's own cleanup misses.

Existing destroy helpers:

- `testAccCheckWorkspaceDestroy` — `WorkspaceV1.GetWorkspace` (tenant-scoped)
- `testAccCheckBlockStorageDestroy` — `StorageV1.GetBlockStorage` (workspace-scoped)
- `testAccCheckImageDestroy` — `StorageV1.GetImage` (tenant-scoped)

**When adding a new resource:** add a matching `testAccCheck<Resource>Destroy` function and wire it into the resource's `TestCase`.

**Missing acceptance test coverage (gaps):**
- No tests for invalid configurations (expect planning errors)

## Running Tests

```bash
# Unit tests only (no TF_ACC needed)
go test -v -cover -timeout=120s -parallel=10 ./...

# Single unit test
go test -v -run TestBlockStorageToResourceModel ./internal/provider/

# Acceptance tests (requires TF_ACC=1 and live cluster)
TF_ACC=1 go test -v -cover -timeout 120m ./...

# Single acceptance test
TF_ACC=1 go test -v -run TestAccBlockStorage ./internal/acctest/
```

## Mocking Strategy

**There is no mocking.** Unit tests test pure mapping functions that have no external dependencies. Acceptance tests hit a real cluster. This is intentional:

- Mocking the SDK would couple tests to internal SDK implementation details
- The only logic worth unit-testing is the `model ↔ SDK` mapping — and that has no side effects

Do not introduce mocking frameworks (mockery, gomock, testify/mock) for this layer.

## Test Dependencies

- `github.com/stretchr/testify` — `assert` and `require`
- `github.com/hashicorp/terraform-plugin-testing` — acceptance test helpers

Do not use `terraform-plugin-sdk/v2/helper/acctest` or `terraform-plugin-sdk/v2/helper/resource` — depguard blocks them.

## What to Test When Implementing a New Resource

**Unit tests to write:**
1. `TestXxxToResourceModel` — covers the SDK→resource model mapping with a fully populated SDK object and all nullable/optional fields
2. `TestXxxToDataSourceModel` — same but for the data source model, with Status fields
3. (If `xxxFromModel` has conditional logic) — test each conditional branch

**Acceptance tests to write:**
1. `TestAccXxx` with at least:
   - Step 1: Create with all required fields; check all output attributes
   - Step 2: Add data source; verify it reads the created resource correctly
