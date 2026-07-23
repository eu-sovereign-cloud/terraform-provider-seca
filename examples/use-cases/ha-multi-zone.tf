## Highly-available pair across zones
##
## Two instances of the same service placed in different zones and put in
## the same anti-affinity group, so the platform never schedules both
## replicas onto the same underlying host - a standard HA pattern.

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
  name = "ha-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "ha-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "ha-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.40.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "ha-rt"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.igw.id
    }
  ]
}

## Subnets are zonal, so each replica's zone needs its own subnet even
## though both share the same network and route table.
resource "seca_subnet" "zone_a" {
  name         = "ha-subnet-a"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.40.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

resource "seca_subnet" "zone_b" {
  name         = "ha-subnet-b"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.40.2.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "b"
}

resource "seca_security_group" "main" {
  name         = "ha-sg"
  workspace_id = seca_workspace.main.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    },
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

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

resource "seca_nic" "replica_a" {
  name         = "replica-a-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.zone_a.id

  addresses = ["10.40.1.10"]
}

resource "seca_block_storage" "replica_a_boot" {
  name         = "replica-a-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "replica_a" {
  name         = "replica-a"
  workspace_id = seca_workspace.main.id

  sku_id              = data.seca_instance_sku.instance_sku.id
  primary_nic_id      = seca_nic.replica_a.id
  security_group_id   = seca_security_group.main.id
  zone                = "a"
  anti_affinity_group = "ha-service"
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.replica_a_boot.id
  }
}

resource "seca_nic" "replica_b" {
  name         = "replica-b-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.zone_b.id

  addresses = ["10.40.2.10"]
}

resource "seca_block_storage" "replica_b_boot" {
  name         = "replica-b-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "replica_b" {
  name         = "replica-b"
  workspace_id = seca_workspace.main.id

  sku_id              = data.seca_instance_sku.instance_sku.id
  primary_nic_id      = seca_nic.replica_b.id
  security_group_id   = seca_security_group.main.id
  zone                = "b"
  anti_affinity_group = "ha-service"
  ssh_keys            = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.replica_b_boot.id
  }
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "replica_a_id" {
  value = seca_instance.replica_a.id
}

output "replica_a_zone" {
  value = seca_instance.replica_a.zone
}

output "replica_b_id" {
  value = seca_instance.replica_b.id
}

output "replica_b_zone" {
  value = seca_instance.replica_b.zone
}
