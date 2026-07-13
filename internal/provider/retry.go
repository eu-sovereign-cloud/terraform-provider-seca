package provider

import (
	"math"
	"time"

	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdk "github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

type RetryModel struct {
	Delay       types.Number `tfsdk:"delay"`
	Interval    types.Number `tfsdk:"interval"`
	MaxAttempts types.Number `tfsdk:"max_attempts"`
}

type retryConfig struct {
	delay       time.Duration
	interval    time.Duration
	maxAttempts int
}

func retryResourceSchema() tfschema.SingleNestedAttribute {
	return tfschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]tfschema.Attribute{
			"delay": tfschema.NumberAttribute{
				Optional: true,
			},
			"interval": tfschema.NumberAttribute{
				Optional: true,
			},
			"max_attempts": tfschema.NumberAttribute{
				Optional: true,
			},
		},
	}
}

func (c retryConfig) with(override *RetryModel) retryConfig {
	if override == nil {
		return c
	}
	if !override.Delay.IsNull() && !override.Delay.IsUnknown() {
		c.delay = numberToDuration(override.Delay)
	}
	if !override.Interval.IsNull() && !override.Interval.IsUnknown() {
		c.interval = numberToDuration(override.Interval)
	}
	if !override.MaxAttempts.IsNull() && !override.MaxAttempts.IsUnknown() {
		c.maxAttempts = numberToInt(override.MaxAttempts)
	}
	return c
}

// withTimeout derives MaxAttempts from the given duration divided by the polling interval,
// overriding any max_attempts set via the retry block. timeout=0 leaves MaxAttempts unchanged.
func (c retryConfig) withTimeout(timeout time.Duration) retryConfig {
	if timeout > 0 && c.interval > 0 {
		c.maxAttempts = int(math.Ceil(float64(timeout) / float64(c.interval)))
	}
	return c
}

func (c retryConfig) untilState(states ...sdk.ResourceState) secapi.ResourceObserverUntilValueConfig[sdk.ResourceState] {
	return secapi.ResourceObserverUntilValueConfig[sdk.ResourceState]{
		ExpectedValues: states,
		Delay:          c.delay,
		Interval:       c.interval,
		MaxAttempts:    c.maxAttempts,
	}
}

func (c retryConfig) observer() secapi.ResourceObserverConfig {
	return secapi.ResourceObserverConfig{
		Delay:       c.delay,
		Interval:    c.interval,
		MaxAttempts: c.maxAttempts,
	}
}
