package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestRoleToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)

	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "role-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/roles/role-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "platform"},
		Extensions:  sdk.Extensions{"ext": "v2"},
		Spec: sdk.RoleSpec{
			Permissions: []sdk.Permission{
				{
					Provider:  "seca.network/v1",
					Resources: []string{"networks/*"},
					Verb:      []string{"get"},
				},
			},
		},
		Status: &sdk.RoleStatus{State: sdk.ResourceStateActive},
	}

	model, diags := roleToDataSourceModel(context.Background(), role)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.authorization/v1/tenants/tenant-1/roles/role-1", model.Id.ValueString())
	assert.Equal(t, "role-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.authorization/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "platform"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v2"}, toStringMap(model.Extensions))

	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
	require.Equal(t, 1, len(model.Permissions.Elements()))
}

func TestRoleToDataSourceModel_NilStatus(t *testing.T) {
	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "role-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/roles/role-1",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Spec: sdk.RoleSpec{
			Permissions: []sdk.Permission{
				{
					Provider:  "seca.storage/v1",
					Resources: []string{"block-storages/*"},
					Verb:      []string{"get"},
				},
			},
		},
		Status: nil,
	}

	model, diags := roleToDataSourceModel(context.Background(), role)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
}
