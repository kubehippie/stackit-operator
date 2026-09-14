# stackit-operator

[![GitHub Repo](https://img.shields.io/badge/github-repo-yellowgreen)](https://github.com/kubehippie/stackit-operator) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/ef1c1d1b1eff4bc48b7aae87a0d2a141)](https://app.codacy.com/gh/kubehippie/stackit-operator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/kubehippie-stackit-operator)](https://artifacthub.io/packages/helm/kubehippie-stackit-operator/stackit-operator)

> [!WARNING]
> **This project is in early development.** The builtin resources and their APIs
> are **not yet stable** and can introduce **breaking changes** at any time,
> including changes to the CRD schemas, defaulting/validation behavior, and
> reconciliation logic. Pin your version, review the changelog before upgrading,
> and do not use this operator for production workloads until it reaches a
> stable release.

This controller can configure some resources for STACKIT cloud. We don't wanted
to use Terraform or some CLI for this anymore and since there haven't been good
controllers out there we built our own version of it.

## Instructions

Generally you should install this project via [Helm][helm], the other options
are not covered by this document as the chart deployment is the preferred way:

```sh
cat << EOF | helm install stackit-operator oci://ghcr.io/kubehippie/charts/stackit-operator --values -
fullnameOverride: stackit-operator
EOF
```

## Development

We are using [Mise][mise] to install all required tools with fixed versions to
keep everything as far as possible compatible. If you don't want to use
[Mise][mise] it is up to you to install the required tools like Go. Beside that
we are using `make` to define all commands to build this project.

```console
git clone https://github.com/kubehippie/stackit-operator.git
cd stackit-operator

mise trust
mise install

make build
./bin/manager -h
```

To easily work on the operator we suggest to use [Tilt][tilt] for the local
development, this work pretty good in combination with Kind to get features like
hot reloading:

```console
kind create cluster \
    --name stackit-operator

tilt up

kind delete cluster \
    --name stackit-operator
```

## Security

If you find a security issue please contact
[thomas@webhippie.de](mailto:thomas@webhippie.de) first.

## Contributing

Fork -> Patch -> Push -> Pull Request

## Authors

-   [Thomas Boerger](https://github.com/tboerger)

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2026 Thomas Boerger <thomas@webhippie.de>
```

[helm]: https://helm.sh/
[mise]: https://mise.jdx.dev/getting-started.html
[tilt]: https://tilt.dev/
