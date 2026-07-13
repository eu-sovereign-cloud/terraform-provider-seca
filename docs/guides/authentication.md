---
page_title: "Authentication — seca Provider"
subcategory: "Guides"
description: |-
  How to obtain a bearer token and configure the seca provider for authentication.
---

# Authentication

The seca provider authenticates to the SECA platform using a **bearer token**. This guide explains how to obtain a token and configure the provider securely.

## Obtaining a Bearer Token

Bearer tokens are issued by the SECA Identity service. Refer to the SECA platform documentation for your tenant to create a service-account token with the appropriate scopes for the resources you intend to manage.

Once you have a token, store it in an environment variable instead of hard-coding it in your Terraform configuration:

```shell
export SECA_TOKEN="<your-token-here>"
```

## Minimal Provider Configuration

The minimum required configuration is `token`, `tenant`, `region`, and the `global_providers` block pointing at the control-plane endpoints for your region:

```terraform
provider "seca" {
  token  = var.seca_token   # or use SECA_TOKEN env var indirectly via variable
  tenant = "my-tenant"
  region = "eu-central-1"

  global_providers = {
    region_v1 = "https://api.seca.cloud/providers/seca.region"
  }
}

variable "seca_token" {
  type      = string
  sensitive = true
}
```

## Full Provider Configuration

When managing IAM resources (`seca_role`, `seca_role_assignment`) you must also provide the `authorization_v1` endpoint. The optional `retry` block fine-tunes polling behaviour for async API operations:

```terraform
provider "seca" {
  token  = var.seca_token
  tenant = "my-tenant"
  region = "eu-central-1"

  global_providers = {
    region_v1        = "https://api.seca.cloud/providers/seca.region"
    authorization_v1 = "https://api.seca.cloud/providers/seca.authorization"
  }

  retry = {
    delay        = 30  # seconds before first status poll
    interval     = 10  # seconds between subsequent polls
    max_attempts = 30  # maximum polling attempts (≈ 5 minutes total)
  }
}
```

## Using Environment Variables

To avoid placing credentials in Terraform files (which might be committed to version control), pass the token through a variable that is supplied at runtime:

```shell
# terraform.tfvars (gitignored) or via -var flag
TF_VAR_seca_token="<your-token-here>" terraform apply
```

Alternatively, use a secrets manager (Vault, AWS Secrets Manager, etc.) to inject `var.seca_token` at plan time.

## Token Scopes

The token must have sufficient permissions for the SECA API calls your configuration will make:

| Resource type | Required API scope |
|---|---|
| `seca_workspace` | `WorkspaceV1` write |
| `seca_network`, `seca_subnet`, `seca_nic`, etc. | `NetworkV1` write |
| `seca_block_storage`, `seca_image` | `StorageV1` write |
| `seca_instance` | `ComputeV1` write |
| `seca_role`, `seca_role_assignment` | `AuthorizationV1` admin |

Data sources require the corresponding read scope only.
