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
	_ datasource.DataSource              = (*InstanceSkuDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InstanceSkuDataSource)(nil)
)

type InstanceSkuDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newInstanceSkuDataSource() datasource.DataSource {
	return &InstanceSkuDataSource{}
}

func (d *InstanceSkuDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_sku"
}

type InstanceSkuDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Vcpus types.Int64 `tfsdk:"vcpus"`
	Ram   types.Int64 `tfsdk:"ram"`
}

func (d *InstanceSkuDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"vcpus": tfschema.Int64Attribute{
				Computed: true,
			},
			"ram": tfschema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *InstanceSkuDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured instance sku data source")
}

func (d *InstanceSkuDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceSkuDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading instance sku data source")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	sku, err := d.client.ComputeV1.GetSku(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance SKU",
			"An error was encountered when reading the instance SKU.\nError: "+err.Error(),
		)
		return
	}

	data, diags := instanceSkuToDataSourceModel(ctx, sku)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func instanceSkuToDataSourceModel(ctx context.Context, sku *sdk.InstanceSku) (InstanceSkuDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := InstanceSkuDataSourceModel{}
	model.Id = types.StringValue(sku.Metadata.Ref)

	model.Name = types.StringValue(sku.Metadata.Name)
	model.Tenant = types.StringValue(sku.Metadata.Tenant)
	model.Region = types.StringValue(sku.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(sku.Metadata.Ref)

	labels, d := fromStringMap(ctx, sku.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, sku.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, sku.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	if sku.Spec != nil {
		model.Vcpus = types.Int64Value(int64(sku.Spec.VCPU))
		model.Ram = types.Int64Value(int64(sku.Spec.Ram))
	}

	return model, diags
}
