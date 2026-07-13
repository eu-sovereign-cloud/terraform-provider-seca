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
	_ resource.Resource                = (*ImageResource)(nil)
	_ resource.ResourceWithConfigure   = (*ImageResource)(nil)
	_ resource.ResourceWithImportState = (*ImageResource)(nil)
)

type ImageResource struct {
	client *secapi.RegionalClient

	tenant string
	region string

	retry retryConfig
}

func newImageResource() resource.Resource {
	return &ImageResource{}
}

func (resource *ImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *ImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

type ImageResourceModel struct {
	imageModel

	Retry    *RetryModel    `tfsdk:"retry"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (resource *ImageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"block_storage_id": tfschema.StringAttribute{
				Required: true,
			},
			"cpu_architecture": tfschema.StringAttribute{
				Required: true,
			},
			"initializer": tfschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"boot": tfschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"retry": retryResourceSchema(),
		},
	}
}

func (r *ImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured image resource")
}

func (r *ImageResource) logFields(ctx context.Context, data ImageResourceModel) context.Context {
	ctx = tflog.SetField(ctx, "tenant_id", r.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	return ctx
}

func (resource *ImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "creating image")

	// Create the image

	image := imageFromModel(resource.tenant, data)

	image, err := resource.client.StorageV1.CreateOrUpdateImage(ctx, image)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating image",
			"An error was encountered when creating the image.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for image to become active")

	// Wait until it is active

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(image.Metadata.Tenant),
		Name:   image.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).withTimeout(createTimeout).untilState(sdk.ResourceStateActive)

	image, err = resource.client.StorageV1.GetImageUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered while waiting for the image to become active.\nError: "+err.Error(),
		)
		return
	}

	savedTimeouts := data.Timeouts
	data, diags2 := imageToResourceModel(ctx, image)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Timeouts = savedTimeouts

	tflog.Info(ctx, "image created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "reading image")

	// Read the image

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(resource.tenant),
		Name:   data.Name.ValueString(),
	}

	image, err := resource.client.StorageV1.GetImage(ctx, tref)
	if err == secapi.ErrResourceNotFound {
		tflog.Debug(ctx, "image not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered when reading the image.\nError: "+err.Error(),
		)
		return
	}

	data, diags := imageToResourceModel(ctx, image)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "updating image")

	// Update the image

	image := imageFromModel(resource.tenant, data)

	image, err := resource.client.StorageV1.CreateOrUpdateImage(ctx, image)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating image",
			"An error was encountered when updating the image.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for image to become active")

	// Wait until it is active

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(image.Metadata.Tenant),
		Name:   image.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).withTimeout(updateTimeout).untilState(sdk.ResourceStateActive)

	image, err = resource.client.StorageV1.GetImageUntilState(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered while waiting for the image to become active.\nError: "+err.Error(),
		)
		return
	}

	savedTimeouts := data.Timeouts
	data, diags2 := imageToResourceModel(ctx, image)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Timeouts = savedTimeouts

	tflog.Info(ctx, "image updated")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (resource *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = resource.logFields(ctx, data)
	tflog.Debug(ctx, "deleting image")

	// Delete the image

	image := &sdk.Image{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: resource.tenant,
			Name:   data.Name.ValueString(),
		},
	}

	err := resource.client.StorageV1.DeleteImage(ctx, image)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting image",
			"An error was encountered when deleting the image.\nError: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "waiting for image to be deleted")

	// Wait until it is deleted

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(image.Metadata.Tenant),
		Name:   image.Metadata.Name,
	}

	config := resource.retry.with(data.Retry).withTimeout(deleteTimeout).observer()

	err = resource.client.StorageV1.WatchImageUntilDeleted(ctx, tref, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered while waiting for the image to become deleted.\nError: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "image deleted")
}

func imageFromModel(tenant string, data ImageResourceModel) *sdk.Image {
	image := &sdk.Image{
		Metadata: &sdk.RegionalResourceMetadata{
			Tenant: tenant,
			Name:   data.Name.ValueString(),
		},
		Labels:      toStringMap(data.Labels),
		Annotations: toStringMap(data.Annotations),
		Extensions:  toStringMap(data.Extensions),
		Spec: sdk.ImageSpec{
			BlockStorageRef: sdk.Reference{
				Resource: data.BlockStorageId.ValueString(),
			},
			CpuArchitecture: sdk.ImageSpecCpuArchitecture(data.CpuArchitecture.ValueString()),
		},
	}

	if !data.Boot.IsNull() && !data.Boot.IsUnknown() {
		image.Spec.Boot = sdk.ImageSpecBoot(data.Boot.ValueString())
	}
	if !data.Initializer.IsNull() && !data.Initializer.IsUnknown() {
		image.Spec.Initializer = sdk.ImageSpecInitializer(data.Initializer.ValueString())
	}

	return image
}

func imageToResourceModel(ctx context.Context, image *sdk.Image) (ImageResourceModel, diag.Diagnostics) {
	common, diags := imageToBaseModel(ctx, image)
	return ImageResourceModel{imageModel: common}, diags
}
