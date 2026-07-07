package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestRoleToResourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "role-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/roles/role-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "core"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.RoleSpec{
			Permissions: []sdk.Permission{
				{
					Provider:  "seca.network/v1",
					Resources: []string{"networks/*", "subnets/*"},
					Verb:      []string{"get", "list"},
				},
			},
		},
	}

	model, diags := roleToResourceModel(context.Background(), role)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.authorization/v1/tenants/tenant-1/roles/role-1", model.Id.ValueString())
	assert.Equal(t, "role-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.authorization/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "core"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	require.Equal(t, 1, len(model.Permissions.Elements()))
}

func TestRoleToResourceModel_EmptyPermissions(t *testing.T) {
	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "role-empty",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/roles/role-empty",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Spec: sdk.RoleSpec{
			Permissions: []sdk.Permission{},
		},
	}

	model, diags := roleToResourceModel(context.Background(), role)
	require.False(t, diags.HasError())

	assert.Equal(t, 0, len(model.Permissions.Elements()))
}

func TestRoleFromModel_RoundTrip(t *testing.T) {
	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "role-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/roles/role-1",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Labels: sdk.Labels{"env": "dev"},
		Spec: sdk.RoleSpec{
			Permissions: []sdk.Permission{
				{
					Provider:  "seca.storage/v1",
					Resources: []string{"block-storages/*"},
					Verb:      []string{"get", "put", "delete"},
				},
			},
		},
	}

	model, diags := roleToResourceModel(context.Background(), role)
	require.False(t, diags.HasError())

	rebuilt := roleFromModel("tenant-1", model)
	assert.Equal(t, "tenant-1", rebuilt.Metadata.Tenant)
	assert.Equal(t, "role-1", rebuilt.Metadata.Name)
	require.Len(t, rebuilt.Spec.Permissions, 1)
	assert.Equal(t, "seca.storage/v1", rebuilt.Spec.Permissions[0].Provider)
	assert.Equal(t, []string{"block-storages/*"}, rebuilt.Spec.Permissions[0].Resources)
	assert.Equal(t, []string{"get", "put", "delete"}, rebuilt.Spec.Permissions[0].Verb)
}
