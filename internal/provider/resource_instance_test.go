package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func newTestInstance() *sdk.Instance {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedAt := createdAt.Add(1 * time.Hour)

	return &sdk.Instance{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Name:           "instance-1",
			Workspace:      "workspace-1",
			Tenant:         "tenant-1",
			Region:         "region-1",
			Ref:            "seca.compute/v1/tenants/tenant-1/workspaces/workspace-1/instances/instance-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "ops"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.InstanceSpec{
			SkuRef:  sdk.Reference{Resource: "instance-skus/DXS"},
			Zone:    sdk.Zone("zone-a"),
			SshKeys: []string{"ssh-ed25519 AAAA example"},
			BootVolume: sdk.VolumeReference{
				DeviceRef: sdk.Reference{Resource: "block-storages/boot-vol"},
			},
			DataVolumes: []sdk.VolumeReference{
				{DeviceRef: sdk.Reference{Resource: "block-storages/data-vol"}},
			},
			PrimaryNicRef:     &sdk.Reference{Resource: "nics/nic-1"},
			AdditionalNicRefs: []sdk.Reference{{Resource: "nics/nic-2"}},
			SecurityGroupRef:  &sdk.Reference{Resource: "security-groups/sg-1"},
			UserData:          "#!/bin/bash\necho hello",
			AntiAffinityGroup: "aag-1",
		},
		Status: &sdk.InstanceStatus{
			State:      sdk.ResourceStateActive,
			PowerState: sdk.InstanceStatusPowerStateOn,
		},
	}
}

func TestInstanceToResourceModel(t *testing.T) {
	inst := newTestInstance()

	model, diags := instanceToResourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.compute/v1/tenants/tenant-1/workspaces/workspace-1/instances/instance-1", model.Id.ValueString())
	assert.Equal(t, "instance-1", model.Name.ValueString())
	assert.Equal(t, "workspace-1", model.WorkspaceId.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.compute/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, "instance-skus/DXS", model.SkuId.ValueString())
	assert.Equal(t, "zone-a", model.Zone.ValueString())
	assert.Equal(t, "nics/nic-1", model.PrimaryNicId.ValueString())
	assert.Equal(t, "block-storages/boot-vol", model.BootVolume.DeviceId.ValueString())
	assert.Equal(t, "security-groups/sg-1", model.SecurityGroupId.ValueString())
	assert.Equal(t, "#!/bin/bash\necho hello", model.UserData.ValueString())
	assert.Equal(t, "aag-1", model.AntiAffinityGroup.ValueString())
	assert.Equal(t, "on", model.PowerState.ValueString())
	assert.True(t, model.PowerStateSince.IsNull())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
}

func TestInstanceToResourceModel_NilStatus(t *testing.T) {
	inst := newTestInstance()
	inst.Status = nil

	model, diags := instanceToResourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.True(t, model.PowerState.IsNull())
	assert.True(t, model.PowerStateSince.IsNull())
}

func TestInstanceToResourceModel_RoundTrip(t *testing.T) {
	inst := newTestInstance()

	model, diags := instanceToResourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	// Verify data volumes list
	var dataVols []instanceVolumeModel
	diags2 := model.DataVolumes.ElementsAs(context.Background(), &dataVols, false)
	require.False(t, diags2.HasError())
	require.Len(t, dataVols, 1)
	assert.Equal(t, "block-storages/data-vol", dataVols[0].DeviceId.ValueString())

	// Verify additional nic ids
	var nicIds []string
	diags3 := model.AdditionalNicIds.ElementsAs(context.Background(), &nicIds, false)
	require.False(t, diags3.HasError())
	require.Len(t, nicIds, 1)
	assert.Equal(t, "nics/nic-2", nicIds[0])

	// Verify ssh keys
	var sshKeys []string
	diags4 := model.SshKeys.ElementsAs(context.Background(), &sshKeys, false)
	require.False(t, diags4.HasError())
	require.Len(t, sshKeys, 1)
	assert.Equal(t, "ssh-ed25519 AAAA example", sshKeys[0])
}
