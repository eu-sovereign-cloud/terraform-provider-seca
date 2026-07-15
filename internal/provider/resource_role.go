package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*RoleResource)(nil)
	_ resource.ResourceWithConfigure   = (*RoleResource)(nil)
	_ resource.ResourceWithImportState = (*RoleResource)(nil)
)

type RoleResource struct {
	client *secapi.GlobalClient
	tenant string
	retry  retryConfig
}

func newRoleResource() resource.Resource {
	return &RoleResource{}
}

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

type RoleResourceModel struct {
	roleModel
	Retry    *RetryModel    `tfsdk:"retry"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *RoleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Blocks: map[string]tfschema.Block{
			"timeouts": timeouts.BlockAll(ctx),
		},
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_provider": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deleted_at": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_at": tfschema.StringAttribute{
				Computed: true,
			},
			"labels": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"annotations": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"extensions": tfschema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"permissions": tfschema.ListNestedAttribute{
				Required: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"provider": tfschema.StringAttribute{
							Required: true,
						},
						"resources": tfschema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"verb": tfschema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected sdk.Clients, got: %T", req.ProviderData),
		)
		return
	}

	r.client = clients.GlobalClient
	r.tenant = clients.Tenant
	r.retry = retryConfig{
		delay:       clients.RetryDelay,
		interval:    clients.RetryInterval,
		maxAttempts: clients.RetryMaxAttempts,
	}

	tflog.Debug(ctx, "configured role resource")
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	if name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier to be the role name, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func (r *RoleResource) logFields(ctx context.Context, data RoleResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating role")

	role := roleFromModel(r.tenant, data)

	role, err := r.client.AuthorizationV1.CreateOrUpdateRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role",
			"An error was encountered when creating the role.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(role.Metadata.Tenant),
		Name:   role.Metadata.Name,
	}

	config := r.retry.with(data.Retry).withTimeout(createTimeout).untilState(sdk.ResourceStateActive)

	role, err = r.client.AuthorizationV1.GetRoleUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"An error was encountered while waiting for the role to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags2 := roleToResourceModel(ctx, role)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry
	result.Timeouts = data.Timeouts

	tflog.Info(ctx, "role created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading role")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(r.tenant),
		Name:   data.Name.ValueString(),
	}

	role, err := r.client.AuthorizationV1.GetRole(ctx, tref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "role not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"An error was encountered when reading the role.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleToResourceModel(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating role")

	role := roleFromModel(r.tenant, data)

	role, err := r.client.AuthorizationV1.CreateOrUpdateRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role",
			"An error was encountered when updating the role.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(role.Metadata.Tenant),
		Name:   role.Metadata.Name,
	}

	config := r.retry.with(data.Retry).withTimeout(updateTimeout).untilState(sdk.ResourceStateActive)

	role, err = r.client.AuthorizationV1.GetRoleUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"An error was encountered while waiting for the role to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags2 := roleToResourceModel(ctx, role)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry
	result.Timeouts = data.Timeouts

	tflog.Info(ctx, "role updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting role")

	role := &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Tenant: r.tenant,
			Name:   data.Name.ValueString(),
		},
	}

	err := r.client.AuthorizationV1.DeleteRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting role",
			"An error was encountered when deleting the role.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role to be deleted")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(r.tenant),
		Name:   data.Name.ValueString(),
	}

	config := r.retry.with(data.Retry).withTimeout(deleteTimeout).observer()

	err = r.client.AuthorizationV1.WatchRoleUntilDeleted(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"An error was encountered while waiting for the role to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "role deleted")
}

func roleFromModel(tenant string, data RoleResourceModel) *sdk.Role {
	var perms []sdk.Permission
	if !data.Permissions.IsNull() && !data.Permissions.IsUnknown() {
		permModels := make([]permissionModel, 0, len(data.Permissions.Elements()))
		data.Permissions.ElementsAs(context.Background(), &permModels, false)
		perms = make([]sdk.Permission, 0, len(permModels))
		for _, p := range permModels {
			var resources []string
			p.Resources.ElementsAs(context.Background(), &resources, false)
			var verb []string
			p.Verb.ElementsAs(context.Background(), &verb, false)
			perms = append(perms, sdk.Permission{
				Provider:  p.Provider.ValueString(),
				Resources: resources,
				Verb:      verb,
			})
		}
	}

	return &sdk.Role{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Tenant: tenant,
			Name:   data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.RoleSpec{
			Permissions: perms,
		},
	}
}

func roleToResourceModel(ctx context.Context, role *sdk.Role) (RoleResourceModel, diag.Diagnostics) {
	base, diags := roleToBaseModel(ctx, role)
	return RoleResourceModel{roleModel: base}, diags
}
