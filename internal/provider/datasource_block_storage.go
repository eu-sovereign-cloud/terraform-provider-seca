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
	_ datasource.DataSource              = (*BlockStorageDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BlockStorageDataSource)(nil)
)

type BlockStorageDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newBlockStorageDataSource() datasource.DataSource {
	return &BlockStorageDataSource{}
}

func (d *BlockStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage"
}

type BlockStorageDataSourceModel struct {
	blockStorageModel

	State types.String `tfsdk:"state"`
}

func (d *BlockStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				Computed: true,
			},
			"name": tfschema.StringAttribute{
				Required: true,
			},
			"workspace_id": tfschema.StringAttribute{
				Required: true,
			},
			"tenant": tfschema.StringAttribute{
				Computed: true,
			},
			"region": tfschema.StringAttribute{
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
			"size_gb": tfschema.Int64Attribute{
				Computed: true,
			},
			"sku_id": tfschema.StringAttribute{
				Computed: true,
			},
			"source_image_id": tfschema.StringAttribute{
				Computed: true,
			},
			"state": tfschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *BlockStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured block storage data source")
}

func (d *BlockStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlockStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading block storage data source")

	// Read the block storage

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	block, err := d.client.StorageV1.GetBlockStorage(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage",
			"An error was encountered when reading the block storage.\nError: "+err.Error(),
		)
		return
	}

	data, diags := blockStorageToDataSourceModel(ctx, block)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func blockStorageToDataSourceModel(ctx context.Context, block *sdk.BlockStorage) (BlockStorageDataSourceModel, diag.Diagnostics) {
	common, diags := blockStorageToBaseModel(ctx, block)
	model := BlockStorageDataSourceModel{blockStorageModel: common}
	model.State = types.StringValue(string(block.Status.State))
	return model, diags
}
