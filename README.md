# Terraform Provider for SECA

[![Tests](https://github.com/eu-sovereign-cloud/terraform-provider-seca/actions/workflows/test.yml/badge.svg)](https://github.com/eu-sovereign-cloud/terraform-provider-seca/actions/workflows/test.yml)
[![Release](https://github.com/eu-sovereign-cloud/terraform-provider-seca/actions/workflows/release.yml/badge.svg)](https://github.com/eu-sovereign-cloud/terraform-provider-seca/actions/workflows/release.yml)

> **Alpha release.** APIs and schema attributes may change between minor versions.

The `seca` provider manages infrastructure on the **SECA (Sovereign European Cloud API)** platform — a sovereign, EU-based cloud infrastructure service. It exposes compute, storage, networking, and IAM resources as Terraform-managed infrastructure.

- Registry: [`registry.terraform.io/providers/eu-sovereign-cloud/seca`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca)
- OpenTofu: [`registry.opentofu.org/eu-sovereign-cloud/seca`](https://registry.opentofu.org/eu-sovereign-cloud/seca)
- Documentation: [Terraform Registry docs](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.13 or [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.8
- [Go](https://golang.org/doc/install) >= 1.22 (only for building from source)

## Quick Start

Add the provider to your Terraform configuration:

```terraform
terraform {
  required_providers {
    seca = {
      source  = "eu-sovereign-cloud/seca"
      version = "~> 0.1"
    }
  }
}

provider "seca" {
  token  = var.seca_token
  tenant = "my-tenant"
  region = "eu-central-1"

  global_providers = {
    region_v1        = "https://api.seca.cloud/providers/seca.region"
    authorization_v1 = "https://api.seca.cloud/providers/seca.authorization"
  }
}

variable "seca_token" {
  type      = string
  sensitive = true
}
```

Set your credentials and apply:

```shell
export TF_VAR_seca_token="<your-bearer-token>"
terraform init
terraform plan
```

## Authentication

The provider authenticates using a **bearer token** issued by the SECA Identity service. See the [Authentication guide](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/guides/authentication) for full details on obtaining a token and configuring the required API scopes.

## Provider Configuration

| Argument | Required | Description |
|---|---|---|
| `token` | Yes | Bearer token for API authentication. |
| `tenant` | Yes | Tenant (organisation) identifier. |
| `region` | Yes | Region identifier (e.g. `eu-central-1`). |
| `global_providers` | Yes | Control-plane endpoint URLs (`region_v1`, `authorization_v1`). |
| `retry` | No | Polling configuration for async operations (`delay`, `interval`, `max_attempts`). |

## Resources

| Resource | Description |
|---|---|
| [`seca_workspace`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/workspace) | Logical grouping of related resources within a tenant. |
| [`seca_image`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/image) | OS image for instance boot volumes. |
| [`seca_block_storage`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/block_storage) | Persistent block storage volume. |
| [`seca_instance`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/instance) | Compute instance (VM). |
| [`seca_network`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/network) | Virtual network with IPv4 CIDR. |
| [`seca_subnet`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/subnet) | Subnet within a network. |
| [`seca_nic`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/nic) | Network interface card attached to a subnet. |
| [`seca_internet_gateway`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/internet_gateway) | Internet gateway for outbound connectivity. |
| [`seca_public_ip`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/public_ip) | Public IP address (IPv4 or IPv6). |
| [`seca_route_table`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/route_table) | Custom route table attached to a network. |
| [`seca_security_group`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/security_group) | Stateful firewall rules for instances. |
| [`seca_role`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/resources/role) | IAM role for RBAC. |

## Data Sources

| Data Source | Description |
|---|---|
| [`seca_region`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/region) | Available regions. |
| [`seca_workspace`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/workspace) | Look up an existing workspace. |
| [`seca_image`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/image) | Look up an existing image. |
| [`seca_block_storage`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/block_storage) | Look up an existing block storage volume. |
| [`seca_instance`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/instance) | Look up an existing compute instance. |
| [`seca_network`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/network) | Look up an existing network. |
| [`seca_subnet`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/subnet) | Look up an existing subnet. |
| [`seca_nic`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/nic) | Look up an existing network interface. |
| [`seca_internet_gateway`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/internet_gateway) | Look up an existing internet gateway. |
| [`seca_public_ip`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/public_ip) | Look up an existing public IP address. |
| [`seca_route_table`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/route_table) | Look up an existing route table. |
| [`seca_security_group`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/security_group) | Look up an existing security group. |
| [`seca_role`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/role) | Look up an existing IAM role. |
| [`seca_storage_sku`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/storage_sku) | Available storage SKUs. |
| [`seca_network_sku`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/network_sku) | Available network SKUs. |
| [`seca_instance_sku`](https://registry.terraform.io/providers/eu-sovereign-cloud/seca/latest/docs/data-sources/instance_sku) | Available compute instance SKUs. |

## Building from Source

```shell
git clone https://github.com/eu-sovereign-cloud/terraform-provider-seca.git
cd terraform-provider-seca
go build -v ./...
```

To test a local build without publishing, use a [Terraform development override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "eu-sovereign-cloud/seca" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

Then install:

```shell
go install .
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and PR guidelines.

## License

[Mozilla Public License 2.0](LICENSE)
