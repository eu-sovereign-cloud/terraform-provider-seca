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
	_ datasource.DataSource              = (*InstanceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InstanceDataSource)(nil)
)

type InstanceDataSource struct {
	client *secapi.RegionalClient
	tenant string
}

func newInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

func (d *InstanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

type InstanceDataSourceModel struct {
	instanceModel

	State types.String `tfsdk:"state"`
}

func (d *InstanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	volumeAttrs := map[string]tfschema.Attribute{
		"device_id": tfschema.StringAttribute{Computed: true},
	}

	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"id":                tfschema.StringAttribute{Computed: true},
			"name":              tfschema.StringAttribute{Required: true},
			"workspace_id":      tfschema.StringAttribute{Required: true},
			"tenant":            tfschema.StringAttribute{Computed: true},
			"region":            tfschema.StringAttribute{Computed: true},
			"resource_provider": tfschema.StringAttribute{Computed: true},
			"created_at":        tfschema.StringAttribute{Computed: true},
			"deleted_at":        tfschema.StringAttribute{Computed: true},
			"last_modified_at":  tfschema.StringAttribute{Computed: true},
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
			"sku_id":              tfschema.StringAttribute{Computed: true},
			"primary_nic_id":      tfschema.StringAttribute{Computed: true},
			"zone":                tfschema.StringAttribute{Computed: true},
			"security_group_id":   tfschema.StringAttribute{Computed: true},
			"user_data":           tfschema.StringAttribute{Computed: true},
			"anti_affinity_group": tfschema.StringAttribute{Computed: true},
			"power_state":         tfschema.StringAttribute{Computed: true},
			"power_state_since":   tfschema.StringAttribute{Computed: true},
			"ssh_keys": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"boot_volume": tfschema.SingleNestedAttribute{
				Computed:   true,
				Attributes: volumeAttrs,
			},
			"data_volumes": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: volumeAttrs,
				},
			},
			"additional_nic_ids": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"state": tfschema.StringAttribute{Computed: true},
		},
	}
}

func (d *InstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured instance data source")
}

func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "workspace_id", data.WorkspaceId.ValueString())
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading instance data source")

	wref := secapi.WorkspaceReference{
		Tenant:    secapi.TenantID(d.tenant),
		Workspace: secapi.WorkspaceID(data.WorkspaceId.ValueString()),
		Name:      data.Name.ValueString(),
	}

	inst, err := d.client.ComputeV1.GetInstance(ctx, wref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading instance",
			"An error was encountered when reading the instance.\nError: "+err.Error(),
		)
		return
	}

	data, diags := instanceToDataSourceModel(ctx, inst)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func instanceToDataSourceModel(ctx context.Context, inst *sdk.Instance) (InstanceDataSourceModel, diag.Diagnostics) {
	common, diags := instanceToBaseModel(ctx, inst)
	model := InstanceDataSourceModel{instanceModel: common}
	if inst.Status != nil {
		model.State = types.StringValue(string(inst.Status.State))
	} else {
		model.State = types.StringNull()
	}
	return model, diags
}
