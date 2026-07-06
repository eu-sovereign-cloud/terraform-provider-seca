package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

var (
	_ datasource.DataSource              = (*ImageDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ImageDataSource)(nil)
)

type ImageDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newImageDataSource() datasource.DataSource {
	return &ImageDataSource{}
}

func (d *ImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

type ImageDataSourceModel struct {
	imageModel

	State types.String `tfsdk:"state"`
}

func (d *ImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
			},
			"name": tfschema.StringAttribute{
				Required: true,
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
			},
			"region": tfschema.StringAttribute{
				Computed: true,
			},
			"resource_provider": tfschema.StringAttribute{
				Computed: true,
			},
			"created_at": tfschema.StringAttribute{
				Computed: true,
			},
			"deleted_at": tfschema.StringAttribute{
				Computed: true,
			},
			"last_modified_at": tfschema.StringAttribute{
				Computed: true,
			},
			"labels": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"annotations": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"extensions": tfschema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"block_storage_id": tfschema.StringAttribute{
				Computed: true,
			},
			"cpu_architecture": tfschema.StringAttribute{
				Computed: true,
			},
			"initializer": tfschema.StringAttribute{
				Computed: true,
			},
			"boot": tfschema.StringAttribute{
				Computed: true,
			},
			"state": tfschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.RegionalClient
	d.tenant = clients.Tenant

	tflog.Debug(ctx, "configured image data source")
}

func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading image data source")

	// Read the image

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	image, err := d.client.StorageV1.GetImage(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading image",
			"An error was encountered when reading the image.\nError: "+err.Error(),
		)
		return
	}

	data, diags := imageToDataSourceModel(ctx, image)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func imageToDataSourceModel(ctx context.Context, image *sdk.Image) (ImageDataSourceModel, diag.Diagnostics) {
	common, diags := imageToBaseModel(ctx, image)
	model := ImageDataSourceModel{imageModel: common}
	model.State = types.StringValue(string(image.Status.State))
	return model, diags
}
