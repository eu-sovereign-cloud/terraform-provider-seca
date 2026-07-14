package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type SecaProvider struct {
	Version string
}

type SecaProviderModel struct {
	Token  types.String `tfsdk:"token"`
	Tenant types.String `tfsdk:"tenant"`
	Region types.String `tfsdk:"region"`

	Retry *RetryModel `tfsdk:"retry"`

	GlobalProviders *SecaGlobalProvidersModel `tfsdk:"global_providers"`
}

type SecaGlobalProvidersModel struct {
	RegionV1        types.String `tfsdk:"region_v1"`
	AuthorizationV1 types.String `tfsdk:"authorization_v1"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SecaProvider{
			Version: version,
		}
	}
}

func (p *SecaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "seca"
	resp.Version = p.Version
}

func (p *SecaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The **seca** provider manages resources on the SECA (Sovereign European Cloud API) platform.\n\n" +
			"Configure the provider with a bearer token, tenant ID, and region. " +
			"Use `global_providers` to point at the SECA control-plane endpoints for your region. " +
			"The optional `retry` block fine-tunes polling behaviour for async operations.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Bearer token used to authenticate API requests. Set via `SECA_TOKEN` environment variable.",
			},
			"tenant": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Tenant (organisation) identifier. All resources are created under this tenant.",
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Region identifier (e.g. `eu-central-1`). Determines which regional endpoints are used.",
			},
			"retry": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Fine-grained polling configuration for async operations. All fields are optional; omitted fields inherit provider defaults.",
				Attributes: map[string]schema.Attribute{
					"delay": schema.NumberAttribute{
						Optional:            true,
						MarkdownDescription: "Initial wait in seconds before the first status poll. Default: 30.",
					},
					"interval": schema.NumberAttribute{
						Optional:            true,
						MarkdownDescription: "Wait in seconds between subsequent status polls. Default: 10.",
					},
					"max_attempts": schema.NumberAttribute{
						Optional:            true,
						MarkdownDescription: "Maximum number of polling attempts before failing. Default: 5. Superseded by per-resource `timeouts` blocks.",
					},
				},
			},
			"global_providers": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Control-plane endpoint URLs. Obtain these from your SECA account portal.",
				Attributes: map[string]schema.Attribute{
					"region_v1": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "URL of the RegionV1 provider endpoint (e.g. `https://api.seca.cloud/providers/seca.region`).",
					},
					"authorization_v1": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "URL of the AuthorizationV1 provider endpoint. Required when managing IAM roles and role assignments.",
					},
				},
			},
		},
	}
}

func (p *SecaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model SecaProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "configuring seca provider")

	config := &clientConfig{
		Token:  model.Token.ValueString(),
		Tenant: model.Tenant.ValueString(),
		Region: model.Region.ValueString(),

		RetryDelay:       defaultRetryDelay,
		RetryInterval:    defaultRetryInterval,
		RetryMaxAttempts: defaultRetryMaxAttempts,
	}

	if model.Retry != nil {
		if !model.Retry.Delay.IsNull() {
			config.RetryDelay = numberToDuration(model.Retry.Delay)
		}
		if !model.Retry.Interval.IsNull() {
			config.RetryInterval = numberToDuration(model.Retry.Interval)
		}
		if !model.Retry.MaxAttempts.IsNull() {
			config.RetryMaxAttempts = numberToInt(model.Retry.MaxAttempts)
		}
	}

	config.GlobalProviders = &clientConfigGlobalProviders{
		RegionV1:        model.GlobalProviders.RegionV1.ValueString(),
		AuthorizationV1: model.GlobalProviders.AuthorizationV1.ValueString(),
	}

	clients, err := initClients(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to initialize SDK client",
			"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = *clients
	resp.ResourceData = *clients

	tflog.Info(ctx, "configured seca provider")
}

func (p *SecaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newWorkspaceResource,
		newImageResource,
		newBlockStorageResource,
		newNetworkResource,
		newInternetGatewayResource,
		newRoleResource,
		newRoleAssignmentResource,
		newRouteTableResource,
		newSubnetResource,
		newSecurityGroupResource,
		newPublicIpResource,
		newNicResource,
		newInstanceResource,
	}
}

func (p *SecaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newWorkspaceDataSource,
		newRegionDataSource,
		newImageDataSource,
		newBlockStorageDataSource,
		newStorageSkuDataSource,
		newNetworkSkuDataSource,
		newNetworkDataSource,
		newInternetGatewayDataSource,
		newRoleDataSource,
		newRouteTableDataSource,
		newSubnetDataSource,
		newSecurityGroupDataSource,
		newPublicIpDataSource,
		newNicDataSource,
		newInstanceSkuDataSource,
		newInstanceDataSource,
	}
}
