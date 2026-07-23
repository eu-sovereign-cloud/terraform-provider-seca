## Multi-team RBAC governance
##
## One workspace shared by two teams with different levels of access:
## "platform-team" gets full read/write on network and compute resources,
## "app-team" gets read-only visibility into the same workspace. Roles and
## role assignments are tenant-scoped and independent of the resource graph
## they govern.

provider "seca" {
  token  = "test-token"
  tenant = "tenant-1"
  region = "region-1"
  global_providers = {
    region_v1        = "http://localhost:3000/providers/seca.region",
    authorization_v1 = "http://localhost:3000/providers/seca.authorization"
  }
}

resource "seca_workspace" "shared" {
  name = "shared-workspace"
}

resource "seca_role" "admin" {
  name = "platform-admin"

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

resource "seca_role" "viewer" {
  name = "app-viewer"

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

resource "seca_role_assignment" "platform_team" {
  name = "platform-team-assignment"

  subs = ["platform-team-svc-account"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = [seca_workspace.shared.name]
    }
  ]
  roles = [seca_role.admin.name]
}

resource "seca_role_assignment" "app_team" {
  name = "app-team-assignment"

  subs = ["app-team-svc-account"]
  scopes = [
    {
      tenants    = ["tenant-1"],
      regions    = ["region-1"],
      workspaces = [seca_workspace.shared.name]
    }
  ]
  roles = [seca_role.viewer.name]
}

output "admin_role_id" {
  value = seca_role.admin.id
}

output "viewer_role_id" {
  value = seca_role.viewer.id
}

output "platform_team_assignment_id" {
  value = seca_role_assignment.platform_team.id
}

output "app_team_assignment_id" {
  value = seca_role_assignment.app_team.id
}
