## Worker cluster with separate data volumes
##
## Three identical batch workers, each with a small boot volume plus a
## separate, larger data volume for scratch space - demonstrating
## seca_instance.data_volumes, which the other use-case files don't exercise
## (they only ever set boot_volume).

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

resource "seca_workspace" "main" {
  name = "batch-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "batch-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "batch-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.50.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "batch-rt"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.igw.id
    }
  ]
}

resource "seca_subnet" "main" {
  name         = "batch-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.50.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

resource "seca_security_group" "main" {
  name         = "batch-sg"
  workspace_id = seca_workspace.main.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = ["203.0.113.10"]
    }
  ]
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "boot_storage_sku" {
  name = "RD100"
}

data "seca_storage_sku" "data_storage_sku" {
  name = "RD500"
}

resource "seca_nic" "worker" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = "${each.key}-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id
}

resource "seca_block_storage" "worker_boot" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = "${each.key}-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.boot_storage_sku.id
}

resource "seca_block_storage" "worker_data" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = "${each.key}-data-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 200
  sku_id  = data.seca_storage_sku.data_storage_sku.id
}

resource "seca_instance" "worker" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = each.key
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.worker[each.key].id
  security_group_id = seca_security_group.main.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.worker_boot[each.key].id
  }

  data_volumes = [
    {
      device_id = seca_block_storage.worker_data[each.key].id
    }
  ]
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "worker_instance_ids" {
  value = { for k, inst in seca_instance.worker : k => inst.id }
}

output "worker_data_volume_ids" {
  value = { for k, vol in seca_block_storage.worker_data : k => vol.id }
}
