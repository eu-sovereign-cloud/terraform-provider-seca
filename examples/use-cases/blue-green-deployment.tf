## Blue/green deployment
##
## Two parallel instance groups on the same network - "blue" (current
## production) and "green" (the new version being staged) - each with its
## own security group and public IP so both can be reached and compared
## independently before cutting traffic over.

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
  name = "bluegreen-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "bluegreen-igw"
  workspace_id = seca_workspace.main.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "bluegreen-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.60.0.0/16"
  }
}

resource "seca_route_table" "main" {
  name         = "bluegreen-rt"
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
  name         = "bluegreen-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.60.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD100"
}

## Blue: the current production version.
resource "seca_security_group" "blue_sg" {
  name         = "blue-sg"
  workspace_id = seca_workspace.main.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    }
  ]
}

resource "seca_public_ip" "blue" {
  name         = "blue-ip"
  workspace_id = seca_workspace.main.id

  version = "IPv4"
}

resource "seca_nic" "blue" {
  name         = "blue-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses     = ["10.60.1.10"]
  public_ip_ids = [seca_public_ip.blue.id]
}

resource "seca_block_storage" "blue_boot" {
  name         = "blue-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "blue" {
  name         = "app-blue"
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.blue.id
  security_group_id = seca_security_group.blue_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.blue_boot.id
  }
}

## Green: the new version being staged, reachable on its own IP for testing
## before it takes over the "blue" role.
resource "seca_security_group" "green_sg" {
  name         = "green-sg"
  workspace_id = seca_workspace.main.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        list = [80, 443]
      }
      source_refs = []
    }
  ]
}

resource "seca_public_ip" "green" {
  name         = "green-ip"
  workspace_id = seca_workspace.main.id

  version = "IPv4"
}

resource "seca_nic" "green" {
  name         = "green-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses     = ["10.60.1.20"]
  public_ip_ids = [seca_public_ip.green.id]
}

resource "seca_block_storage" "green_boot" {
  name         = "green-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.storage_sku.id
}

resource "seca_instance" "green" {
  name         = "app-green"
  workspace_id = seca_workspace.main.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.green.id
  security_group_id = seca_security_group.green_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.green_boot.id
  }
}

output "workspace_id" {
  value = seca_workspace.main.id
}

output "blue_public_ip" {
  value = seca_public_ip.blue.address
}

output "blue_instance_id" {
  value = seca_instance.blue.id
}

output "green_public_ip" {
  value = seca_public_ip.green.address
}

output "green_instance_id" {
  value = seca_instance.green.id
}
