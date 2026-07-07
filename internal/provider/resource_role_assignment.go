package provider

import (
	"context"
	"fmt"

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
	_ resource.Resource                = (*RoleAssignmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*RoleAssignmentResource)(nil)
	_ resource.ResourceWithImportState = (*RoleAssignmentResource)(nil)
)

type RoleAssignmentResource struct {
	client *secapi.GlobalClient
	tenant string
	retry  retryConfig
}

func newRoleAssignmentResource() resource.Resource {
	return &RoleAssignmentResource{}
}

func (r *RoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_assignment"
}

type RoleAssignmentResourceModel struct {
	roleAssignmentModel
	Retry *RetryModel `tfsdk:"retry"`
}

func (r *RoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
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
			"subs": tfschema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"scopes": tfschema.ListNestedAttribute{
				Required: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"tenants": tfschema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"regions": tfschema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"workspaces": tfschema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
			"roles": tfschema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *RoleAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured role assignment resource")
}

func (r *RoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	if name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier to be the role assignment name, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func (r *RoleAssignmentResource) logFields(ctx context.Context, data RoleAssignmentResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *RoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating role assignment")

	ra := roleAssignmentFromModel(r.tenant, data)

	ra, err := r.client.AuthorizationV1.CreateOrUpdateRoleAssignment(ctx, ra)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role assignment",
			"An error was encountered when creating the role assignment.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role assignment to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(ra.Metadata.Tenant),
		Name:   ra.Metadata.Name,
	}

	config := r.retry.with(data.Retry).untilState(sdk.ResourceStateActive)

	ra, err = r.client.AuthorizationV1.GetRoleAssignmentUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role assignment",
			"An error was encountered while waiting for the role assignment to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleAssignmentToResourceModel(ctx, ra)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry

	tflog.Info(ctx, "role assignment created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading role assignment")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(r.tenant),
		Name:   data.Name.ValueString(),
	}

	ra, err := r.client.AuthorizationV1.GetRoleAssignment(ctx, tref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "role assignment not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role assignment",
			"An error was encountered when reading the role assignment.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleAssignmentToResourceModel(ctx, ra)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RoleAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating role assignment")

	ra := roleAssignmentFromModel(r.tenant, data)

	ra, err := r.client.AuthorizationV1.CreateOrUpdateRoleAssignment(ctx, ra)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role assignment",
			"An error was encountered when updating the role assignment.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role assignment to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(ra.Metadata.Tenant),
		Name:   ra.Metadata.Name,
	}

	config := r.retry.with(data.Retry).untilState(sdk.ResourceStateActive)

	ra, err = r.client.AuthorizationV1.GetRoleAssignmentUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role assignment",
			"An error was encountered while waiting for the role assignment to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleAssignmentToResourceModel(ctx, ra)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result.Retry = data.Retry

	tflog.Info(ctx, "role assignment updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *RoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting role assignment")

	ra := &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Tenant: r.tenant,
			Name:   data.Name.ValueString(),
		},
	}

	err := r.client.AuthorizationV1.DeleteRoleAssignment(ctx, ra)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting role assignment",
			"An error was encountered when deleting the role assignment.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for role assignment to be deleted")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(r.tenant),
		Name:   data.Name.ValueString(),
	}

	config := r.retry.with(data.Retry).observer()

	err = r.client.AuthorizationV1.WatchRoleAssignmentUntilDeleted(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role assignment",
			"An error was encountered while waiting for the role assignment to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "role assignment deleted")
}

func roleAssignmentFromModel(tenant string, data RoleAssignmentResourceModel) *sdk.RoleAssignment {
	var subs []string
	if !data.Subs.IsNull() && !data.Subs.IsUnknown() {
		data.Subs.ElementsAs(context.Background(), &subs, false)
	}

	var roles []string
	if !data.Roles.IsNull() && !data.Roles.IsUnknown() {
		data.Roles.ElementsAs(context.Background(), &roles, false)
	}

	var scopes []sdk.RoleAssignmentScope
	if !data.Scopes.IsNull() && !data.Scopes.IsUnknown() {
		scopeModels := make([]scopeModel, 0, len(data.Scopes.Elements()))
		data.Scopes.ElementsAs(context.Background(), &scopeModels, false)
		scopes = make([]sdk.RoleAssignmentScope, 0, len(scopeModels))
		for _, s := range scopeModels {
			var tenants []string
			if !s.Tenants.IsNull() && !s.Tenants.IsUnknown() {
				s.Tenants.ElementsAs(context.Background(), &tenants, false)
			}
			var regions []string
			if !s.Regions.IsNull() && !s.Regions.IsUnknown() {
				s.Regions.ElementsAs(context.Background(), &regions, false)
			}
			var workspaces []string
			if !s.Workspaces.IsNull() && !s.Workspaces.IsUnknown() {
				s.Workspaces.ElementsAs(context.Background(), &workspaces, false)
			}
			scopes = append(scopes, sdk.RoleAssignmentScope{
				Tenants:    tenants,
				Regions:    regions,
				Workspaces: workspaces,
			})
		}
	}

	return &sdk.RoleAssignment{
		Metadata: &sdk.GlobalTenantResourceMetadata{
			Tenant: tenant,
			Name:   data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.RoleAssignmentSpec{
			Subs:   subs,
			Scopes: scopes,
			Roles:  roles,
		},
	}
}

func roleAssignmentToResourceModel(ctx context.Context, ra *sdk.RoleAssignment) (RoleAssignmentResourceModel, diag.Diagnostics) {
	base, diags := roleAssignmentToBaseModel(ctx, ra)
	return RoleAssignmentResourceModel{roleAssignmentModel: base}, diags
}
