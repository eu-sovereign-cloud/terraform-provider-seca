package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestInstanceToDataSourceModel(t *testing.T) {
	inst := newTestInstance()

	model, diags := instanceToDataSourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.compute/v1/tenants/tenant-1/workspaces/workspace-1/instances/instance-1", model.Id.ValueString())
	assert.Equal(t, "instance-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.compute/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, "instance-skus/DXS", model.SkuId.ValueString())
	assert.Equal(t, "zone-a", model.Zone.ValueString())
	assert.Equal(t, "block-storages/boot-vol", model.BootVolume.DeviceId.ValueString())
	assert.Equal(t, "on", model.PowerState.ValueString())
	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
}

func TestInstanceToDataSourceModel_NilStatus(t *testing.T) {
	inst := newTestInstance()
	inst.Status = nil

	model, diags := instanceToDataSourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.True(t, model.PowerState.IsNull())
	assert.True(t, model.PowerStateSince.IsNull())
	assert.True(t, model.State.IsNull())
}
