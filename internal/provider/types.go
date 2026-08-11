package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func numberToDuration(n types.Number) time.Duration {
	if n.IsNull() || n.IsUnknown() {
		return 0
	}
	seconds, _ := n.ValueBigFloat().Float64()
	return time.Duration(seconds * float64(time.Second))
}

func numberToInt(n types.Number) int {
	if n.IsNull() || n.IsUnknown() {
		return 0
	}
	value, _ := n.ValueBigFloat().Int64()
	return int(value)
}

func toStringMap(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok {
			result[k] = s.ValueString()
		}
	}
	return result
}

func fromStringMap(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if len(m) == 0 {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

// preserveEmptyList keeps the configured representation of a semantically empty
// list. An Optional-only attribute has to hold exactly the config value in state
// after apply, but the SDK spec fields backing these lists are `omitempty`,
// which collapses `[]` and an omitted attribute into the same wire form. The
// mappers read both back as null, so a config that wrote `[]` would otherwise
// fail with "Provider produced inconsistent result after apply".
//
// Only values that are already empty on both sides are substituted — real drift
// reaches state as the API reported it. Call it from Create/Read/Update
// alongside the other config restores (`result.Retry = data.Retry`).
func preserveEmptyList(mapped, configured types.List) types.List {
	if configured.IsUnknown() || len(mapped.Elements()) > 0 || len(configured.Elements()) > 0 {
		return mapped
	}

	return configured
}

func fromTime(t time.Time) types.String {
	if t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

func fromTimePtr(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return fromTime(*t)
}

func fromRefPtr(ref *sdk.Reference) types.String {
	if ref == nil {
		return types.StringNull()
	}
	return types.StringValue(ref.Resource)
}

// refToResourceProvider extracts the "{provider}/{version}" prefix from a
// full resource Ref URN (e.g. "seca.storage/v1" from
// "seca.storage/v1/tenants/t1/workspaces/w1/block-storages/bs1").
func refToResourceProvider(ref string) types.String {
	if ref == "" {
		return types.StringNull()
	}
	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 2 {
		return types.StringValue(ref)
	}
	return types.StringValue(parts[0] + "/" + parts[1])
}

// urnName extracts the name of the resource held in the given URN collection
// segment, e.g. urnName("seca.workspace/v1/tenants/t1/workspaces/w1", "workspaces")
// returns "w1". A value that is already a bare name is returned unchanged.
//
// Reference attributes (workspace_id, network_id, …) accept either form: a config
// may point them at a sibling resource's `id` — which always carries the full URN
// — or at its `name`. The SECA API takes these identifiers as single URL path
// segments (and as `Metadata.Workspace` / `Metadata.Network`, both capped at 64
// characters), so they must be reduced to the bare name before reaching the SDK.
func urnName(id string, collection string) string {
	_, after, found := strings.Cut(id, "/"+collection+"/")
	if !found {
		return id
	}
	name, _, _ := strings.Cut(after, "/")
	return name
}

// workspaceName reduces a workspace_id attribute value to the bare workspace name.
func workspaceName(id string) string {
	return urnName(id, "workspaces")
}

// networkName reduces a network_id attribute value to the bare network name.
func networkName(id string) string {
	return urnName(id, "networks")
}

// urnStart returns the index at which the URN that ends s begins, or -1 when s
// does not end in a URN. A URN is "{provider}/{version}/tenants/{tenant}/…", so
// the last "/tenants/" marks its third segment and the provider and version
// segments precede it.
func urnStart(s string) int {
	tenants := strings.LastIndex(s, "/tenants/")
	if tenants < 0 {
		return -1
	}
	version := strings.LastIndex(s[:tenants], "/")
	if version < 0 {
		return -1
	}
	return strings.LastIndex(s[:version], "/") + 1
}

// cutImportName splits the resource name off the right of a composite import
// identifier. Names never contain "/", so the name is always the last segment;
// what precedes it is the scope, which may itself be a URN.
func cutImportName(id string) (scope string, name string, ok bool) {
	i := strings.LastIndex(id, "/")
	if i <= 0 || i == len(id)-1 {
		return "", "", false
	}
	return id[:i], id[i+1:], true
}

// cutImportScope splits the trailing scope identifier off a composite import
// identifier, e.g. the network_id out of "<workspace_id>/<network_id>". The
// scope may be written as a bare name or as a full URN ending in
// "{collection}/{name}" — the latter contains "/" itself, so the identifier
// cannot simply be split on every "/".
func cutImportScope(id string, collection string) (prefix string, scopeID string, ok bool) {
	segments := strings.Split(id, "/")
	if len(segments) < 2 {
		return "", "", false
	}

	if segments[len(segments)-2] == collection {
		start := urnStart(id)
		if start <= 0 {
			return "", "", false
		}
		return id[:start-1], id[start:], true
	}

	i := strings.LastIndex(id, "/")
	if i <= 0 || i == len(id)-1 {
		return "", "", false
	}
	return id[:i], id[i+1:], true
}

// fromNonEmptyString converts an empty string to null, or wraps a non-empty
// string as a types.String value. Used for optional SDK string fields that
// use "" to represent "not set".
func fromNonEmptyString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
