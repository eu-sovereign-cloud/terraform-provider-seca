package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

// TestInstanceToResourceModel_PowerState verifies that power_state and
// power_state_since are correctly populated from the API response.
//
// Design decision (GA-29, Option A): power_state and power_state_since are
// read-only Computed attributes. They reflect the live power state returned by
// the API and are never directly settable by the operator; there is no desired-state
// attribute. This follows the aws_instance precedent. Note that Create() still starts
// the instance automatically (see InstanceResource.Create), since a newly created
// instance is not considered provisioned until it reaches power state "on".
func TestInstanceToResourceModel_PowerState(t *testing.T) {
	powerStateSince := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	inst := newTestInstance()
	inst.Status = &sdk.InstanceStatus{
		State:           sdk.ResourceStateActive,
		PowerState:      sdk.InstanceStatusPowerStateOn,
		PowerStateSince: &powerStateSince,
	}

	model, diags := instanceToResourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.Equal(t, "on", model.PowerState.ValueString())
	assert.Equal(t, "2024-06-01T12:00:00Z", model.PowerStateSince.ValueString())
}

func TestInstanceToResourceModel_PowerStateOff(t *testing.T) {
	inst := newTestInstance()
	inst.Status = &sdk.InstanceStatus{
		State:      sdk.ResourceStateActive,
		PowerState: sdk.InstanceStatusPowerStateOff,
	}

	model, diags := instanceToResourceModel(context.Background(), inst)
	require.False(t, diags.HasError())

	assert.Equal(t, "off", model.PowerState.ValueString())
	assert.True(t, model.PowerStateSince.IsNull())
}
