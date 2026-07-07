package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestInstanceSkuToDataSourceModel(t *testing.T) {
	sku := &sdk.InstanceSku{
		Metadata: &sdk.SkuResourceMetadata{
			Name:   "DXS",
			Tenant: "tenant-1",
			Region: "region-1",
			Ref:    "seca.compute/v1/tenants/tenant-1/instance-skus/DXS",
		},
		Labels:      sdk.Labels{"tier": "gold"},
		Annotations: sdk.Annotations{"team": "compute"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: &sdk.InstanceSkuSpec{
			VCPU: 4,
			Ram:  16,
		},
	}

	model, diags := instanceSkuToDataSourceModel(context.Background(), sku)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.compute/v1/tenants/tenant-1/instance-skus/DXS", model.Id.ValueString())
	assert.Equal(t, "DXS", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "region-1", model.Region.ValueString())
	assert.Equal(t, "seca.compute/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, map[string]string{"tier": "gold"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "compute"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	assert.Equal(t, int64(4), model.Vcpus.ValueInt64())
	assert.Equal(t, int64(16), model.Ram.ValueInt64())
}
