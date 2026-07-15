# Contributing to the SECA Terraform Provider

Thank you for your interest in contributing. This document covers the development workflow, testing requirements, and release process.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Release Process](#release-process)
- [OpenTofu Registry](#opentofu-registry)

## Development Setup

### Prerequisites

- Go >= 1.22
- Terraform CLI >= 1.13 (for acceptance tests)
- `make` (GNU Make)

### Clone and Build

```shell
git clone https://github.com/eu-sovereign-cloud/terraform-provider-seca.git
cd terraform-provider-seca
git submodule update --init --recursive
go mod download
go build -v ./...
```

### Local Provider Override

To use a local build in a Terraform configuration without publishing, add a development override to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "eu-sovereign-cloud/seca" = "/your/GOPATH/bin"
  }
  direct {}
}
```

Install your local build:

```shell
go install .
```

With this in place, `terraform init` will use the local binary and skip downloading from the registry.

## Project Structure

```
internal/provider/      # All resource, data source, and provider implementations
docs/                   # Generated documentation (do not edit by hand)
examples/               # HCL examples used by tfplugindocs and the registry
templates/              # Documentation templates for tfplugindocs
ai/                     # AI development scaffold and architecture docs (authoritative reference)
spec/                   # API specification used for SDK generation
tools/                  # Go tool dependencies (golangci-lint, tfplugindocs)
```

The `ai/` directory is the authoritative reference for conventions, patterns, and guardrails. Read `ai/guardrails.md` before making any code changes.

## Making Changes

### Before you start

- Read `ai/guardrails.md` — it contains hard rules that, if violated, will break the provider.
- Check `ai/known-issues.md` for known limitations that may affect your change.
- Route to the right AI doc by task (see CLAUDE.md for the table).

### Code conventions

- Use the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) only. Do **not** use `terraform-plugin-sdk/v2`; the linter enforces this.
- All mutating API calls are asynchronous — every Create, Update, and Delete must poll until the resource reaches `Active` state. See `ai/async-operations.md`.
- Resource names are always `RequiresReplace` — the SECA API uses name as the stable identifier.
- Follow naming conventions in `ai/provider-conventions.md`.

### Run the linter

```shell
make lint
```

### Format code

```shell
make fmt
```

## Testing

### Unit tests

```shell
make test
```

### Acceptance tests

Acceptance tests hit the real SECA API and require credentials:

```shell
export TF_ACC=1
export SECA_TOKEN="<your-bearer-token>"
export SECA_TENANT="<your-tenant>"
export SECA_REGION="<your-region>"
export SECA_REGION_V1_ENDPOINT="https://api.seca.cloud/providers/seca.region"
export SECA_AUTHORIZATION_V1_ENDPOINT="https://api.seca.cloud/providers/seca.authorization"

make testacc
```

Acceptance tests create real resources on the SECA platform. They are designed to clean up after themselves, but be aware that partial failures may leave orphaned resources.

### CI

The `test.yml` workflow runs on every PR and push. It:

1. Builds the provider and runs the linter.
2. Checks that `make generate` produces no uncommitted diff (docs must be generated and committed).
3. Runs acceptance tests against Terraform 1.13 and 1.14.

PRs will not be merged if CI is red.

## Documentation

Documentation is generated from Go source annotations and the templates in `templates/`. Do not edit files under `docs/` directly.

After any schema change, regenerate docs and commit the result:

```shell
make generate
git add docs/
git commit -m "docs: regenerate provider documentation"
```

The CI `generate` job will fail if the committed docs are out of sync with the source.

### Examples

Each resource and data source must have a corresponding example under `examples/resources/<name>/` or `examples/data-sources/<name>/`. The `tfplugindocs` tool embeds these into the generated docs.

## Submitting a Pull Request

1. Fork the repository and create a feature branch from `main`.
2. Make your changes, following the conventions above.
3. Run `make lint`, `make fmt`, and `make generate` locally and commit any changes.
4. Open a PR against `main` with a clear description of what changed and why.
5. Link any related GitHub issues.

PR titles should follow the pattern `type: short description`, where `type` is one of `feat`, `fix`, `docs`, `refactor`, `test`, or `chore`.

All PRs require at least one approving review before merge.

## Release Process

Releases are fully automated via GitHub Actions and [GoReleaser](https://goreleaser.com/). A new release is triggered by pushing a semver tag to `main`:

```shell
git tag v0.1.0
git push origin v0.1.0
```

The `release.yml` workflow will:

1. Import the GPG private key from the `GPG_PRIVATE_KEY` repository secret.
2. Build binaries for all supported OS/arch combinations.
3. Sign the SHA256SUMS checksum file with the GPG key.
4. Create a GitHub Release with all binaries, checksums, signature, and the registry manifest.

The Terraform Registry polls GitHub Releases and picks up the new version automatically within a few minutes.

### Required secrets

| Secret | Description |
|---|---|
| `GPG_PRIVATE_KEY` | ASCII-armoured GPG private key used to sign releases. |
| `PASSPHRASE` | Passphrase for the GPG key. Leave empty (or omit the secret) if the key has no passphrase. |

### Version scheme

This provider follows [Semantic Versioning](https://semver.org/):

- **0.x.y** — alpha/beta: schema may change in breaking ways between minor versions.
- **1.0.0** — stable: schema changes follow the backward-compatibility policy in `ai/requirements.md`.

## OpenTofu Registry

The release artifacts produced by GoReleaser are compatible with the [OpenTofu Registry](https://registry.opentofu.org). To list the provider there:

1. Fork [opentofu/registry](https://github.com/opentofu/registry).
2. Create the file `providers/e/eu-sovereign-cloud/seca.json` with:

   ```json
   {
     "version": 1,
     "repository": "https://github.com/eu-sovereign-cloud/terraform-provider-seca"
   }
   ```

3. Open a PR against `opentofu/registry`. The OpenTofu team will review and merge it.

Once merged, the OpenTofu Registry will automatically index all existing and future GitHub Releases. No additional workflow changes are needed.
