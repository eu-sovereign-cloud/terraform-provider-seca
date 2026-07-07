package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/stretchr/testify/assert"
)

func TestCIDRValidator_Valid(t *testing.T) {
	cases := []string{
		"10.0.0.0/8",
		"192.168.0.0/16",
		"172.16.0.0/12",
		"0.0.0.0/0",
	}

	v := CIDRValidator()
	for _, cidr := range cases {
		req := validator.StringRequest{
			Path:        path.Root("ipv4"),
			ConfigValue: types.StringValue(cidr),
		}
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), req, resp)
		assert.False(t, resp.Diagnostics.HasError(), "expected %q to be valid", cidr)
	}
}

func TestCIDRValidator_Invalid(t *testing.T) {
	cases := []string{
		"not-a-cidr",
		"10.0.0.0",
		"300.0.0.0/8",
		"",
		"10.0.0.0/33",
	}

	v := CIDRValidator()
	for _, cidr := range cases {
		req := validator.StringRequest{
			Path:        path.Root("ipv4"),
			ConfigValue: types.StringValue(cidr),
		}
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), req, resp)
		assert.True(t, resp.Diagnostics.HasError(), "expected %q to be invalid", cidr)
	}
}

func TestCIDRValidator_NullUnknown(t *testing.T) {
	v := CIDRValidator()

	reqNull := validator.StringRequest{
		Path:        path.Root("ipv4"),
		ConfigValue: types.StringNull(),
	}
	respNull := &validator.StringResponse{}
	v.ValidateString(context.Background(), reqNull, respNull)
	assert.False(t, respNull.Diagnostics.HasError())

	reqUnknown := validator.StringRequest{
		Path:        path.Root("ipv4"),
		ConfigValue: types.StringUnknown(),
	}
	respUnknown := &validator.StringResponse{}
	v.ValidateString(context.Background(), reqUnknown, respUnknown)
	assert.False(t, respUnknown.Diagnostics.HasError())
}

func TestStringEnumValidator_Valid(t *testing.T) {
	v := StringEnumValidator("ingress", "egress")

	for _, val := range []string{"ingress", "egress"} {
		req := validator.StringRequest{
			Path:        path.Root("direction"),
			ConfigValue: types.StringValue(val),
		}
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), req, resp)
		assert.False(t, resp.Diagnostics.HasError(), "expected %q to be valid", val)
	}
}

func TestStringEnumValidator_Invalid(t *testing.T) {
	v := StringEnumValidator("ingress", "egress")

	for _, val := range []string{"both", "in", "out", ""} {
		req := validator.StringRequest{
			Path:        path.Root("direction"),
			ConfigValue: types.StringValue(val),
		}
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), req, resp)
		assert.True(t, resp.Diagnostics.HasError(), "expected %q to be invalid", val)
	}
}

func TestStringEnumValidator_NullUnknown(t *testing.T) {
	v := StringEnumValidator("IPv4", "IPv6")

	reqNull := validator.StringRequest{
		Path:        path.Root("version"),
		ConfigValue: types.StringNull(),
	}
	respNull := &validator.StringResponse{}
	v.ValidateString(context.Background(), reqNull, respNull)
	assert.False(t, respNull.Diagnostics.HasError())
}

func TestPortRangeValidator_Valid(t *testing.T) {
	v := PortRangeValidator()
	for _, port := range []int64{1, 80, 443, 8080, 65535} {
		req := validator.Int64Request{
			Path:           path.Root("port"),
			ConfigValue:    types.Int64Value(port),
			PathExpression: path.MatchRoot("port"),
			Config:         validator.Int64Request{}.Config,
		}
		_ = attr.Value(types.Int64Value(port))
		resp := &validator.Int64Response{}
		v.ValidateInt64(context.Background(), req, resp)
		assert.False(t, resp.Diagnostics.HasError(), "expected port %d to be valid", port)
	}
}

func TestPortRangeValidator_Invalid(t *testing.T) {
	v := PortRangeValidator()
	for _, port := range []int64{0, -1, 65536, 100000} {
		req := validator.Int64Request{
			Path:        path.Root("port"),
			ConfigValue: types.Int64Value(port),
		}
		resp := &validator.Int64Response{}
		v.ValidateInt64(context.Background(), req, resp)
		assert.True(t, resp.Diagnostics.HasError(), "expected port %d to be invalid", port)
	}
}

func TestPortRangeValidator_NullUnknown(t *testing.T) {
	v := PortRangeValidator()

	reqNull := validator.Int64Request{
		Path:        path.Root("port"),
		ConfigValue: types.Int64Null(),
	}
	respNull := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), reqNull, respNull)
	assert.False(t, respNull.Diagnostics.HasError())
}
