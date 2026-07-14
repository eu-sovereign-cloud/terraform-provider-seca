---
page_title: "Troubleshooting — seca Provider"
subcategory: "Guides"
description: |-
  Common issues and solutions when working with the seca Terraform provider.
---

# Troubleshooting

This guide covers the most common problems encountered when using the seca provider and how to resolve them.

## Enabling Debug Logging

Set `TF_LOG=DEBUG` (or `TF_LOG=TRACE` for maximum verbosity) before running Terraform to capture detailed provider logs:

```shell
TF_LOG=DEBUG TF_LOG_PATH=./terraform.log terraform apply
```

The provider uses structured logging via `tflog`. Look for log lines prefixed with the resource type and operation to trace exactly where a failure occurred.

## Resources Stuck in "Creating" or "Updating" State

The SECA API is eventually consistent — `terraform apply` polls until a resource reaches `Active` state. If a resource stays in `Creating` or `Updating` for a long time, the provider will return a polling timeout error.

### Default Polling Limits

| Resource | Default create/update timeout |
|---|---|
| `seca_workspace` | 10 minutes |
| `seca_image` | 15 minutes |
| `seca_block_storage` | 10 minutes |
| `seca_network`, `seca_subnet`, `seca_nic`, etc. | 5 minutes |
| `seca_instance` | 20 minutes |

### Extending Timeouts

If your environment is slower than the defaults (e.g., large image uploads, heavily loaded control plane), increase the timeout via the per-resource `timeouts` block:

```terraform
resource "seca_image" "large" {
  name             = "large-image"
  block_storage_id = seca_block_storage.source.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"

  timeouts {
    create = "30m"
    delete = "20m"
  }
}

resource "seca_instance" "heavy" {
  name           = "heavy-instance"
  workspace_id   = seca_workspace.main.id
  sku_id         = data.seca_instance_sku.xlarge.id
  primary_nic_id = seca_nic.primary.id
  zone           = "a"
  boot_volume    = { device_id = seca_block_storage.os.id }

  timeouts {
    create = "40m"
    delete = "30m"
  }
}
```

### Fine-Tuning Polling Interval

For resources that provision quickly, reduce the polling interval to get faster feedback:

```terraform
provider "seca" {
  # ...
  retry = {
    delay        = 5   # seconds before first poll
    interval     = 5   # seconds between polls
    max_attempts = 60  # stop after 60 × 5s = 5 minutes
  }
}
```

## Eventual Consistency: Resources Not Found After Creation

Occasionally, reading a resource immediately after creation returns `404 Not Found` because the SECA API has not yet propagated the change. The provider handles this transparently by polling `GetXxxUntilState()`. If you see a `404` error in plan output (not during apply), re-run `terraform plan` — the transient state usually resolves within seconds.

## Common API Errors

### `401 Unauthorized`

Your bearer token has expired or is invalid. Re-obtain a fresh token and update your `seca_token` variable.

### `403 Forbidden`

The token has insufficient scopes for the operation. Check the required API scope for the resource you are managing (see the [Authentication guide](authentication.md)).

### `409 Conflict`

A resource with the same name already exists. Either import the existing resource (`terraform import`) or choose a different name.

### `422 Unprocessable Entity`

The API rejected the request due to invalid field values (e.g., a CIDR block that overlaps with an existing network). Check the error message for the specific field and correct it in your configuration.

### `503 Service Unavailable`

The SECA control plane is temporarily unavailable. Wait a few minutes and retry. Consider increasing `retry.interval` and `retry.max_attempts` if this happens frequently in your environment.

## Terraform State Drift

If resources are modified outside of Terraform (e.g., via the SECA portal or API directly), the next `terraform plan` will show unexpected differences. Run `terraform refresh` to synchronise state with the current API state, then review the plan before applying.

## Destroying Resources That Are in Use

Attempting to destroy a resource that other resources depend on (e.g., deleting a workspace while instances are still running) will fail with an API error. Destroy dependent resources first, or use `depends_on` / resource ordering to control the destroy order.

## Getting Help

If you encounter an issue not covered here, please open a bug report at the provider repository including:

1. The Terraform version (`terraform version`)
2. The provider version
3. The relevant Terraform configuration (redact sensitive values)
4. The full error message and debug log excerpt (`TF_LOG=DEBUG`)
