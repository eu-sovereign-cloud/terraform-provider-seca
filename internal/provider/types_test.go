package provider

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestPreserveEmptyList(t *testing.T) {
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	null := types.ListNull(types.StringType)
	filled := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})

	// A config that wrote `[]` keeps `[]`; one that omitted the attribute keeps null.
	assert.Equal(t, empty, preserveEmptyList(null, empty))
	assert.Equal(t, null, preserveEmptyList(null, null))
	// A non-empty value on either side is a real value and passes through.
	assert.Equal(t, filled, preserveEmptyList(filled, empty))
	assert.Equal(t, null, preserveEmptyList(null, filled))
	assert.Equal(t, null, preserveEmptyList(null, types.ListUnknown(types.StringType)))
}

func TestNumberToDuration(t *testing.T) {
	tests := []struct {
		name string
		in   types.Number
		want time.Duration
	}{
		{"null", types.NumberNull(), 0},
		{"unknown", types.NumberUnknown(), 0},
		{"zero", types.NumberValue(big.NewFloat(0)), 0},
		{"whole seconds", types.NumberValue(big.NewFloat(30)), 30 * time.Second},
		{"fractional seconds", types.NumberValue(big.NewFloat(1.5)), 1500 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, numberToDuration(tt.in))
		})
	}
}

func TestNumberToInt(t *testing.T) {
	tests := []struct {
		name string
		in   types.Number
		want int
	}{
		{"null", types.NumberNull(), 0},
		{"unknown", types.NumberUnknown(), 0},
		{"zero", types.NumberValue(big.NewFloat(0)), 0},
		{"positive", types.NumberValue(big.NewFloat(42)), 42},
		{"negative", types.NumberValue(big.NewFloat(-7)), -7},
		{"truncates fraction", types.NumberValue(big.NewFloat(9.9)), 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, numberToInt(tt.in))
		})
	}
}

func TestToStringMap(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		assert.Nil(t, toStringMap(types.MapNull(types.StringType)))
	})

	t.Run("unknown", func(t *testing.T) {
		assert.Nil(t, toStringMap(types.MapUnknown(types.StringType)))
	})

	t.Run("empty", func(t *testing.T) {
		m, diags := types.MapValueFrom(context.Background(), types.StringType, map[string]string{})
		require.False(t, diags.HasError())
		assert.Empty(t, toStringMap(m))
	})

	t.Run("populated", func(t *testing.T) {
		want := map[string]string{"env": "prod", "team": "core"}
		m, diags := types.MapValueFrom(context.Background(), types.StringType, want)
		require.False(t, diags.HasError())
		assert.Equal(t, want, toStringMap(m))
	})
}

func TestFromStringMap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		m, diags := fromStringMap(context.Background(), nil)
		require.False(t, diags.HasError())
		assert.True(t, m.IsNull())
	})

	t.Run("empty", func(t *testing.T) {
		m, diags := fromStringMap(context.Background(), map[string]string{})
		require.False(t, diags.HasError())
		assert.True(t, m.IsNull())
	})

	t.Run("populated", func(t *testing.T) {
		in := map[string]string{"env": "prod", "team": "core"}
		m, diags := fromStringMap(context.Background(), in)
		require.False(t, diags.HasError())
		assert.False(t, m.IsNull())
		assert.Equal(t, in, toStringMap(m))
	})
}

func TestFromTime(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		assert.True(t, fromTime(time.Time{}).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ts := time.Date(2026, 6, 23, 10, 30, 0, 0, time.UTC)
		got := fromTime(ts)
		assert.False(t, got.IsNull())
		assert.Equal(t, ts.Format(time.RFC3339), got.ValueString())
	})
}

func TestFromTimePtr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.True(t, fromTimePtr(nil).IsNull())
	})

	t.Run("zero value", func(t *testing.T) {
		ts := time.Time{}
		assert.True(t, fromTimePtr(&ts).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ts := time.Date(2026, 6, 23, 10, 30, 0, 0, time.UTC)
		got := fromTimePtr(&ts)
		assert.False(t, got.IsNull())
		assert.Equal(t, ts.Format(time.RFC3339), got.ValueString())
	})
}

func TestFromRefPtr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.True(t, fromRefPtr(nil).IsNull())
	})

	t.Run("value", func(t *testing.T) {
		ref := &sdk.Reference{Resource: "some-resource"}
		got := fromRefPtr(ref)
		assert.False(t, got.IsNull())
		assert.Equal(t, "some-resource", got.ValueString())
	})
}

func TestRefToResourceProvider(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
		null bool
	}{
		{
			name: "empty string",
			ref:  "",
			null: true,
		},
		{
			name: "workspace ref",
			ref:  "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1",
			want: "seca.workspace/v1",
		},
		{
			name: "storage ref (image)",
			ref:  "seca.storage/v1/tenants/tenant-1/images/image-1",
			want: "seca.storage/v1",
		},
		{
			name: "storage ref (block storage)",
			ref:  "seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/bs-1",
			want: "seca.storage/v1",
		},
		{
			name: "region ref",
			ref:  "seca.region/v1/regions/region-1",
			want: "seca.region/v1",
		},
		{
			name: "storage sku ref",
			ref:  "seca.storage/v1/tenants/tenant-1/storage-skus/RD500",
			want: "seca.storage/v1",
		},
		{
			name: "no slash — returns as-is",
			ref:  "noslash",
			want: "noslash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refToResourceProvider(tt.ref)
			if tt.null {
				assert.True(t, got.IsNull())
			} else {
				assert.False(t, got.IsNull())
				assert.Equal(t, tt.want, got.ValueString())
			}
		})
	}
}

func TestWorkspaceName(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "bare name",
			id:   "workspace-1",
			want: "workspace-1",
		},
		{
			name: "workspace urn",
			id:   "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1",
			want: "workspace-1",
		},
		{
			name: "workspace-scoped resource urn",
			id:   "seca.storage/v1/tenants/tenant-1/workspaces/workspace-1/block-storages/bs-1",
			want: "workspace-1",
		},
		{
			name: "empty",
			id:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, workspaceName(tt.id))
		})
	}
}

func TestNetworkName(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "bare name",
			id:   "network-1",
			want: "network-1",
		},
		{
			name: "network urn",
			id:   "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1",
			want: "network-1",
		},
		{
			name: "network-scoped resource urn",
			id:   "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1/subnets/subnet-1",
			want: "network-1",
		},
		{
			name: "empty",
			id:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, networkName(tt.id))
		})
	}
}

func TestCutImportName(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantScope string
		wantName  string
		wantOk    bool
	}{
		{
			name:      "bare workspace name",
			id:        "workspace-1/block-storage-1",
			wantScope: "workspace-1",
			wantName:  "block-storage-1",
			wantOk:    true,
		},
		{
			name:      "workspace urn",
			id:        "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1/block-storage-1",
			wantScope: "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1",
			wantName:  "block-storage-1",
			wantOk:    true,
		},
		{
			name:   "no separator",
			id:     "block-storage-1",
			wantOk: false,
		},
		{
			name:   "trailing separator",
			id:     "workspace-1/",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, name, ok := cutImportName(tt.id)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantScope, scope)
				assert.Equal(t, tt.wantName, name)
			}
		})
	}
}

func TestCutImportScope(t *testing.T) {
	workspaceURN := "seca.workspace/v1/tenants/tenant-1/workspaces/workspace-1"
	networkURN := "seca.network/v1/tenants/tenant-1/workspaces/workspace-1/networks/network-1"

	tests := []struct {
		name        string
		id          string
		wantPrefix  string
		wantScopeID string
		wantOk      bool
	}{
		{
			name:        "both bare names",
			id:          "workspace-1/network-1",
			wantPrefix:  "workspace-1",
			wantScopeID: "network-1",
			wantOk:      true,
		},
		{
			name:        "both urns",
			id:          workspaceURN + "/" + networkURN,
			wantPrefix:  workspaceURN,
			wantScopeID: networkURN,
			wantOk:      true,
		},
		{
			name:        "workspace urn, bare network name",
			id:          workspaceURN + "/network-1",
			wantPrefix:  workspaceURN,
			wantScopeID: "network-1",
			wantOk:      true,
		},
		{
			name:        "bare workspace name, network urn",
			id:          "workspace-1/" + networkURN,
			wantPrefix:  "workspace-1",
			wantScopeID: networkURN,
			wantOk:      true,
		},
		{
			name:   "no separator",
			id:     "network-1",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, scopeID, ok := cutImportScope(tt.id, "networks")
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantPrefix, prefix)
				assert.Equal(t, tt.wantScopeID, scopeID)
			}
		})
	}
}
