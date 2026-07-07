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
	_ datasource.DataSource              = (*RoleDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*RoleDataSource)(nil)
)

type RoleDataSource struct {
	client *secapi.GlobalClient
	tenant string
}

func newRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

func (d *RoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

type RoleDataSourceModel struct {
	roleModel
	State types.String `tfsdk:"state"`
}

func (d *RoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"permissions": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"provider": tfschema.StringAttribute{
							Computed: true,
						},
						"resources": tfschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"verb": tfschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
			"state": tfschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *RoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.GlobalClient
	d.tenant = clients.Tenant

	tflog.Debug(ctx, "configured role data source")
}

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading role data source")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	role, err := d.client.AuthorizationV1.GetRole(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"An error was encountered when reading the role.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleToDataSourceModel(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func roleToDataSourceModel(ctx context.Context, role *sdk.Role) (RoleDataSourceModel, diag.Diagnostics) {
	base, diags := roleToBaseModel(ctx, role)

	model := RoleDataSourceModel{roleModel: base}

	if role.Status != nil {
		model.State = types.StringValue(string(role.Status.State))
	} else {
		model.State = types.StringNull()
	}

	return model, diags
}
