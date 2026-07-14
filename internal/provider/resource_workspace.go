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
	_ resource.Resource                = (*WorkspaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*WorkspaceResource)(nil)
	_ resource.ResourceWithImportState = (*WorkspaceResource)(nil)
)

type WorkspaceResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

func (resource *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

type WorkspaceResourceModel struct {
	workspaceModel

	Retry    *RetryModel    `tfsdk:"retry"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (resource *WorkspaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"region": tfschema.StringAttribute{
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
			"retry": retryResourceSchema(),
		},
	}
}

func (r *WorkspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.RegionalClient

	r.tenant = clients.Tenant
	r.region = clients.Region

	r.retry = retryConfig{
		delay:       clients.RetryDelay,
		interval:    clients.RetryInterval,
		maxAttempts: clients.RetryMaxAttempts,
	}

	tflog.Debug(ctx, "configured workspace resource")
}

func (r *WorkspaceResource) logFields(ctx context.Context, data WorkspaceResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (resource *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	ctx = resource.logFields(ctx, plan)
	tflog.Debug(ctx, "creating workspace")

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   plan.Name.ValueString(),
		},
		Labels:      toStringMap(plan.Labels),
		Annotations: toStringMap(plan.Annotations),
		Extensions:  toStringMap(plan.Extensions),
	}

	workspace, err := resource.client.WorkspaceV1.CreateOrUpdateWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace",
			"An error was encountered when creating the workspace.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for workspace to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := resource.retry.with(plan.Retry).withTimeout(createTimeout).untilState(sdk.ResourceStateActive)

	workspace, err = resource.client.WorkspaceV1.GetWorkspaceUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become active.\nError: "+err.Error(),
		)
		return
	}

	state, diags2 := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Timeouts = plan.Timeouts

	tflog.Info(ctx, "workspace created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (resource *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "reading workspace")

	// Read the workspace

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(resource.tenant),
		Name:   data.Name.ValueString(),
	}

	workspace, err := resource.client.WorkspaceV1.GetWorkspace(ctx, tref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "workspace not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered when reading the workspace.\nError: "+err.Error(),
		)
		return
	}

	data, diags := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	ctx = resource.logFields(ctx, plan)
	tflog.Debug(ctx, "updating workspace")

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   plan.Name.ValueString(),
		},
		Labels:      toStringMap(plan.Labels),
		Annotations: toStringMap(plan.Annotations),
		Extensions:  toStringMap(plan.Extensions),
	}

	workspace, err := resource.client.WorkspaceV1.CreateOrUpdateWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating workspace",
			"An error was encountered when updating the workspace.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for workspace to become active")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := resource.retry.with(plan.Retry).withTimeout(updateTimeout).untilState(sdk.ResourceStateActive)

	workspace, err = resource.client.WorkspaceV1.GetWorkspaceUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become active.\nError: "+err.Error(),
		)
		return
	}

	state, diags2 := workspaceToResourceModel(ctx, workspace)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Timeouts = plan.Timeouts

	tflog.Info(ctx, "workspace updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (resource *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "deleting workspace")

	workspace := &sdk.Workspace{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   data.Name.ValueString(),
		},
	}

	err := resource.client.WorkspaceV1.DeleteWorkspace(ctx, workspace)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting workspace",
			"An error was encountered when deleting the workspace.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for workspace to be deleted")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(workspace.Metadata.Tenant),
		Name:   workspace.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).withTimeout(deleteTimeout).observer()

	err = resource.client.WorkspaceV1.WatchWorkspaceUntilDeleted(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace",
			"An error was encountered while waiting for the workspace to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "workspace deleted")
}

func workspaceToResourceModel(ctx context.Context, workspace *sdk.Workspace) (WorkspaceResourceModel, diag.Diagnostics) {
	common, diags := workspaceToBaseModel(ctx, workspace)
	return WorkspaceResourceModel{workspaceModel: common}, diags
}
