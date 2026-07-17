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
	_ datasource.DataSource              = (*RoleAssignmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*RoleAssignmentDataSource)(nil)
)

type RoleAssignmentDataSource struct {
	client *secapi.GlobalClient
	tenant string
}

func newRoleAssignmentDataSource() datasource.DataSource {
	return &RoleAssignmentDataSource{}
}

func (d *RoleAssignmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_assignment"
}

type RoleAssignmentDataSourceModel struct {
	roleAssignmentModel
	State types.String `tfsdk:"state"`
}

func (d *RoleAssignmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"subs": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"scopes": tfschema.ListNestedAttribute{
				Computed: true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"tenants": tfschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"regions": tfschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"workspaces": tfschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
			"roles": tfschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"state": tfschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *RoleAssignmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	tflog.Debug(ctx, "configured role assignment data source")
}

func (d *RoleAssignmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleAssignmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tenant_id", d.tenant)
	ctx = tflog.SetField(ctx, "name", data.Name.ValueString())
	tflog.Debug(ctx, "reading role assignment data source")

	tref := secapi.TenantReference{
		Tenant: secapi.TenantID(d.tenant),
		Name:   data.Name.ValueString(),
	}

	ra, err := d.client.AuthorizationV1.GetRoleAssignment(ctx, tref)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role assignment",
			"An error was encountered when reading the role assignment.\nError: "+err.Error(),
		)
		return
	}

	result, diags := roleAssignmentToDataSourceModel(ctx, ra)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func roleAssignmentToDataSourceModel(ctx context.Context, ra *sdk.RoleAssignment) (RoleAssignmentDataSourceModel, diag.Diagnostics) {
	base, diags := roleAssignmentToBaseModel(ctx, ra)

	model := RoleAssignmentDataSourceModel{roleAssignmentModel: base}

	if ra.Status != nil {
		model.State = types.StringValue(string(ra.Status.State))
	} else {
		model.State = types.StringNull()
	}

	return model, diags
}
