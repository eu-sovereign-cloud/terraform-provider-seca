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
		Annotations: sdk.Annotations{"owner": "team-a"},
		Extensions:  sdk.Extensions{"ext": "v1"},
		Spec: sdk.RoleAssignmentSpec{
			Subs: []string{"user-1", "user-2"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					Tenants:    []string{"tenant-1"},
					Regions:    []string{"region-1"},
					Workspaces: []string{"workspace-1"},
				},
			},
			Roles: []string{"role-1", "role-2"},
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
	assert.Equal(t, map[string]string{"owner": "team-a"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v1"}, toStringMap(model.Extensions))

	require.Equal(t, 2, len(model.Subs.Elements()))
	require.Equal(t, 1, len(model.Scopes.Elements()))
	require.Equal(t, 2, len(model.Roles.Elements()))
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
			Subs: []string{"svc-account-1"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					Tenants:    []string{"tenant-1"},
					Regions:    []string{"region-eu-1"},
					Workspaces: []string{"ws-prod"},
				},
			},
			Roles: []string{"admin-role"},
		},
	}

	model, diags := roleAssignmentToResourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	rebuilt := roleAssignmentFromModel("tenant-1", model)
	assert.Equal(t, "tenant-1", rebuilt.Metadata.Tenant)
	assert.Equal(t, "ra-1", rebuilt.Metadata.Name)
	require.Len(t, rebuilt.Spec.Subs, 1)
	assert.Equal(t, "svc-account-1", rebuilt.Spec.Subs[0])
	require.Len(t, rebuilt.Spec.Scopes, 1)
	assert.Equal(t, []string{"tenant-1"}, rebuilt.Spec.Scopes[0].Tenants)
	assert.Equal(t, []string{"region-eu-1"}, rebuilt.Spec.Scopes[0].Regions)
	assert.Equal(t, []string{"ws-prod"}, rebuilt.Spec.Scopes[0].Workspaces)
	require.Len(t, rebuilt.Spec.Roles, 1)
	assert.Equal(t, "admin-role", rebuilt.Spec.Roles[0])
}
