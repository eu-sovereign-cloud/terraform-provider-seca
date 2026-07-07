package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// cidrValidator validates that a string is a valid CIDR block.
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

// CIDRValidator returns a validator that checks the value is a valid CIDR block.
func CIDRValidator() validator.String { return cidrValidator{} }

// stringEnumValidator validates that a string is one of the allowed values.
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

// StringEnumValidator returns a validator that checks the value is one of the allowed strings.
func StringEnumValidator(allowed ...string) validator.String {
	return stringEnumValidator{allowed: allowed}
}

// portRangeValidator validates that an int64 is a valid port number (1-65535).
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

// PortRangeValidator returns a validator that checks the value is a valid port (1-65535).
func PortRangeValidator() validator.Int64 { return portRangeValidator{} }
