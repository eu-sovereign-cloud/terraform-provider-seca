data "seca_workspace" "example" {
  name = "workspace-1"
}

data "seca_network" "example" {
  name         = "network-1"
  workspace_id = data.seca_workspace.example.id
}

data "seca_subnet" "example" {
  name         = "subnet-1"
  workspace_id = data.seca_workspace.example.id
  network_id   = data.seca_network.example.id
}

output "subnet_provider" {
  value = data.seca_subnet.example.provider
}
output "subnet_tenant" {
  value = data.seca_subnet.example.tenant
}
output "subnet_workspace" {
  value = data.seca_subnet.example.workspace
}

output "subnet_labels" {
  value = data.seca_subnet.example.labels
}
output "subnet_annotations" {
  value = data.seca_subnet.example.annotations
}
output "subnet_extensions" {
  value = data.seca_subnet.example.extensions
}

output "subnet_cidr" {
  value = data.seca_subnet.example.cidr
}
output "subnet_zone" {
  value = data.seca_subnet.example.zone
}
output "subnet_route_table_id" {
  value = data.seca_subnet.example.route_table_id
}
output "subnet_sku_id" {
  value = data.seca_subnet.example.sku_id
}

output "subnet_state" {
  value = data.seca_subnet.example.state
}
