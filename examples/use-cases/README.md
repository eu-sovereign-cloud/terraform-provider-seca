# Use cases

Complete, runnable configurations that combine multiple resources to show how the provider is used for a realistic scenario, rather than documenting one resource in isolation. They are not picked up by the documentation generation tool (see [`../README.md`](../README.md)) - each is a standalone `*.tf` file with a header comment explaining the scenario it demonstrates.

| File | Scenario |
|---|---|
| [`simple-server-with-public-ip.tf`](simple-server-with-public-ip.tf) | Smallest end-to-end setup: workspace, routed network/subnet, security group, and a single instance with a public IP. |
| [`bastion-host.tf`](bastion-host.tf) | A single public-facing bastion is the only path to a private app instance, which has no public IP and only accepts SSH from the bastion. |
| [`private-secure-workspace.tf`](private-secure-workspace.tf) | Fully private instance: egress-only internet gateway, security group restricted to the corporate network, no `seca_public_ip` anywhere. |
| [`three-tier-web-app.tf`](three-tier-web-app.tf) | Web/app/db subnets each with their own scoped security group and one instance per tier; the web tier gets a reserved public IP. |
| [`ha-multi-zone.tf`](ha-multi-zone.tf) | Two replicas of a service placed in different zones and the same anti-affinity group, so they never land on the same host. |
| [`blue-green-deployment.tf`](blue-green-deployment.tf) | Two parallel instance groups ("blue" and "green") on the same network, each with its own security group and public IP, for staged cutover. |
| [`golden-image-fleet.tf`](golden-image-fleet.tf) | Build a custom image from a block storage volume, then clone identical worker instances from it via `source_image_id`. |
| [`worker-cluster-data-volumes.tf`](worker-cluster-data-volumes.tf) | Batch workers with a small boot volume plus a separate, larger `data_volumes` entry for scratch space. |
| [`labeled-multi-environment.tf`](labeled-multi-environment.tf) | One workspace holding dev/staging/prod environments tagged with an `environment` label for filtering and organization. |
| [`multi-team-rbac.tf`](multi-team-rbac.tf) | Tenant-scoped roles and role assignments giving one team read/write and another read-only access to the same workspace. |
| [`team-onboarding-scaffold.tf`](team-onboarding-scaffold.tf) | "New project" template bundling a starter workspace/network/security-group with the RBAC a new team needs from day one. |
