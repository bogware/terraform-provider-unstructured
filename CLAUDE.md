# CLAUDE.md

## Project Overview

This is a Terraform provider built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) (protocol v6). It was scaffolded from HashiCorp's `terraform-provider-scaffolding-framework` template. The Go module is currently `github.com/hashicorp/terraform-provider-scaffolding-framework` and the provider type name is `scaffolding`. Licensed under MPL-2.0 with IBM Corp. copyright headers.

## Repository Structure

```
├── main.go                          # Provider entrypoint (providerserver.Serve)
├── internal/provider/               # All provider logic (resources, data sources, functions, etc.)
│   ├── provider.go                  # ScaffoldingProvider - implements provider.Provider
│   ├── provider_test.go             # Test factories (testAccProtoV6ProviderFactories)
│   ├── example_resource.go          # Managed resource (CRUD lifecycle)
│   ├── example_resource_test.go     # Acceptance tests for the resource
│   ├── example_data_source.go       # Read-only data source
│   ├── example_data_source_test.go  # Acceptance tests for the data source
│   ├── example_function.go          # Provider-defined function (echo)
│   ├── example_function_test.go     # Unit tests for the function
│   ├── example_ephemeral_resource.go      # Ephemeral resource (non-persisted)
│   ├── example_ephemeral_resource_test.go # Acceptance tests for ephemeral resource
│   ├── example_action.go            # Action (side-effect operations)
│   └── example_action_test.go       # Acceptance tests for the action
├── examples/                        # Terraform HCL examples (used by doc generation)
├── docs/                            # Generated documentation (do not edit by hand)
├── tools/tools.go                   # Code generation tools (copywrite, tfplugindocs)
├── .goreleaser.yml                  # GoReleaser config for binary releases
├── .golangci.yml                    # Linter config (golangci-lint v2)
├── .copywrite.hcl                   # Copyright header automation config
├── terraform-registry-manifest.json # Registry protocol metadata (v6)
└── META.d/                          # Internal catalog metadata
```

## Build & Development Commands

All commands use the `GNUmakefile`:

| Command | What it does |
|---------|-------------|
| `make` | Runs fmt, lint, install, generate (default target) |
| `make build` | `go build -v ./...` |
| `make install` | Build + `go install -v ./...` |
| `make fmt` | `gofmt -s -w -e .` |
| `make lint` | `golangci-lint run` |
| `make generate` | Runs `go generate` in `tools/` (copyright headers, terraform fmt, tfplugindocs) |
| `make test` | Unit tests: `go test -v -cover -timeout=120s -parallel=10 ./...` |
| `make testacc` | Acceptance tests: `TF_ACC=1 go test -v -cover -timeout 120m ./...` |

## Testing

- **Unit tests**: Run with `make test` (no external dependencies needed).
- **Acceptance tests**: Run with `make testacc`. Requires `TF_ACC=1` env var and a Terraform binary in PATH. These create real resources.
- Tests live alongside source files in `internal/provider/` with `_test.go` suffix.
- Test factories are in `provider_test.go` — use `testAccProtoV6ProviderFactories` for standard tests and `testAccProtoV6ProviderFactoriesWithEcho` for ephemeral resource tests.
- Acceptance tests use `resource.Test()` with `TestStep` configs and `statecheck` assertions.
- Function tests use `resource.UnitTest()` with `tfversion.SkipBelow()` for version gating.
- Action tests require Terraform >= 1.14 (`tfversion.Version1_14_0`).
- Ephemeral resource tests require Terraform >= 1.10.
- The `testAccPreCheck` function in `provider_test.go` should validate required environment variables before tests run.

## Linting

Uses `golangci-lint` v2 with these enabled linters: `copyloopvar`, `depguard`, `durationcheck`, `errcheck`, `forcetypeassert`, `godot`, `ineffassign`, `makezero`, `misspell`, `nilerr`, `predeclared`, `staticcheck`, `unconvert`, `unparam`, `unused`, `usetesting`. Formatting via `gofmt`.

Key `depguard` rules: do NOT import from `terraform-plugin-sdk/v2` — use `terraform-plugin-framework` and `terraform-plugin-testing` instead.

## Code Conventions

- **Framework**: Uses Terraform Plugin Framework (not the older SDK). All resources/data sources implement framework interfaces.
- **Interface compliance**: Use compile-time `var _ interface = &Type{}` assertions at the top of each file.
- **Naming**: Resources/data sources follow `{ProviderTypeName}_{ResourceName}` pattern (e.g., `scaffolding_example`). Constructor functions use `New` prefix (e.g., `NewExampleResource`).
- **Models**: Each resource/data source has a corresponding `*Model` struct with `tfsdk` struct tags.
- **Configure pattern**: Resources receive provider client data via the `Configure` method, type-asserting from `req.ProviderData`.
- **Copyright headers**: All `.go` files must have `// Copyright IBM Corp. 2021, 2025` and `// SPDX-License-Identifier: MPL-2.0` headers. Run `make generate` to auto-apply.
- **Documentation**: Generated by `tfplugindocs` from examples in `examples/`. Edit examples, not docs directly.

## Adding a New Resource

1. Create `internal/provider/<name>.go` implementing `resource.Resource` (and optionally `resource.ResourceWithImportState`).
2. Create `internal/provider/<name>_test.go` with acceptance tests.
3. Register it in `provider.go` → `Resources()` method.
4. Add example HCL in `examples/resources/<provider>_<name>/resource.tf`.
5. Run `make generate` to regenerate docs.

## Adding a New Data Source

Same pattern as resources but implement `datasource.DataSource` and register in `DataSources()`.

## CI/CD

- **Tests** (`.github/workflows/test.yml`): On push/PR — builds, lints (`golangci-lint`), checks `make generate` is up to date, runs acceptance tests against Terraform 1.13.x and 1.14.x.
- **Release** (`.github/workflows/release.yml`): Triggered by `v*` tags, uses GoReleaser with GPG signing.
- **Dependabot**: Daily Go module updates, weekly GitHub Actions updates.

## Go Module

- Go version: 1.25.5 (set in `go.mod`)
- Module path: `github.com/hashicorp/terraform-provider-scaffolding-framework`
- Key dependencies: `terraform-plugin-framework`, `terraform-plugin-go`, `terraform-plugin-log`, `terraform-plugin-testing`
