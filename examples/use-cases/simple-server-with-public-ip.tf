## Simple server with a public IP
##
## The smallest end-to-end scenario: one workspace, one network/subnet routed
## to the internet through a gateway, one security group allowing SSH and
## HTTP, and a single instance whose NIC has a public IP attached.

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
  name = "simple-server"
}

resource "seca_internet_gateway" "igw" {
  name         = "simple-server-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "simple-server-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.10.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "simple-server-rt"
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
  name         = "simple-server-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.10.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

resource "seca_security_group" "main" {
  name         = "simple-server-sg"
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

resource "seca_public_ip" "main" {
  name         = "simple-server-ip"
  workspace_id = seca_workspace.main.id

  version = "IPv4"
}

resource "seca_nic" "main" {
  name         = "simple-server-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses     = ["10.10.1.10"]
  public_ip_ids = [seca_public_ip.main.id]
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "boot" {
  name         = "simple-server-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "main" {
  name         = "simple-server-1"
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.main.id
  security_group_id = seca_security_group.main.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.boot.id
  }
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "instance_id" {
  value = seca_instance.main.id
}

output "public_ip_address" {
  value = seca_public_ip.main.address
}

output "instance_power_state" {
  value = seca_instance.main.power_state
}
