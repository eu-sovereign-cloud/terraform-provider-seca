## Labeled multi-environment segmentation
##
## One workspace holding three parallel environments - dev, staging, prod -
## each with its own network/subnet/security-group/instance, tagged with an
## "environment" label so they can be filtered and organized. None of the
## other use-case files exercise labels, even though the attribute exists on
## every resource.

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

locals {
  environments = {
    dev     = "10.80.0.0/16"
    staging = "10.81.0.0/16"
    prod    = "10.82.0.0/16"
  }
}

resource "seca_workspace" "shared" {
  name = "multi-env-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "multi-env-igw"
  workspace_id = seca_workspace.shared.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "env" {
  for_each = local.environments

  name         = "${each.key}-network"
  workspace_id = seca_workspace.shared.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = each.value
  }

  labels = {
    environment = each.key
  }
}

resource "seca_route_table" "env" {
  for_each = local.environments

  name         = "${each.key}-rt"
  workspace_id = seca_workspace.shared.id
  network_id   = seca_network.env[each.key].id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.igw.id
    }
  ]

  labels = {
    environment = each.key
  }
}

resource "seca_subnet" "env" {
  for_each = local.environments

  name         = "${each.key}-subnet"
  workspace_id = seca_workspace.shared.id
  network_id   = seca_network.env[each.key].id

  cidr = {
    ipv4 = cidrsubnet(each.value, 8, 1)
  }
  route_table_id = seca_route_table.env[each.key].id
  zone           = "a"

  labels = {
    environment = each.key
  }
}

resource "seca_security_group" "env" {
  for_each = local.environments

  name         = "${each.key}-sg"
  workspace_id = seca_workspace.shared.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = ["192.168.0.0/16"]
    }
  ]

  labels = {
    environment = each.key
  }
}

resource "seca_nic" "env" {
  for_each = local.environments

  name         = "${each.key}-nic"
  workspace_id = seca_workspace.shared.id
  subnet_id    = seca_subnet.env[each.key].id

  labels = {
    environment = each.key
  }
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "env_boot" {
  for_each = local.environments

  name         = "${each.key}-boot-volume"
  workspace_id = seca_workspace.shared.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.storage_sku.id

  labels = {
    environment = each.key
  }
}

resource "seca_instance" "env" {
  for_each = local.environments

  name         = "${each.key}-app"
  workspace_id = seca_workspace.shared.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.env[each.key].id
  security_group_id = seca_security_group.env[each.key].id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.env_boot[each.key].id
  }

  labels = {
    environment = each.key
  }
}

output "workspace_id" {
  value = seca_workspace.shared.id
}

output "instance_ids_by_environment" {
  value = { for env, inst in seca_instance.env : env => inst.id }
}
