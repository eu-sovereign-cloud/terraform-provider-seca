package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
)

type blockStorageModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	SizeGB        types.Int64  `tfsdk:"size_gb"`
	SkuId         types.String `tfsdk:"sku_id"`
	SourceImageId types.String `tfsdk:"source_image_id"`
}

func blockStorageToBaseModel(ctx context.Context, block *sdk.BlockStorage) (blockStorageModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := blockStorageModel{}
	model.Id = types.StringValue(block.Metadata.Ref)

	model.Name = types.StringValue(block.Metadata.Name)
	model.WorkspaceId = types.StringValue(block.Metadata.Workspace)
	model.Tenant = types.StringValue(block.Metadata.Tenant)
	model.Region = types.StringValue(block.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(block.Metadata.Ref)
	model.CreatedAt = fromTime(block.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(block.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(block.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, block.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, block.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, block.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.SizeGB = types.Int64Value(int64(block.Spec.SizeGB))
	model.SkuId = types.StringValue(block.Spec.SkuRef.Resource)
	model.SourceImageId = fromRefPtr(block.Spec.SourceImageRef)

	return model, diags
}

type imageModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	BlockStorageId  types.String `tfsdk:"block_storage_id"`
	CpuArchitecture types.String `tfsdk:"cpu_architecture"`
	Initializer     types.String `tfsdk:"initializer"`
	Boot            types.String `tfsdk:"boot"`
}

func imageToBaseModel(ctx context.Context, image *sdk.Image) (imageModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := imageModel{}
	model.Id = types.StringValue(image.Metadata.Ref)

	model.Name = types.StringValue(image.Metadata.Name)
	model.Tenant = types.StringValue(image.Metadata.Tenant)
	model.Region = types.StringValue(image.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(image.Metadata.Ref)
	model.CreatedAt = fromTime(image.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(image.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(image.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, image.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, image.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, image.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.BlockStorageId = types.StringValue(image.Spec.BlockStorageRef.Resource)
	model.CpuArchitecture = types.StringValue(string(image.Spec.CpuArchitecture))
	model.Initializer = types.StringValue(string(image.Spec.Initializer))
	model.Boot = types.StringValue(string(image.Spec.Boot))

	return model, diags
}

type workspaceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`
}

func workspaceToBaseModel(ctx context.Context, workspace *sdk.Workspace) (workspaceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := workspaceModel{}
	model.Id = types.StringValue(workspace.Metadata.Ref)

	model.Name = types.StringValue(workspace.Metadata.Name)
	model.Tenant = types.StringValue(workspace.Metadata.Tenant)
	model.Region = types.StringValue(workspace.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(workspace.Metadata.Ref)
	model.CreatedAt = fromTime(workspace.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(workspace.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(workspace.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, workspace.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, workspace.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, workspace.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	return model, diags
}

type networkModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	SkuId           types.String     `tfsdk:"sku_id"`
	Cidr            NetworkCidrModel `tfsdk:"cidr"`
	AdditionalCidrs types.List       `tfsdk:"additional_cidrs"`
}

func networkToBaseModel(ctx context.Context, net *sdk.Network) (networkModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := networkModel{}
	model.Id = types.StringValue(net.Metadata.Ref)
	model.Name = types.StringValue(net.Metadata.Name)
	model.WorkspaceId = types.StringValue(net.Metadata.Workspace)
	model.Tenant = types.StringValue(net.Metadata.Tenant)
	model.Region = types.StringValue(net.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(net.Metadata.Ref)
	model.CreatedAt = fromTime(net.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(net.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(net.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, net.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, net.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, net.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.SkuId = types.StringValue(net.Spec.SkuRef.Resource)

	if net.Status != nil {
		model.Cidr = cidrFromSDK(net.Status.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Status.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	} else {
		model.Cidr = cidrFromSDK(net.Spec.Cidr)
		additionalCidrs, d := fromCidrList(ctx, net.Spec.AdditionalCidrs)
		diags.Append(d...)
		model.AdditionalCidrs = additionalCidrs
	}

	return model, diags
}

type internetGatewayModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	Tenant           types.String `tfsdk:"tenant"`
	Region           types.String `tfsdk:"region"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	EgressOnly types.Bool `tfsdk:"egress_only"`
}

func internetGatewayToBaseModel(ctx context.Context, gtw *sdk.InternetGateway) (internetGatewayModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := internetGatewayModel{}
	model.Id = types.StringValue(gtw.Metadata.Ref)
	model.Name = types.StringValue(gtw.Metadata.Name)
	model.WorkspaceId = types.StringValue(gtw.Metadata.Workspace)
	model.Tenant = types.StringValue(gtw.Metadata.Tenant)
	model.Region = types.StringValue(gtw.Metadata.Region)
	model.ResourceProvider = refToResourceProvider(gtw.Metadata.Ref)
	model.CreatedAt = fromTime(gtw.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(gtw.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(gtw.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, gtw.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, gtw.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, gtw.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	model.EgressOnly = types.BoolValue(gtw.Spec.EgressOnly)

	return model, diags
}

var permissionAttrTypes = map[string]attr.Type{
	"provider":  types.StringType,
	"resources": types.ListType{ElemType: types.StringType},
	"verb":      types.ListType{ElemType: types.StringType},
}

type permissionModel struct {
	Provider  types.String `tfsdk:"provider"`
	Resources types.List   `tfsdk:"resources"`
	Verb      types.List   `tfsdk:"verb"`
}

type roleModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Tenant           types.String `tfsdk:"tenant"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Permissions types.List `tfsdk:"permissions"`
}

func roleToBaseModel(ctx context.Context, role *sdk.Role) (roleModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := roleModel{}
	model.Id = types.StringValue(role.Metadata.Ref)
	model.Name = types.StringValue(role.Metadata.Name)
	model.Tenant = types.StringValue(role.Metadata.Tenant)
	model.ResourceProvider = refToResourceProvider(role.Metadata.Ref)
	model.CreatedAt = fromTime(role.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(role.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(role.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, role.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, role.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, role.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	perms := make([]permissionModel, 0, len(role.Spec.Permissions))
	for _, p := range role.Spec.Permissions {
		resources, d := types.ListValueFrom(ctx, types.StringType, p.Resources)
		diags.Append(d...)
		verb, d := types.ListValueFrom(ctx, types.StringType, p.Verb)
		diags.Append(d...)
		perms = append(perms, permissionModel{
			Provider:  types.StringValue(p.Provider),
			Resources: resources,
			Verb:      verb,
		})
	}

	permissions, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: permissionAttrTypes}, perms)
	diags.Append(d...)
	model.Permissions = permissions

	return model, diags
}

var scopeAttrTypes = map[string]attr.Type{
	"tenants":    types.ListType{ElemType: types.StringType},
	"regions":    types.ListType{ElemType: types.StringType},
	"workspaces": types.ListType{ElemType: types.StringType},
}

type scopeModel struct {
	Tenants    types.List `tfsdk:"tenants"`
	Regions    types.List `tfsdk:"regions"`
	Workspaces types.List `tfsdk:"workspaces"`
}

type roleAssignmentModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Tenant           types.String `tfsdk:"tenant"`
	ResourceProvider types.String `tfsdk:"resource_provider"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DeletedAt        types.String `tfsdk:"deleted_at"`
	LastModifiedAt   types.String `tfsdk:"last_modified_at"`

	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
	Extensions  types.Map `tfsdk:"extensions"`

	Subs   types.List `tfsdk:"subs"`
	Scopes types.List `tfsdk:"scopes"`
	Roles  types.List `tfsdk:"roles"`
}

func roleAssignmentToBaseModel(ctx context.Context, ra *sdk.RoleAssignment) (roleAssignmentModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := roleAssignmentModel{}
	model.Id = types.StringValue(ra.Metadata.Ref)
	model.Name = types.StringValue(ra.Metadata.Name)
	model.Tenant = types.StringValue(ra.Metadata.Tenant)
	model.ResourceProvider = refToResourceProvider(ra.Metadata.Ref)
	model.CreatedAt = fromTime(ra.Metadata.CreatedAt)
	model.DeletedAt = fromTimePtr(ra.Metadata.DeletedAt)
	model.LastModifiedAt = fromTime(ra.Metadata.LastModifiedAt)

	labels, d := fromStringMap(ctx, ra.Labels)
	diags.Append(d...)
	model.Labels = labels

	annotations, d := fromStringMap(ctx, ra.Annotations)
	diags.Append(d...)
	model.Annotations = annotations

	extensions, d := fromStringMap(ctx, ra.Extensions)
	diags.Append(d...)
	model.Extensions = extensions

	subs, d := types.ListValueFrom(ctx, types.StringType, ra.Spec.Subs)
	diags.Append(d...)
	model.Subs = subs

	scopes := make([]scopeModel, 0, len(ra.Spec.Scopes))
	for _, s := range ra.Spec.Scopes {
		tenants, d := types.ListValueFrom(ctx, types.StringType, s.Tenants)
		diags.Append(d...)
		regions, d := types.ListValueFrom(ctx, types.StringType, s.Regions)
		diags.Append(d...)
		workspaces, d := types.ListValueFrom(ctx, types.StringType, s.Workspaces)
		diags.Append(d...)
		scopes = append(scopes, scopeModel{
			Tenants:    tenants,
			Regions:    regions,
			Workspaces: workspaces,
		})
	}

	scopesList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: scopeAttrTypes}, scopes)
	diags.Append(d...)
	model.Scopes = scopesList

	roles, d := types.ListValueFrom(ctx, types.StringType, ra.Spec.Roles)
	diags.Append(d...)
	model.Roles = roles

	return model, diags
}
