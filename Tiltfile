load('ext://helm_resource', 'helm_resource')
allow_k8s_contexts('kind-stackit-operator')
update_settings(k8s_upsert_timeout_secs=5*60)

helm_resource(
  'cert-manager',
  'oci://quay.io/jetstack/charts/cert-manager',
  namespace='cert-manager',
  flags=[
    '--create-namespace',
    '--values=test/e2e/testdata/cert-manager.yaml',
    '--timeout=5m',
    '--wait',
    '--hide-notes',
  ],
  labels=['dependencies'],
)

docker_build(
  'ghcr.io/kubehippie/stackit-operator',
  '.',
  dockerfile='Dockerfile.tilt',
  entrypoint='/manager',
  live_update=[
    sync('.', '/workspace'),
    run(
      'go build -o /workspace/manager ./cmd/main.go',
      trigger=['cmd/', 'internal/', 'api/', 'go.mod', 'go.sum']
    ),
  ],
)

k8s_yaml(kustomize('config/default'))

k8s_resource(
  objects=[
    'stackit-operator-system:Namespace',
    'serviceaccountcredentials.stackit-operator.webhippie.de:CustomResourceDefinition',
    'clusterserviceaccountcredentials.stackit-operator.webhippie.de:CustomResourceDefinition',
    'postgresdatabases.postgres.stackit-operator.webhippie.de:CustomResourceDefinition',
    'postgresusers.postgres.stackit-operator.webhippie.de:CustomResourceDefinition',
  ],
  new_name='operator-crds',
  labels=['operator'],
)

k8s_resource(
  objects=[
    'stackit-operator-serviceaccountcredential-admin-role:ClusterRole:default',
    'stackit-operator-serviceaccountcredential-editor-role:ClusterRole:default',
    'stackit-operator-serviceaccountcredential-viewer-role:ClusterRole:default',
    'stackit-operator-clusterserviceaccountcredential-admin-role:ClusterRole:default',
    'stackit-operator-clusterserviceaccountcredential-editor-role:ClusterRole:default',
    'stackit-operator-clusterserviceaccountcredential-viewer-role:ClusterRole:default',
    'stackit-operator-postgres-postgresdatabase-admin-role:ClusterRole:default',
    'stackit-operator-postgres-postgresdatabase-editor-role:ClusterRole:default',
    'stackit-operator-postgres-postgresdatabase-viewer-role:ClusterRole:default',
    'stackit-operator-postgres-postgresuser-admin-role:ClusterRole:default',
    'stackit-operator-postgres-postgresuser-editor-role:ClusterRole:default',
    'stackit-operator-postgres-postgresuser-viewer-role:ClusterRole:default',
  ],
  new_name='operator-clusterroles',
  labels=['operator'],
)

k8s_resource(
  objects=[
    'stackit-operator-selfsigned-issuer:Issuer:stackit-operator-system',
    'stackit-operator-metrics-certs:Certificate:stackit-operator-system',
    'stackit-operator-serving-cert:Certificate:stackit-operator-system',
    'stackit-operator-mutating-webhook-configuration:MutatingWebhookConfiguration:default',
    'stackit-operator-validating-webhook-configuration:ValidatingWebhookConfiguration:default',
  ],
  new_name='operator-certs',
  resource_deps=['cert-manager', 'operator-crds'],
  labels=['operator'],
)

k8s_resource(
  'stackit-operator-controller-manager',
  new_name='operator-manager',
  extra_pod_selectors=[{'control-plane': 'controller-manager'}],
  resource_deps=['operator-certs'],
  labels=['operator'],
)
