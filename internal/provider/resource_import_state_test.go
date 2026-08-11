package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// TestResourceModelsReadNullState guards the resource models against the null
// values that ImportState leaves behind. ImportState only seeds the identifying
// attributes; every other attribute is null in state when the framework calls
// Read, and Read reads the whole state into the model. A model field that
// cannot hold null — a plain nested struct rather than a pointer or a
// types.Object — fails there with "Received null value, however the target type
// cannot handle null values", which surfaces as an import error and never as a
// compile error.
func TestResourceModelsReadNullState(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		resource func() resource.Resource
		model    any
	}{
		{"workspace", newWorkspaceResource, &WorkspaceResourceModel{}},
		{"image", newImageResource, &ImageResourceModel{}},
		{"block_storage", newBlockStorageResource, &BlockStorageResourceModel{}},
		{"network", newNetworkResource, &NetworkResourceModel{}},
		{"internet_gateway", newInternetGatewayResource, &InternetGatewayResourceModel{}},
		{"role", newRoleResource, &RoleResourceModel{}},
		{"role_assignment", newRoleAssignmentResource, &RoleAssignmentResourceModel{}},
		{"route_table", newRouteTableResource, &RouteTableResourceModel{}},
		{"subnet", newSubnetResource, &SubnetResourceModel{}},
		{"security_group", newSecurityGroupResource, &SecurityGroupResourceModel{}},
		{"public_ip", newPublicIpResource, &PublicIpResourceModel{}},
		{"nic", newNicResource, &NicResourceModel{}},
		{"instance", newInstanceResource, &InstanceResourceModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &resource.SchemaResponse{}
			tt.resource().Schema(ctx, resource.SchemaRequest{}, resp)
			require.False(t, resp.Diagnostics.HasError())

			state := tfsdk.State{
				Schema: resp.Schema,
				Raw:    nullObjectValue(resp.Schema.Type().TerraformType(ctx)),
			}

			diags := state.Get(ctx, tt.model)
			require.False(t, diags.HasError(), "reading a null state failed: %v", diags.Errors())
		})
	}
}

// nullObjectValue builds an object value of the given type whose every
// attribute is null, mirroring the state a resource is read with right after
// ImportState.
func nullObjectValue(typ tftypes.Type) tftypes.Value {
	obj, ok := typ.(tftypes.Object)
	if !ok {
		return tftypes.NewValue(typ, nil)
	}

	attrs := make(map[string]tftypes.Value, len(obj.AttributeTypes))
	for name, attrType := range obj.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	return tftypes.NewValue(obj, attrs)
}
