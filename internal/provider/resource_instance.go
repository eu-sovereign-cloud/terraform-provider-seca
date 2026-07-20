package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ resource.Resource                = (*InstanceResource)(nil)
	_ resource.ResourceWithConfigure   = (*InstanceResource)(nil)
	_ resource.ResourceWithImportState = (*InstanceResource)(nil)
)

type InstanceResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newInstanceResource() resource.Resource {
	return &InstanceResource{}
}

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

type InstanceResourceModel struct {
	instanceModel

	Retry    *RetryModel    `tfsdk:"retry"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceID, name, ok := strings.Cut(req.ID, "/")
	if !ok || workspaceID == "" || name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier in the format \"workspace_id/name\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func (r *InstanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	volumeAttrs := map[string]tfschema.Attribute{
		"device_id": tfschema.StringAttribute{
			Required: true,
		},
	}

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
			"workspace_id": tfschema.StringAttribute{
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
			"sku_id": tfschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary_nic_id": tfschema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": tfschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_keys": tfschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"boot_volume": tfschema.SingleNestedAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: volumeAttrs,
			},
			"data_volumes": tfschema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: volumeAttrs,
				},
			},
			"additional_nic_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_id": tfschema.StringAttribute{
				Optional: true,
			},
			"user_data": tfschema.StringAttribute{
				Optional: true,
			},
			"anti_affinity_group": tfschema.StringAttribute{
				Optional: true,
			},
			"power_state": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"power_state_since": tfschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured instance resource")
}

func (r *InstanceResource) logFields(ctx context.Context, data InstanceResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "creating instance")

	inst, diags := instanceFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := r.client.ComputeV1.CreateOrUpdateInstance(ctx, inst)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating instance",
			"An error was encountered when creating the instance.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for instance to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(inst.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(inst.Metadata.Workspace),
		Name:      inst.Metadata.Name,
	}

	config := r.retry.with(data.Retry).withTimeout(createTimeout).untilState(sdk.ResourceStateActive)

	inst, err = r.client.ComputeV1.GetInstanceUntilState(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"An error was encountered while waiting for the instance to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := instanceToResourceModel(ctx, inst)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	result.Retry = data.Retry
	result.Timeouts = data.Timeouts

	tflog.Info(ctx, "instance created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "reading instance")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(r.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	inst, err := r.client.ComputeV1.GetInstance(ctx, wref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "instance not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"An error was encountered when reading the instance.\nError: "+err.Error(),
		)
		return
	}

	result, diags := instanceToResourceModel(ctx, inst)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	result.Retry = data.Retry
	result.Timeouts = data.Timeouts

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "updating instance")

	inst, diags := instanceFromModel(ctx, r.tenant, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := r.client.ComputeV1.CreateOrUpdateInstance(ctx, inst)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating instance",
			"An error was encountered when updating the instance.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for instance to become active")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(inst.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(inst.Metadata.Workspace),
		Name:      inst.Metadata.Name,
	}

	config := r.retry.with(data.Retry).withTimeout(updateTimeout).untilState(sdk.ResourceStateActive)

	inst, err = r.client.ComputeV1.GetInstanceUntilState(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"An error was encountered while waiting for the instance to become active.\nError: "+err.Error(),
		)
		return
	}

	result, diags := instanceToResourceModel(ctx, inst)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	result.Retry = data.Retry
	result.Timeouts = data.Timeouts

	tflog.Info(ctx, "instance updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel
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

	ctx = r.logFields(ctx, data)
	tflog.Debug(ctx, "deleting instance")

	inst := &sdk.Instance{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    r.tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
	}

	err := r.client.ComputeV1.DeleteInstance(ctx, inst)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting instance",
			"An error was encountered when deleting the instance.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for instance to be deleted")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(inst.Metadata.Tenant),
		Workspace: secapi.WorkspaceID(inst.Metadata.Workspace),
		Name:      inst.Metadata.Name,
	}

	config := r.retry.with(data.Retry).withTimeout(deleteTimeout).observer()

	err = r.client.ComputeV1.WatchInstanceUntilDeleted(ctx, wref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"An error was encountered while waiting for the instance to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "instance deleted")
}

func instanceFromModel(ctx context.Context, tenant string, data InstanceResourceModel) (*sdk.Instance, diag.Diagnostics) {
	var diags diag.Diagnostics

	inst := &sdk.Instance{
		Metadata: &sdk.RegionalWorkspaceResourceMetadata{
			Tenant:    tenant,
			Workspace: data.WorkspaceId.ValueString(),
			Name:      data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.InstanceSpec{
			SkuRef: sdk.Reference{
				Resource: data.SkuId.ValueString(),
			},
			BootVolume: sdk.VolumeReference{
				DeviceRef: sdk.Reference{
					Resource: data.BootVolume.DeviceId.ValueString(),
				},
			},
		},
	}

	if !data.PrimaryNicId.IsNull() && !data.PrimaryNicId.IsUnknown() {
		ref := sdk.Reference{Resource: data.PrimaryNicId.ValueString()}
		inst.Spec.PrimaryNicRef = &ref
	}

	if !data.Zone.IsNull() && !data.Zone.IsUnknown() {
		inst.Spec.Zone = data.Zone.ValueString()
	}

	if !data.SecurityGroupId.IsNull() && !data.SecurityGroupId.IsUnknown() {
		ref := sdk.Reference{Resource: data.SecurityGroupId.ValueString()}
		inst.Spec.SecurityGroupRef = &ref
	}

	if !data.UserData.IsNull() && !data.UserData.IsUnknown() {
		inst.Spec.UserData = data.UserData.ValueString()
	}

	if !data.AntiAffinityGroup.IsNull() && !data.AntiAffinityGroup.IsUnknown() {
		inst.Spec.AntiAffinityGroup = data.AntiAffinityGroup.ValueString()
	}

	if !data.SshKeys.IsNull() && !data.SshKeys.IsUnknown() {
		var sshKeys []string
		d := data.SshKeys.ElementsAs(ctx, &sshKeys, false)
		diags.Append(d...)
		inst.Spec.SshKeys = sshKeys
	}

	if !data.DataVolumes.IsNull() && !data.DataVolumes.IsUnknown() {
		var vols []instanceVolumeModel
		d := data.DataVolumes.ElementsAs(ctx, &vols, false)
		diags.Append(d...)
		for _, v := range vols {
			inst.Spec.DataVolumes = append(inst.Spec.DataVolumes, sdk.VolumeReference{
				DeviceRef: sdk.Reference{Resource: v.DeviceId.ValueString()},
			})
		}
	}

	if !data.AdditionalNicIds.IsNull() && !data.AdditionalNicIds.IsUnknown() {
		var nicIds []string
		d := data.AdditionalNicIds.ElementsAs(ctx, &nicIds, false)
		diags.Append(d...)
		for _, id := range nicIds {
			inst.Spec.AdditionalNicRefs = append(inst.Spec.AdditionalNicRefs, sdk.Reference{Resource: id})
		}
	}

	return inst, diags
}

func instanceToResourceModel(ctx context.Context, inst *sdk.Instance) (InstanceResourceModel, diag.Diagnostics) {
	common, diags := instanceToBaseModel(ctx, inst)
	return InstanceResourceModel{instanceModel: common}, diags
}
