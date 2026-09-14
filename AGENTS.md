# stackit-operator — Agent Guide

This document provides guidance for AI coding agents (GitHub Copilot, Codex,
Claude, etc.) working in this repository.

---

## Git Workflow — MANDATORY

- **Never commit changes** unless explicitly instructed to do so.
- **Never create a branch** unless explicitly instructed to do so.
- **Never open a pull request** unless explicitly instructed to do so.
- Leave all changes as unstaged working-tree modifications by default.

---

## Project Overview

`stackit-operator` is a **Kubernetes operator** written in Go using
[controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
(kubebuilder pattern). It manages [STACKIT](https://www.stackit.de/) Postgres
Flex databases and users declaratively via Kubernetes custom resources.

The operator authenticates against the STACKIT API using credentials held in
`ServiceAccountCredentials` (namespaced) or `ClusterServiceAccountCredentials`
(cluster-scoped) resources, resolves a Postgres Flex instance from a
human-readable `instanceName`, and reconciles the desired state expressed in
`PostgresDatabase`/`PostgresUser` CRDs into the live STACKIT Postgres Flex API
using [stackit-sdk-go](https://github.com/stackitcloud/stackit-sdk-go).

---

## Repository Layout

```text
cmd/
  main.go                        # Entrypoint — sets up the controller-runtime manager

api/
  common/
    ref.go                       # Shared types: StackitCredentialsRef, SecretKeySelector, SecretKeyRefOrVal
    zz_generated.deepcopy.go     # Auto-generated — do not edit
  v1alpha1/
    serviceaccountcredentials_types.go        # ServiceAccountCredentials (namespaced)
    clusterserviceaccountcredentials_types.go # ClusterServiceAccountCredentials (cluster-scoped)
    groupversion_info.go
    zz_generated.deepcopy.go     # Auto-generated — do not edit
  postgres/v1alpha1/
    postgresdatabase_types.go    # PostgresDatabase
    postgresuser_types.go        # PostgresUser
    groupversion_info.go
    zz_generated.deepcopy.go     # Auto-generated — do not edit

internal/
  controller/
    serviceaccountcredentials_controller.go        # Verifies STACKIT API connectivity and updates .status.connected
    clusterserviceaccountcredentials_controller.go # Same, for the cluster-scoped variant
    stackit_client.go             # Resolves StackitCredentialsRef into an authenticated STACKIT SDK client
    secret_watch.go                # Requeues credential consumers when a referenced Secret changes
    controller_permissions.go     # Shared RBAC markers for secrets/configmaps
    postgres/
      postgres_client.go          # Postgres Flex session/instance-lookup helpers (resolves instanceName -> instance ID)
      postgresdatabase_controller.go
      postgresuser_controller.go  # Also writes a connection-details Secret and manages password rotation
  webhook/
    v1alpha1/                     # Defaulting + validating webhooks for ServiceAccountCredentials/ClusterServiceAccountCredentials
    postgres/v1alpha1/            # Defaulting + validating webhooks for PostgresDatabase/PostgresUser
      credentialsref_helpers.go   # Shared validation for spec.credentialsRef

config/
  crd/bases/                     # Generated CRD YAML — do not edit manually
  rbac/                          # Generated RBAC roles — do not edit manually
  webhook/                       # Webhook manifests — do not edit manually
  samples/                       # Example CRs for each type
  default/                       # Kustomize overlay used for deployment
  manager/

chart/                           # Helm chart (generated via kubebuilder helm plugin)

test/
  e2e/                           # End-to-end tests using Ginkgo + Kind
  utils/                         # Shared test helpers

grafana/                         # Grafana dashboard definitions

dist/                            # Generated dist/install.yaml (kustomize build output)

mise.toml                         # mise tool version definitions (dev shell)
Makefile                         # All build, test, lint, and deploy targets
PROJECT                          # kubebuilder project metadata
```

---

## Shell Environment

This repository uses [mise](https://mise.jdx.dev/) (`mise.toml`) to manage
tool versions for the dev shell (Go, Helm, kubebuilder, kind, kubectl, etc.).

To activate the tools:

```bash
mise install
mise use
```

---

## API Groups & CRD Types

### `stackit-operator.webhippie.de/v1alpha1` (core)

| Kind | Scope | Description |
|------|-------|-------------|
| `ServiceAccountCredentials` | Namespaced | STACKIT API credentials (project ID, service account key/token) |
| `ClusterServiceAccountCredentials` | Cluster-scoped | Same as above, usable across namespaces |

### `postgres.stackit-operator.webhippie.de/v1alpha1`

| Kind | Scope | Description |
|------|-------|-------------|
| `PostgresDatabase` | Namespaced | A database on a Postgres Flex instance, resolved by `instanceName` |
| `PostgresUser` | Namespaced | A user on a Postgres Flex instance; writes connection details to a Secret |

Both `PostgresDatabase` and `PostgresUser` reference a `StackitCredentialsRef`
(`spec.credentialsRef`) pointing at a `ServiceAccountCredentials` or
`ClusterServiceAccountCredentials` resource, and identify their target
instance via `spec.instanceName` — the operator resolves this to the
underlying STACKIT instance ID at runtime.

---

## Common Types (`api/common`)

| Type | Description |
|------|-------------|
| `StackitCredentialsRef` | Reference to a `ServiceAccountCredentials` (namespaced) or `ClusterServiceAccountCredentials` (cluster-scoped) resource |
| `SecretKeySelector` | Reference to a specific key in a Kubernetes Secret |
| `SecretKeyRefOrVal` | Either an inline value or a `SecretKeySelector` (used for keys/tokens/passwords) |

`StackitCredentialsRef.Kind` defaults to `ServiceAccountCredentials` and is
enum-validated to `ServiceAccountCredentials`/`ClusterServiceAccountCredentials`.

---

## Build & Development Commands

All commands are defined in `Makefile`. Key targets:

| Target | Description |
|--------|-------------|
| `make build` | Build the `bin/manager` binary |
| `make run` | Run the controller locally (requires kubeconfig) |
| `make test` | Run unit/integration tests via envtest |
| `make test-e2e` | Run e2e tests against a temporary Kind cluster |
| `make lint` | Run `golangci-lint` |
| `make lint-fix` | Run linter with auto-fix |
| `make manifests` | Regenerate RBAC/CRD/webhook manifests via `controller-gen` |
| `make generate` | Regenerate DeepCopy methods via `controller-gen` |
| `make schema` | Regenerate JSON Schema files from CRDs for editor integration (yaml-language-server) |
| `make fmt` | Run `go fmt ./...` |
| `make vet` | Run `go vet ./...` |
| `make docker-build` | Build the container image (`ghcr.io/kubehippie/stackit-operator`) |
| `make build-installer` | Generate `dist/install.yaml` via kustomize |

All tooling (`controller-gen`, `kustomize`, `golangci-lint`, `setup-envtest`)
is downloaded locally into `bin/` by the Makefile — no manual installation
needed.

After changing any `_types.go` file always run:

```bash
make generate manifests schema
```

---

## Generated Files — Do Not Edit

The following files are fully generated and must not be edited by hand:

- `api/*/zz_generated.deepcopy.go` — regenerated by `make generate`
- `config/crd/bases/*.yaml` — regenerated by `make manifests`
- `config/rbac/role.yaml` and the per-resource `*_admin/editor/viewer_role.yaml` files — regenerated by `make manifests`
- `config/webhook/manifests.yaml` — regenerated by `make manifests`

---

## Testing

### Unit / integration tests

Uses [Ginkgo](https://onsi.github.io/ginkgo/) +
[Gomega](https://onsi.github.io/gomega/) with `controller-runtime`'s
`envtest` (real API server, no full cluster).

```bash
make test
```

### End-to-end tests

Uses a [Kind](https://kind.sigs.k8s.io/) cluster spun up automatically.

```bash
make test-e2e
```

The Kind cluster is named `stackit-operator-test-e2e` and is torn down after
the test run (`make cleanup-test-e2e`).

---

## CI Workflows

| Workflow | File | Purpose |
|----------|------|---------|
| General | `.github/workflows/general.yml` | Build, test, lint |
| Docker | `.github/workflows/docker.yml` | Build and push container image |
| Release | `.github/workflows/release.yml` | Semantic release and changelog |
| Helm docs | `.github/workflows/helmdocs.yml` | Regenerate chart documentation |
| Automerge | `.github/workflows/automerge.yml` | Renovate automation |

Required status checks on `master`: `lint`, `chart`, `tests`, `docker`.

---

## Deployment

The preferred installation method is the Helm chart:

```bash
helm install stackit-operator \
  oci://ghcr.io/kubehippie/charts/stackit-operator \
  --values values.yaml
```

Raw Kustomize manifests:

```bash
make build-installer   # outputs dist/install.yaml
```

---

## Contribution Conventions

- Use pull requests for all changes; squash or rebase merge only.
- Run `make fmt vet lint` before pushing.
- After any `_types.go` change, run `make generate manifests schema` and
  commit the updated generated files together with the type change. The
  `schema` step keeps the editor JSON Schemas (`config/schema/`) in sync with
  the CRDs, since it is not implied by `manifests`/`generate` alone.
- Keep the Helm chart in sync with CRD/config changes (`make chart`).
- For security issues, contact `thomas@webhippie.de`.
