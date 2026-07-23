## Team onboarding scaffold
##
## A "new project" template: a fresh workspace with a minimal starter
## network/subnet/security-group, plus the RBAC roles and assignments a new
## team needs from day one. Unlike multi-team-rbac.tf (which is RBAC only),
## this bundles governance together with the baseline infrastructure a team
## actually gets handed.

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

resource "seca_workspace" "team" {
  name = "new-team-workspace"
}

resource "seca_internet_gateway" "igw" {
  name         = "new-team-igw"
  workspace_id = seca_workspace.team.id
}

data "seca_network_sku" "network_sku" {
  name = "N10K"
}

resource "seca_network" "starter" {
  name         = "new-team-network"
  workspace_id = seca_workspace.team.id

  sku_id = data.seca_network_sku.network_sku.id
  cidr = {
    ipv4 = "10.70.0.0/16"
  }
}

resource "seca_route_table" "starter" {
  name         = "new-team-rt"
  workspace_id = seca_workspace.team.id
  network_id   = seca_network.starter.id

  routes = [
    {
      destination_cidr_block = "0.0.0.0/0"
      target_id              = seca_internet_gateway.igw.id
    }
  ]
}

resource "seca_subnet" "starter" {
  name         = "new-team-subnet"
  workspace_id = seca_workspace.team.id
  network_id   = seca_network.starter.id

  cidr = {
    ipv4 = "10.70.1.0/24"
  }
  route_table_id = seca_route_table.starter.id
  zone           = "a"
}

## Baseline: SSH only, from the corporate range, until the team opens up
## anything more.
resource "seca_security_group" "baseline" {
  name         = "new-team-baseline-sg"
  workspace_id = seca_workspace.team.id

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
}

## Governance: the team owns its workspace, platform ops keeps read access.
resource "seca_role" "team_owner" {
  name = "new-team-owner"

  permissions = [
    {
      provider = "seca.network/v1",
      resources = [
        "networks/*",
        "subnets/*",
        "route-tables/*",
        "nics/*",
        "internet-gateways/*",
        "security-groups/*",
        "public-ips/*"
      ],
      verb = ["get", "list", "create", "update", "delete"]
    },
    {
      provider = "seca.compute/v1",
      resources = [
        "instances/*"
      ],
      verb = ["get", "list", "create", "update", "delete"]
    }
  ]
}

resource "seca_role" "platform_ops_viewer" {
  name = "platform-ops-viewer"

  permissions = [
    {
      provider = "seca.network/v1",
      resources = [
        "networks/*",
        "subnets/*",
        "route-tables/*",
        "nics/*",
        "internet-gateways/*",
        "security-groups/*",
        "public-ips/*"
      ],
      verb = ["get", "list"]
    },
    {
      provider = "seca.compute/v1",
      resources = [
        "instances/*"
      ],
      verb = ["get", "list"]
    }
  ]
}

resource "seca_role_assignment" "team_owner_assignment" {
  name = "new-team-owner-assignment"

  subs = ["new-team-svc-account"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = [seca_workspace.team.name]
    }
  ]
  roles = [seca_role.team_owner.name]
}

resource "seca_role_assignment" "platform_ops_assignment" {
  name = "platform-ops-viewer-assignment"

  subs = ["platform-ops-svc-account"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = [seca_workspace.team.name]
    }
  ]
  roles = [seca_role.platform_ops_viewer.name]
}

output "workspace_id" {
  value = seca_workspace.team.id
}

output "team_owner_role_id" {
  value = seca_role.team_owner.id
}

output "team_owner_assignment_id" {
  value = seca_role_assignment.team_owner_assignment.id
}
