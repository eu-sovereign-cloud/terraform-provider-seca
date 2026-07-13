package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type cidrValidator struct{}

func (v cidrValidator) Description(_ context.Context) string {
	return "value must be a valid CIDR block (e.g. 10.0.0.0/16)"
}

func (v cidrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	_, _, err := net.ParseCIDR(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR block",
			fmt.Sprintf("Expected a valid CIDR block (e.g. \"10.0.0.0/16\"), got: %q", req.ConfigValue.ValueString()),
		)
	}
}

func CIDRValidator() validator.String { return cidrValidator{} }

type cidrV4Validator struct{}

func (v cidrV4Validator) Description(_ context.Context) string {
	return "value must be a valid IPv4 CIDR block (e.g. 10.0.0.0/16)"
}

func (v cidrV4Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cidrV4Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	ip, _, err := net.ParseCIDR(req.ConfigValue.ValueString())
	if err != nil || ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 CIDR block",
			fmt.Sprintf("Expected a valid IPv4 CIDR block (e.g. \"10.0.0.0/16\"), got: %q", req.ConfigValue.ValueString()),
		)
	}
}

func CIDRv4Validator() validator.String { return cidrV4Validator{} }

type cidrV6Validator struct{}

func (v cidrV6Validator) Description(_ context.Context) string {
	return "value must be a valid IPv6 CIDR block (e.g. 2001:db8::/32)"
}

func (v cidrV6Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cidrV6Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	ip, _, err := net.ParseCIDR(req.ConfigValue.ValueString())
	if err != nil || ip.To4() != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv6 CIDR block",
			fmt.Sprintf("Expected a valid IPv6 CIDR block (e.g. \"2001:db8::/32\"), got: %q", req.ConfigValue.ValueString()),
		)
	}
}

func CIDRv6Validator() validator.String { return cidrV6Validator{} }

type stringEnumValidator struct {
	allowed []string
}

func (v stringEnumValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s", strings.Join(v.allowed, ", "))
}

func (v stringEnumValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringEnumValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, a := range v.allowed {
		if val == a {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid enum value",
		fmt.Sprintf("Expected one of [%s], got: %q", strings.Join(v.allowed, ", "), val),
	)
}

func StringEnumValidator(allowed ...string) validator.String {
	return stringEnumValidator{allowed: allowed}
}

type portRangeValidator struct{}

func (v portRangeValidator) Description(_ context.Context) string {
	return "value must be a valid port number between 1 and 65535"
}

func (v portRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v portRangeValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val < 1 || val > 65535 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid port number",
			fmt.Sprintf("Expected a port number between 1 and 65535, got: %d", val),
		)
	}
}

func PortRangeValidator() validator.Int64 { return portRangeValidator{} }
