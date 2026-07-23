## Bastion host
##
## A single public-facing bastion is the only instance with a public IP and
## the only one whose security group accepts SSH from the internet. The app
## instance has no public IP at all, and its security group only accepts SSH
## from the bastion's private address - so reaching it requires hopping
## through the bastion first.

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
  name = "bastion-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "bastion-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "bastion-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.20.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "bastion-rt"
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
  name         = "bastion-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.20.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

## Only the bastion accepts SSH from the internet.
resource "seca_security_group" "bastion_sg" {
  name         = "bastion-sg"
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

## The app instance only accepts SSH from the bastion's private address.
resource "seca_security_group" "app_sg" {
  name         = "app-sg"
  workspace_id = seca_workspace.main.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 22
      }
      source_refs = ["10.20.1.10"]
    }
  ]
}

resource "seca_public_ip" "bastion" {
  name         = "bastion-ip"
  workspace_id = seca_workspace.main.id

  version = "IPv4"
}

resource "seca_nic" "bastion" {
  name         = "bastion-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses     = ["10.20.1.10"]
  public_ip_ids = [seca_public_ip.bastion.id]
}

resource "seca_nic" "app" {
  name         = "app-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses = ["10.20.1.20"]
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "bastion_boot" {
  name         = "bastion-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "bastion" {
  name         = "bastion-1"
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.bastion.id
  security_group_id = seca_security_group.bastion_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.bastion_boot.id
  }
}

resource "seca_block_storage" "app_boot" {
  name         = "app-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "app" {
  name         = "app-1"
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.app.id
  security_group_id = seca_security_group.app_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.app_boot.id
  }
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "bastion_public_ip" {
  value = seca_public_ip.bastion.address
}

output "bastion_instance_id" {
  value = seca_instance.bastion.id
}

output "app_instance_id" {
  value = seca_instance.app.id
}
