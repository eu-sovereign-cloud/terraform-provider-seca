## Golden image + fleet
##
## Build a custom image once from a base block storage volume, then clone
## three identical worker instances from that image via source_image_id -
## the image-then-fleet pattern for provisioning many identical servers.

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
  name = "fleet-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "fleet-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "fleet-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.30.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "fleet-rt"
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
  name         = "fleet-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.30.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

resource "seca_security_group" "main" {
  name         = "fleet-sg"
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

## Build the golden image once from a base volume.
data "seca_storage_sku" "image_storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "image_base" {
  name         = "fleet-image-base-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.image_storage_sku.id
}

resource "seca_image" "golden" {
  name = "fleet-golden-image"

  block_storage_id = seca_block_storage.image_base.id
  cpu_architecture = "amd64"
  initializer      = "cloudinit-22"
  boot             = "UEFI"
}

## Clone three identical workers from the golden image.
data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "worker_storage_sku" {
  name = "RD500"
}

resource "seca_block_storage" "worker_boot" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = "${each.key}-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb         = 20
  sku_id          = data.seca_storage_sku.worker_storage_sku.id
  source_image_id = seca_image.golden.id
}

resource "seca_nic" "worker" {
  for_each = toset(["worker-1", "worker-2", "worker-3"])

  name         = "${each.key}-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id
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
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "golden_image_id" {
  value = seca_image.golden.id
}

output "worker_instance_ids" {
  value = { for k, inst in seca_instance.worker : k => inst.id }
}
