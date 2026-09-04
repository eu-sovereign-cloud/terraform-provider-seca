## Instance booted from a catalog image
##
## Minimal graph for one SSH-able server whose boot volume is created from a
## public catalog image instead of an empty volume:
##
##   workspace -> network + public_ip -> route_table -> subnet -> nic
##             -> image-backed block_storage -> instance
##
## Differences from `simple-server-with-public-ip.tf`:
##   * the boot volume sets `source_image_id` from the `seca_image` data
##     source, so the image (and the injected `ssh_keys`) come from the region
##     catalog;
##   * no `seca_internet_gateway` - the default route targets the reserved
##     public IP directly;
##   * no `seca_security_group` - the instance inherits the network default.
##
## Only the values that must match your environment are variables; everything
## else is inlined to keep the graph readable. Note that `terraform apply` only
## records the desired state: image-backed volumes and the NIC are provisioned
## by the platform when the instance is powered on.

variable "tenant" {
  description = "SECA tenant. Must be the tenant the SKU/image catalog is seeded under."
  type        = string
}

variable "region" {
  description = "SECA region, e.g. \"itbg-bergamo\"."
  type        = string
}

variable "region_v1_endpoint" {
  description = "seca.region provider URL, e.g. \"https://demo.secapi.cloud/providers/seca.region\"."
  type        = string
}

variable "ssh_keys" {
  description = "SSH public keys injected into the boot volume - this is how you log in."
  type        = list(string)
}

provider "seca" {
  tenant = var.tenant
  region = var.region
  global_providers = {
    region_v1 = var.region_v1_endpoint
  }
}

resource "seca_workspace" "main" {
  name = "image-boot"
}

## --- Networking -------------------------------------------------------------

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "main" {
  name         = "image-boot-network"
  workspace_id = seca_workspace.main.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.100.0.0/16"
  }
}

## Reserved up front and used as the default-route target below; it is attached
## to the primary NIC when the instance starts.
resource "seca_public_ip" "main" {
  name         = "image-boot-ip"
  workspace_id = seca_workspace.main.id

  version = "IPv4"
}

## The API requires at least one route on a route table.
resource "seca_route_table" "main" {
  name         = "image-boot-rt"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_public_ip.main.id
    }
  ]
}

resource "seca_subnet" "main" {
  name         = "image-boot-subnet"
  workspace_id = seca_workspace.main.id
  network_id   = seca_network.main.id

  cidr = {
    ipv4 = "10.100.1.0/24"
  }
  route_table_id = seca_route_table.main.id
  zone           = "a"
}

resource "seca_nic" "main" {
  name         = "image-boot-nic"
  workspace_id = seca_workspace.main.id
  subnet_id    = seca_subnet.main.id

  addresses     = ["10.100.1.10"]
  public_ip_ids = [seca_public_ip.main.id]
}

## --- Storage ----------------------------------------------------------------

## Public catalog image; its name must exist in the region catalog.
data "seca_image" "boot" {
  name = "ubuntu-24-04"
}

data "seca_storage_sku" "storage_sku" {
  name = "RD500"
}

resource "seca_block_storage" "boot" {
  name         = "image-boot-volume"
  workspace_id = seca_workspace.main.id

  size_gb         = 10
  sku_id          = data.seca_storage_sku.storage_sku.id
  source_image_id = data.seca_image.boot.id
}

## --- Compute ----------------------------------------------------------------

data "seca_instance_sku" "instance_sku" {
  name = "DXS"
}

resource "seca_instance" "main" {
  name         = "image-boot-1"
  workspace_id = seca_workspace.main.id

  sku_id         = data.seca_instance_sku.instance_sku.id
  primary_nic_id = seca_nic.main.id
  zone           = "a"
  ssh_keys       = var.ssh_keys

  boot_volume = {
    device_id = seca_block_storage.boot.id
  }

  timeouts {
    create = "30m"
    update = "30m"
    delete = "30m"
  }
}

output "instance_id" {
  value = seca_instance.main.id
}

output "public_ip_address" {
  description = "SSH target. Empty until the instance is started and the address is assigned."
  value       = seca_public_ip.main.ip_address
}
