# Changelog

## [1.0.5](https://github.com/kubehippie/stackit-operator/compare/v1.0.4...v1.0.5) (2026-09-14)

### Bugfixes

* postgres sdk also do not like the region ([9b65e6c](https://github.com/kubehippie/stackit-operator/commit/9b65e6c213f01b798c33827adaed640458e59559))

## [1.0.4](https://github.com/kubehippie/stackit-operator/compare/v1.0.3...v1.0.4) (2026-09-14)

### Bugfixes

* allow dashes within postgres database names ([234ba2b](https://github.com/kubehippie/stackit-operator/commit/234ba2b19580a9fbfa176a4998f3bb834dc48eb6))

## [1.0.3](https://github.com/kubehippie/stackit-operator/compare/v1.0.2...v1.0.3) (2026-09-14)

### Bugfixes

* stupid me missed more rbac rules ([74195f9](https://github.com/kubehippie/stackit-operator/commit/74195f95c2fe4d8a04536a90675baac16287c976))

## [1.0.2](https://github.com/kubehippie/stackit-operator/compare/v1.0.1...v1.0.2) (2026-09-14)

### Bugfixes

* define region only for specific client sdks ([93c1cf8](https://github.com/kubehippie/stackit-operator/commit/93c1cf806c9a642f134918ac82ccc3f9f14a416d))
* resolve typo within chart rbac definitions ([7f38d01](https://github.com/kubehippie/stackit-operator/commit/7f38d01980fc48d05eb295afad52a21234e0e462))

## [1.0.1](https://github.com/kubehippie/stackit-operator/compare/v1.0.0...v1.0.1) (2026-09-14)

### Bugfixes

* add missing role rules for crds ([82923dc](https://github.com/kubehippie/stackit-operator/commit/82923dcaa085bf0921de645158be712dabaeab48))
* **deps:** update gcr.io/distroless/static:nonroot docker digest to e2e927e ([#5](https://github.com/kubehippie/stackit-operator/issues/5)) ([f9949e9](https://github.com/kubehippie/stackit-operator/commit/f9949e9e625d196ed73742979f333bfa85dc08fc))

### Dependencies

* **patch:** update module sigs.k8s.io/controller-runtime to v0.25.1 ([#6](https://github.com/kubehippie/stackit-operator/issues/6)) ([0d1cd1a](https://github.com/kubehippie/stackit-operator/commit/0d1cd1a775e08bc7ff189995ec7c8a3d6363c556))

## 1.0.0 (2026-09-13)

### Features

* configure postgres user secret and integrate rotation ([d7796a7](https://github.com/kubehippie/stackit-operator/commit/d7796a7cbfb505f50747fe408da99f6372f0fef8))
* initial commit ([2ee8a3e](https://github.com/kubehippie/stackit-operator/commit/2ee8a3e874a93c3b0ad938dc7d6b18bdc6f53ebc))

### Bugfixes

* drop wrong copied stuff from keycloak ([cdc0dc4](https://github.com/kubehippie/stackit-operator/commit/cdc0dc44d15b927262e2073a7ef8f1dba040437b))
* use latest until initial release for config ([a078611](https://github.com/kubehippie/stackit-operator/commit/a0786112906fd5af6ea236582de29322f1bafcb6))

### Dependencies

* **mise:** update dependency helm to v4.3.0 ([7777378](https://github.com/kubehippie/stackit-operator/commit/777737878d57b2024581d0d9c45df14389b4f84d))
* **mise:** update dependency kubebuilder to v4.16.0 ([#2](https://github.com/kubehippie/stackit-operator/issues/2)) ([fb64a77](https://github.com/kubehippie/stackit-operator/commit/fb64a77d9550fa4648e87757c5338838cf67b608))
* **mise:** update dependency prek to v0.5.3 ([2f5a23b](https://github.com/kubehippie/stackit-operator/commit/2f5a23bc639388ff540bf61c1a7d40204f9e76ca))
