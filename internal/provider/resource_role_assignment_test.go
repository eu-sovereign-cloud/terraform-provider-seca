package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestRoleAssignmentToResourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)
	deletedAt := createdAt.Add(2 * time.Hour)

	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "ra-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-1",
			CreatedAt:      createdAt,
			DeletedAt:      &deletedAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "platform"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.RoleAssignmentSpec{
			Subs:  []string{"sa-1", "sa-2"},
			Roles: []string{"role-1"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					Tenants:    []string{"tenant-1"},
					Regions:    []string{"region-1"},
					Workspaces: []string{"ws-1"},
				},
			},
		},
	}

	model, diags := roleAssignmentToResourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-1", model.Id.ValueString())
	assert.Equal(t, "ra-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.authorization/v1", model.ResourceProvider.ValueString())

	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, deletedAt.Format(time.RFC3339), model.DeletedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "platform"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	require.Equal(t, 2, len(model.Subs.Elements()))
	require.Equal(t, 1, len(model.Scopes.Elements()))
	require.Equal(t, 1, len(model.Roles.Elements()))
}

func TestRoleAssignmentToResourceModel_EmptyScopes(t *testing.T) {
	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "ra-empty",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-empty",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Spec: sdk.RoleAssignmentSpec{
			Subs:   []string{"sa-1"},
			Roles:  []string{"role-1"},
			Scopes: []sdk.RoleAssignmentScope{},
		},
	}

	model, diags := roleAssignmentToResourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	assert.Equal(t, 0, len(model.Scopes.Elements()))
}

func TestRoleAssignmentFromModel_RoundTrip(t *testing.T) {
	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "ra-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-1",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Spec: sdk.RoleAssignmentSpec{
			Subs:  []string{"sa-1"},
			Roles: []string{"role-1"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					Tenants:    []string{"tenant-1"},
					Regions:    []string{"region-1"},
					Workspaces: []string{"ws-1"},
				},
			},
		},
	}

	model, diags := roleAssignmentToResourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	rebuilt, diags := roleAssignmentFromModel(context.Background(), "tenant-1", model)
	require.False(t, diags.HasError())

	assert.Equal(t, "tenant-1", rebuilt.Metadata.Tenant)
	assert.Equal(t, "ra-1", rebuilt.Metadata.Name)
	require.Len(t, rebuilt.Spec.Subs, 1)
	assert.Equal(t, "sa-1", rebuilt.Spec.Subs[0])
	require.Len(t, rebuilt.Spec.Roles, 1)
	assert.Equal(t, "role-1", rebuilt.Spec.Roles[0])
	require.Len(t, rebuilt.Spec.Scopes, 1)
	assert.Equal(t, []string{"tenant-1"}, rebuilt.Spec.Scopes[0].Tenants)
	assert.Equal(t, []string{"region-1"}, rebuilt.Spec.Scopes[0].Regions)
	assert.Equal(t, []string{"ws-1"}, rebuilt.Spec.Scopes[0].Workspaces)
}

func TestRoleAssignmentFromModel_NilLists(t *testing.T) {
	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "ra-nil",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-nil",
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
		},
		Spec: sdk.RoleAssignmentSpec{
			Subs:  []string{"sa-1"},
			Roles: []string{"role-1"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					// all optional fields nil — tests nil-safe mapping
				},
			},
		},
	}

	model, diags := roleAssignmentToResourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	rebuilt, diags := roleAssignmentFromModel(context.Background(), "tenant-1", model)
	require.False(t, diags.HasError())

	require.Len(t, rebuilt.Spec.Scopes, 1)
	assert.Empty(t, rebuilt.Spec.Scopes[0].Tenants)
	assert.Empty(t, rebuilt.Spec.Scopes[0].Regions)
	assert.Empty(t, rebuilt.Spec.Scopes[0].Workspaces)
}
