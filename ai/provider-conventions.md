# Provider Conventions

This document captures every naming, schema, and structural convention observed across the existing implementation. New resources and data sources must follow these conventions exactly.

## File Naming

| Type | File name pattern |
|---|---|
| Resource | `resource_<name>.go` |
| Data source | `datasource_<name>.go` |
| Resource unit test | `resource_<name>_test.go` |
| Data source unit test | `datasource_<name>_test.go` |

Acceptance tests live in `internal/acctest/` as `<name>_test.go`.

## Type Naming

| Purpose | Naming pattern | Example |
|---|---|---|
| Resource struct | `<Name>Resource` | `BlockStorageResource` |
| Data source struct | `<Name>DataSource` | `BlockStorageDataSource` |
| Resource model | `<Name>Model` | `BlockStorageModel` |
| Data source model | `<Name>DataSourceModel` | `BlockStorageDataSourceModel` |
| Constructor (resource) | `new<Name>Resource()` | `newBlockStorageResource()` |
| Constructor (data source) | `new<Name>DataSource()` | `newBlockStorageDataSource()` |
| Model→SDK mapper | `<name>FromModel(tenant, data)` | `blockStorageFromModel()` |
| SDK→resource model | `<name>ToResourceModel(ctx, sdk)` | `blockStorageToResourceModel()` |
| SDK→data source model | `<name>ToDataSourceModel(ctx, sdk)` | `blockStorageToDataSourceModel()` |

## Interface Compliance

Every resource and data source file must declare compile-time interface checks:

```go
var (
    _ resource.Resource              = (*BlockStorageResource)(nil)
    _ resource.ResourceWithConfigure = (*BlockStorageResource)(nil)
)
```

For data sources:
```go
var (
    _ datasource.DataSource              = (*BlockStorageDataSource)(nil)
    _ datasource.DataSourceWithConfigure = (*BlockStorageDataSource)(nil)
)
```

## TypeName Convention

```go
func (r *BlockStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_block_storage"
}
```

The suffix is always snake_case matching the Terraform resource name.

## Standard Schema Attributes

Every resource schema must include these attributes in this order:

```go
"id":                Computed: true  (populated from Metadata.Ref)
"name":              Required: true + RequiresReplace()  (immutable API identifier)
"tenant":            Computed: true  (from provider config, read-only)
"region":            Computed: true  (from provider config, read-only)
"resource_provider": Computed: true + UseStateForUnknown()  (e.g. "seca.storage/v1"; extracted from Metadata.Ref via refToResourceProvider())
"created_at":        Computed: true  (RFC3339 string)
"deleted_at":        Computed: true  (RFC3339 string, nullable)
"last_modified_at":  Computed: true  (RFC3339 string)
"labels":            Optional: true + Computed: true  + MapAttribute{ElementType: types.StringType}
"annotations":       Optional: true + Computed: true  + MapAttribute{ElementType: types.StringType}
"extensions":        Optional: true + Computed: true  + MapAttribute{ElementType: types.StringType}
```

`labels`, `annotations`, and `extensions` are `Optional+Computed` because the API
may return entries the config never set (for example an operation timestamp the
backend stamps into `annotations` on create). With `Optional` alone, Terraform
rejects the server-supplied value with "Provider produced inconsistent result
after apply". The trade-off is that removing one of these attributes from a
config no longer clears it — the prior value is reused; set `= {}` to clear.

Workspace-scoped resources add:
```go
"workspace_id": Required: true + RequiresReplace()
```

Data source schemas follow the same pattern with these differences:
- `resource_provider` is `Computed: true` without `UseStateForUnknown()` (data sources have no plan modifiers)
- `labels`, `annotations`, `extensions` are `Computed: true` (not Optional)
- Resource-specific status fields (e.g., `state`) are added as `Computed: true`
- `workspace_id` on data sources is `Required: true` without `RequiresReplace()`

**Exception — SKU-type schemas backed by `SkuResourceMetadata`:** SKU catalog
data sources (e.g., `seca_storage_sku`) are backed by `sdk.SkuResourceMetadata`,
which has no timestamp fields. These schemas intentionally omit `created_at`,
`deleted_at`, and `last_modified_at` — they expose only `id`, `name`, `tenant`,
`region`, plus their spec fields. Do not add timestamp attributes to SKU schemas;
the underlying metadata cannot populate them. See [glossary.md](glossary.md) for
the `SkuResourceMetadata` field list.

## Spec-required collections

When the OpenAPI schema lists a collection field under `required` **and** gives it
`minItems: 1`, the Terraform attribute must be `Required: true` plus
`Validators: []validator.List{ListSizeValidator(min, max)}` mirroring the
`minItems`/`maxItems` bounds. Declaring such a field `Optional` pushes the failure
to apply time, where the API rejects it with a 422 that the SDK collapses into
`unknow error` — the config error is invisible to the user.

Current instances:

| Attribute | Spec | Schema |
|---|---|---|
| `seca_route_table.routes` | `RouteTableSpec.routes`, minItems 1, maxItems 1000 | `Required` + `ListSizeValidator(1, 1000)` |
| `seca_nic.addresses` | `NicSpec.addresses`, minItems 1, maxItems 32 | `Required` + `ListSizeValidator(1, 32)` |

A NIC that should get its address assigned by the CSP still has to say so
explicitly — `addresses = ["0.0.0.0"]` requests automatic IPv4 assignment.

## ForceNew / RequiresReplace Rules

`stringplanmodifier.RequiresReplace()` is applied when changing the field would require destroying and recreating the resource at the API level:

- `name` — always RequiresReplace (SECA resource names are immutable identifiers)
- `workspace_id` — always RequiresReplace (resources cannot move workspaces)
- Other immutable spec fields — add RequiresReplace with a comment explaining the API constraint

Never add RequiresReplace to Computed fields or to fields that the API supports in-place updates for.

## Model Field Ordering

Within a model struct, follow this ordering convention:
1. `Id`
2. `Name`
3. `WorkspaceId` (if workspace-scoped)
4. `Tenant`, `Region`, `ResourceProvider`
5. `CreatedAt`, `DeletedAt`, `LastModifiedAt`
6. `Labels`, `Annotations`, `Extensions`
7. Resource-specific spec fields
8. Status fields (data sources only)

### Nested object fields must be nullable

A field backing a `SingleNestedAttribute` must be a **pointer** to its nested model (`BootVolume *instanceVolumeModel`), never the bare struct. `ImportState` seeds only the identity attributes, so every other attribute — including nested objects — is null in state when the framework calls `Read()`, and `Read()` reads the whole state into the model. A bare struct cannot hold null, and the import fails at runtime with:

```
Received null value, however the target type cannot handle null values.
Path: boot_volume
Target Type: provider.instanceVolumeModel
```

There is no compile-time signal for this, so `TestResourceModelsReadNullState` in `internal/provider/resource_import_state_test.go` reads an all-null state into every resource model and must be extended when a resource is added. `xxxFromModel()` must then nil-check the pointer before dereferencing it (a Required nested attribute is always set on the plan path, but the guard keeps the mapping panic-free).

Collection attributes (`ListNestedAttribute`, `MapAttribute`) do not need this — they are modelled as `types.List` / `types.Map`, which are null-aware already.

### Empty collections on Optional attributes map back to null

An `Optional`-only attribute (no `Computed`) must hold *exactly* the config value in state after apply. Nearly every collection in the SDK spec structs carries `json:"...,omitempty"`, so the API drops empty ones and reads back as a nil slice. Mapping that nil to an empty `types.List` writes `[]` into state where the config had null, and Terraform rejects it:

```
Provider produced inconsistent result after apply
.rule_refs: was null, but now cty.ListValEmpty(cty.String)
```

So `xxxToListValue()` returns `types.ListNull(...)` for a nil or empty slice — never `types.ListValueMust(t, []attr.Value{})`. The same reasoning applies to non-pointer scalars with `omitempty`: `Ports.From`/`Ports.To` arrive as `0` whether they were absent or zero, and since port 0 is never valid, `0` maps to `types.Int64Null()`.

`omitempty` also collapses `[]` and an omitted attribute into the same wire form, so a config that wrote `source_refs = []` cannot be distinguished from one that omitted it. Mapping alone cannot serve both. `preserveEmptyList()` in `types.go` reconciles the mapped value against the config value, restoring the configured representation whenever both sides are semantically empty; only values that are actually empty are touched, so real drift still reaches state. Call it in `Create()`/`Read()`/`Update()` alongside the other config restores:

```go
result.Retry = data.Retry
result.Timeouts = data.Timeouts
result.Routes = preserveEmptyList(result.Routes, data.Routes)
```

Collections nested inside a list of objects need the same treatment element-wise — see `sgPreserveRuleShape()` / `sgPreservePortsShape()` in `resource_security_group.go`, which walk the rules by position (the API echoes them back in the order they were sent) and reconcile each rule's `source_refs` and `ports.list`.

`ImportState` has no config value to reconcile against, so an imported resource always gets the null form. A config that writes `[]` still applies and plans cleanly, but `ImportStateVerify` in an acceptance test will flag the attribute — write the omitted form in acctest configs.

## Configure() Pattern

Every resource and data source Configure() must:

```go
func (r *XxxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return  // called before provider.Configure(); safe to ignore
    }

    clients, ok := req.ProviderData.(clients)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected provider data type",
            fmt.Sprintf("Expected sdk.Clients, got: %T", req.ProviderData),
        )
        return
    }

    r.client = clients.RegionalClient  // or GlobalClient for global resources
    r.tenant = clients.Tenant
    r.region = clients.Region
    r.retry = retryConfig{
        delay:       clients.RetryDelay,
        interval:    clients.RetryInterval,
        maxAttempts: clients.RetryMaxAttempts,
    }
}
```

`r.retry` (a `retryConfig`, defined in `retry.go`) holds the resolved provider-level retry values. Each resource also exposes an optional per-resource `retry` block — add `Retry *SecaRetryModel` to the model, declare it in the schema via `retryResourceSchema()`, and resolve the effective config at poll time with `r.retry.with(data.Retry)` (see [async-operations.md](async-operations.md)).

Data sources that do not perform async operations do not need the retry field or block.

## Diagnostic Error Messages

Error messages follow a two-part convention:
- **Summary**: `"Error <verb>ing <resource type>"` — e.g., `"Error creating block storage"`
- **Detail**: `"An error was encountered when <verb>ing the <resource type>.\nError: " + err.Error()`

For polling errors, the summary uses the **read** verb (the polling step is a `GetXxxUntilState` call, not the mutation), regardless of whether the caller is `Create()` or `Update()`:
- **Summary**: `"Error reading <resource type>"` — e.g., `"Error reading block storage"`
- **Detail**: `"An error was encountered while waiting for the <resource type> to become active.\nError: " + err.Error()`

## id Field Value

`id` is always set to `Metadata.Ref`, which is the full `<kind>/<name>` reference string returned by the API (e.g., `"block-storages/my-vol"`). Never set `id` to just the name.

## Scope Reference Attributes (`workspace_id`, `network_id`)

`id` holds the full URN, so a config that wires resources together with
`seca_workspace.example.id` / `seca_network.example.id` — the idiomatic Terraform
reference — puts a URN into `workspace_id` / `network_id`. The SECA API takes
those identifiers as **bare names**: they become single URL path segments and
`Metadata.Workspace` / `Metadata.Network` values (both capped at 64 characters).
A URN sent there produces a wrong URL or a validation failure.

Both forms are therefore accepted, and reconciled at the two boundaries:

- **Outbound** — every value taken from the config/state must pass through
  `workspaceName()` / `networkName()` (`types.go`) before it reaches the SDK:
  `secapi.WorkspaceID(workspaceName(data.WorkspaceId.ValueString()))`,
  `Workspace: workspaceName(data.WorkspaceId.ValueString())`. These extract the
  segment following `/workspaces/` or `/networks/` and return a bare name
  unchanged. Values read back off an API response (`block.Metadata.Workspace`)
  are already bare and need no conversion.
- **Inbound** — the value written to state must be the value the config supplied,
  verbatim. `xxxToBaseModel()` fills `WorkspaceId`/`NetworkId` from
  `Metadata.Workspace`/`Metadata.Network` (always the bare name), so every
  `Create()`/`Read()`/`Update()` restores the configured form right next to
  `result.Retry = data.Retry`, and each workspace-scoped data source `Read()`
  does the same after mapping. `workspace_id` is `Required`, so Terraform rejects
  any state value that differs from the config value ("Provider produced
  inconsistent result after apply"); a plan modifier cannot normalize it either,
  because a non-computed attribute's planned value must equal its config value.

Import IDs use bare names (`workspace-1/block-storage-1`). Importing into a
config that references `.id` therefore leaves state holding the name while the
config holds the URN — a `workspace_id` diff that forces replacement. Use the
`.name` form in configs that will be populated by import.

Attributes that map to `sdk.Reference` (`sku_id`, `source_image_id`,
`route_table_id`, `target_id`, …) put their raw value into `Reference.Resource`,
which expects the URN tail (`storage-skus/RD500`), not a full URN — the same
normalization is still missing there.

## Tenant Handling

Tenant is sourced from provider config and passed to resources via `clients.Tenant`. It is:
- Never read from the Terraform config by resources (it is always Computed)
- Always injected into the `Metadata.Tenant` field when building SDK objects in `xxxFromModel()`
- Always passed as the first argument to `xxxFromModel(tenant string, data Model)`

## Import

All resources implement `resource.ResourceWithImportState`. The import ID must carry exactly the identity fields `Read()` looks up — the tenant always comes from provider config, never from the import ID:

- **Tenant-scoped** (`seca_workspace`, `seca_image`): the ID is the resource `name`, set via `resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)`.
- **Workspace-scoped** (`seca_block_storage`): the ID is a composite `<workspace_id>/<name>`, parsed with `strings.Cut` and written with `resp.State.SetAttribute()` for each part (with a clear diagnostic if the format is wrong).

Do **not** passthrough to `id`: `Read()` keys off `name`/`workspace_id`, not the `Metadata.Ref` stored in `id` (see [id Field Value](#id-field-value)). Name the `ImportState` receiver `r`, not `resource`, so the body can reach `resource.ImportStatePassthroughID` (the usual `resource` receiver name shadows the package).

### Documenting the import ID format

Import docs are generated by `tfplugindocs` (via `make generate`) — never hand-edit `docs/`. Each resource's import ID format is documented by two example files that the generator renders into the resource's **## Import** section:

- `examples/resources/<type>/import.sh` — the CLI `terraform import` command
- `examples/resources/<type>/import-by-string-id.tf` — the config-driven `import {}` block with the `id` attribute (Terraform ≥ 1.5). Note the exact filename: `tfplugindocs` ignores a plain `import.tf`.

Keep the IDs in those examples in sync with the `ImportState` implementation.
