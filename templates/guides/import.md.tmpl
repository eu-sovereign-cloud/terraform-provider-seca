---
page_title: "Importing Resources — seca Provider"
subcategory: "Guides"
description: |-
  How to import existing SECA resources into Terraform state using terraform import.
---

# Importing Existing Resources

If you have resources already provisioned on the SECA platform that you want to bring under Terraform management, use `terraform import`. This guide shows the import identifier format for each resource type.

## Import ID Format

Resources are identified in one of two ways depending on their scope:

| Scope | Format | Example |
|---|---|---|
| Tenant-scoped | `<name>` | `my-workspace` |
| Workspace-scoped | `<workspace_id>/<name>` | `ws-abc123/my-network` |

The import command syntax is:
```shell
terraform import <resource_address> <import_id>
```

## Resource Import Examples

### seca_workspace

Workspaces are tenant-scoped; use the workspace name:

```shell
terraform import seca_workspace.main my-workspace
```

```terraform
resource "seca_workspace" "main" {
  name = "my-workspace"
}
```

### seca_image

Images are tenant-scoped; use the image name:

```shell
terraform import seca_image.base ubuntu-22-amd64
```

```terraform
resource "seca_image" "base" {
  name             = "ubuntu-22-amd64"
  block_storage_id = data.seca_block_storage.source.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}
```

### seca_block_storage

Block storage volumes are workspace-scoped:

```shell
terraform import seca_block_storage.data "ws-abc123/my-volume"
```

```terraform
resource "seca_block_storage" "data" {
  name         = "my-volume"
  workspace_id = seca_workspace.main.id
  size_gb      = 100
  sku_id       = data.seca_storage_sku.fast.id
}
```

### seca_network

Networks are workspace-scoped:

```shell
terraform import seca_network.vpc "ws-abc123/my-network"
```

```terraform
resource "seca_network" "vpc" {
  name         = "my-network"
  workspace_id = seca_workspace.main.id
  sku_id       = data.seca_network_sku.default.id
  cidr = {
    ipv4 = "10.100.0.0/16"
  }
}
```

### seca_nic

NICs are workspace-scoped:

```shell
terraform import seca_nic.primary "ws-abc123/my-nic"
```

### seca_instance

Instances are workspace-scoped:

```shell
terraform import seca_instance.web "ws-abc123/my-instance"
```

```terraform
resource "seca_instance" "web" {
  name           = "my-instance"
  workspace_id   = seca_workspace.main.id
  sku_id         = data.seca_instance_sku.small.id
  primary_nic_id = seca_nic.primary.id
  zone           = "a"
  boot_volume = {
    device_id = seca_block_storage.os.id
  }
}
```

### seca_role

Roles are tenant-scoped:

```shell
terraform import seca_role.network_reader network-reader
```

### seca_role_assignment

Role assignments are tenant-scoped:

```shell
terraform import seca_role_assignment.my_assign my-assignment
```

## Handling Drift After Import

After importing, run `terraform plan` to detect any differences between the imported state and your Terraform configuration. Common sources of drift:

- **Computed fields**: Fields like `tenant`, `region`, `created_at` are set by the API and should be left out of your configuration or declared as `Computed`.
- **Optional fields with defaults**: If you omit `labels`, `annotations`, or `extensions` in your configuration but the resource has them, Terraform will plan to remove them. Add them to your config to avoid unintended changes.
- **Immutable fields**: Some fields (e.g. `cidr` on `seca_network`, `sku_id` on `seca_block_storage`) cannot be changed after creation. If they differ, Terraform will plan a destroy-and-recreate.

Always review the plan carefully after import before applying.
