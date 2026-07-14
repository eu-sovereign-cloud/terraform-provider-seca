package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

func TestRoleAssignmentToDataSourceModel(t *testing.T) {
	createdAt := time.Now()
	modifiedAt := createdAt.Add(1 * time.Hour)

	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Name:           "ra-1",
			Tenant:         "tenant-1",
			Ref:            "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-1",
			CreatedAt:      createdAt,
			LastModifiedAt: modifiedAt,
		},
		Labels:      sdk.Labels{"env": "prod"},
		Annotations: sdk.Annotations{"team": "platform"},
		Extensions:  sdk.Extensions{"ext": "v2"},
		Spec: sdk.RoleAssignmentSpec{
			Subs:  []string{"sa-1"},
			Roles: []string{"role-admin"},
			Scopes: []sdk.RoleAssignmentScope{
				{
					Tenants:    []string{"tenant-1"},
					Workspaces: []string{"ws-1", "ws-2"},
				},
			},
		},
		Status: &sdk.RoleAssignmentStatus{State: sdk.ResourceStateActive},
	}

	model, diags := roleAssignmentToDataSourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	assert.Equal(t, "seca.authorization/v1/tenants/tenant-1/role-assignments/ra-1", model.Id.ValueString())
	assert.Equal(t, "ra-1", model.Name.ValueString())
	assert.Equal(t, "tenant-1", model.Tenant.ValueString())
	assert.Equal(t, "seca.authorization/v1", model.ResourceProvider.ValueString())
	assert.Equal(t, createdAt.Format(time.RFC3339), model.CreatedAt.ValueString())
	assert.Equal(t, modifiedAt.Format(time.RFC3339), model.LastModifiedAt.ValueString())

	assert.Equal(t, map[string]string{"env": "prod"}, toStringMap(model.Labels))
	assert.Equal(t, map[string]string{"team": "platform"}, toStringMap(model.Annotations))
	assert.Equal(t, map[string]string{"ext": "v2"}, toStringMap(model.Extensions))

	assert.Equal(t, string(sdk.ResourceStateActive), model.State.ValueString())
	require.Equal(t, 1, len(model.Subs.Elements()))
	require.Equal(t, 1, len(model.Scopes.Elements()))
	require.Equal(t, 1, len(model.Roles.Elements()))
}

func TestRoleAssignmentToDataSourceModel_NilStatus(t *testing.T) {
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
				{Tenants: []string{"tenant-1"}},
			},
		},
		Status: nil,
	}

	model, diags := roleAssignmentToDataSourceModel(context.Background(), ra)
	require.False(t, diags.HasError())

	assert.True(t, model.State.IsNull())
}
