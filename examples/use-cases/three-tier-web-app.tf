## Three-tier web application
##
## A workspace with a network split into web / app / db subnets, each with a
## security group scoped to only the traffic that tier needs, and one
## instance per tier. The web tier gets a reserved public IP.

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

resource "seca_workspace" "app" {
  name = "ecommerce-prod"
}

resource "seca_internet_gateway" "igw" {
  name         = "ecommerce-igw"
  workspace_id = seca_workspace.app.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "ecommerce-network"
  workspace_id = seca_workspace.app.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.0.0.0/16"
  }
}

resource "seca_route_table" "public_rt" {
  name         = "ecommerce-public-rt"
  workspace_id = seca_workspace.app.id
  network_id   = seca_network.main.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.igw.id
    }
  ]
}

resource "seca_subnet" "web" {
  name         = "web-subnet"
  workspace_id = seca_workspace.app.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.0.1.0/24"
  }
  route_table_id = seca_route_table.public_rt.id
  zone           = "a"
}

resource "seca_subnet" "app" {
  name         = "app-subnet"
  workspace_id = seca_workspace.app.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.0.2.0/24"
  }
  route_table_id = seca_route_table.public_rt.id
  zone           = "a"
}

resource "seca_subnet" "db" {
  name         = "db-subnet"
  workspace_id = seca_workspace.app.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.0.3.0/24"
  }
  route_table_id = seca_route_table.public_rt.id
  zone           = "a"
}

resource "seca_security_group" "web_sg" {
  name         = "web-sg"
  workspace_id = seca_workspace.app.id

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

resource "seca_security_group" "app_sg" {
  name         = "app-sg"
  workspace_id = seca_workspace.app.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 8080
      }
      source_refs = ["10.0.1.0/24"]
    }
  ]
}

resource "seca_security_group" "db_sg" {
  name         = "db-sg"
  workspace_id = seca_workspace.app.id

  rules = [
    {
      direction = "ingress"
      protocol  = "tcp"
      ports = {
        from = 5432
      }
      source_refs = ["10.0.2.0/24"]
    }
  ]
}

resource "seca_public_ip" "web" {
  name         = "ecommerce-web-ip"
  workspace_id = seca_workspace.app.id

  version = "IPv4"
}

resource "seca_nic" "web" {
  name         = "web-nic"
  workspace_id = seca_workspace.app.id
  subnet_id    = seca_subnet.web.id

  addresses     = ["10.0.1.10"]
  public_ip_ids = [seca_public_ip.web.id]
}

resource "seca_nic" "app" {
  name         = "app-nic"
  workspace_id = seca_workspace.app.id
  subnet_id    = seca_subnet.app.id

  addresses = ["10.0.2.10"]
}

resource "seca_nic" "db" {
  name         = "db-nic"
  workspace_id = seca_workspace.app.id
  subnet_id    = seca_subnet.db.id

  addresses = ["10.0.3.10"]
}

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

data "seca_storage_sku" "web_storage_sku" {
  name = "RD100"
}

resource "seca_block_storage" "web_storage" {
  name         = "web-boot-volume"
  workspace_id = seca_workspace.app.id

  size_gb = 10
  sku_id  = data.seca_storage_sku.web_storage_sku.id
}

resource "seca_instance" "web" {
  name         = "web-1"
  workspace_id = seca_workspace.app.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.web.id
  security_group_id = seca_security_group.web_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.web_storage.id
  }
}

data "seca_storage_sku" "app_storage_sku" {
  name = "RD500"
}

resource "seca_block_storage" "app_storage" {
  name         = "app-boot-volume"
  workspace_id = seca_workspace.app.id

  size_gb = 20
  sku_id  = data.seca_storage_sku.app_storage_sku.id
}

resource "seca_instance" "app" {
  name         = "app-1"
  workspace_id = seca_workspace.app.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.app.id
  security_group_id = seca_security_group.app_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.app_storage.id
  }
}

data "seca_storage_sku" "db_storage_sku" {
  name = "RD500"
}

resource "seca_block_storage" "db_storage" {
  name         = "db-boot-volume"
  workspace_id = seca_workspace.app.id

  size_gb = 50
  sku_id  = data.seca_storage_sku.db_storage_sku.id
}

resource "seca_instance" "db" {
  name         = "db-1"
  workspace_id = seca_workspace.app.id

  sku_id            = data.seca_instance_sku.instance_sku.id
  primary_nic_id    = seca_nic.db.id
  security_group_id = seca_security_group.db_sg.id
  zone              = "a"
  ssh_keys          = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl example@secapi.cloud"]

  boot_volume = {
    device_id = seca_block_storage.db_storage.id
  }
}

output "workspace_id" {
  value = seca_workspace.app.id
}

output "web_public_ip" {
  value = seca_public_ip.web.address
}

output "web_instance_id" {
  value = seca_instance.web.id
}

output "web_instance_power_state" {
  value = seca_instance.web.power_state
}

output "app_instance_id" {
  value = seca_instance.app.id
}

output "app_instance_power_state" {
  value = seca_instance.app.power_state
}

output "db_instance_id" {
  value = seca_instance.db.id
}

output "db_instance_power_state" {
  value = seca_instance.db.power_state
}
