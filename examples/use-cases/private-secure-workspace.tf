## Private, secure workspace
##
## An instance with no public exposure at all: the internet gateway is
## egress-only (outbound internet access, no inbound path), the only
## security group rule allows traffic from inside the corporate network, and
## there is no seca_public_ip resource anywhere in this file.

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

resource "seca_workspace" "internal" {
  name = "internal-services"
}

resource "seca_internet_gateway" "egress_only" {
  name         = "internal-egress-gw"
  workspace_id = seca_workspace.internal.id

  egress_only = true
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "internal" {
  name         = "internal-network"
  workspace_id = seca_workspace.internal.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.50.0.0/16"
  }
}

resource "seca_route_table" "internal_rt" {
  name         = "internal-rt"
  workspace_id = seca_workspace.internal.id
  network_id   = seca_network.internal.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.egress_only.id
    }
  ]
}

resource "seca_subnet" "internal" {
  name         = "internal-subnet"
  workspace_id = seca_workspace.internal.id
  network_id   = seca_network.internal.id

  cidr = {
    ipv4 = "10.50.1.0/24"
  }
  route_table_id = seca_route_table.internal_rt.id
  zone           = "a"
}

resource "seca_security_group" "internal_only" {
  name         = "internal-only-sg"
  workspace_id = seca_workspace.internal.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = ["192.168.0.0/16"]
    },
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [8443]
      }
      source_refs = ["192.168.0.0/16"]
    }
  ]
}

resource "seca_nic" "internal" {
  name         = "internal-nic"
  workspace_id = seca_workspace.internal.id
  subnet_id    = seca_subnet.internal.id

  addresses = ["10.50.1.10"]
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "boot" {
  name         = "internal-boot-volume"
  workspace_id = seca_workspace.internal.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "internal" {
  name         = "internal-service-1"
  workspace_id = seca_workspace.internal.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.internal.id
  security_group_id = seca_security_group.internal_only.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.boot.id
  }
}

output "workspace_id" {
  value = seca_workspace.internal.id
}

output "instance_id" {
  value = seca_instance.internal.id
}

output "instance_power_state" {
  value = seca_instance.internal.power_state
}
