# Runtime Radar (RR) Helm Chart

This Helm chart deploys the Runtime Radar platform, a comprehensive runtime monitoring and security solution for Kubernetes environments.

## Introduction

Runtime Radar is a runtime security platform that provides:

- **Runtime Monitoring**: Real-time container activity monitoring using eBPF-based Tetragon
- **Policy Enforcement**: Dynamic security policy management and enforcement
- **Event Processing**: Collection and analysis of security events
- **History and Auditing**: Complete audit trail of security events
- **Multi-cluster Management**: Centralized management of multiple Kubernetes clusters
- **API Access**: RESTful API for integration with external systems

## Architecture

The chart deploys the following components:

### Core Components

- **auth-center**: Authentication and authorization service
- **policy-enforcer**: Security policy enforcement engine
- **history-api**: Historical event query API
- **cluster-manager**: Multi-cluster orchestration
- **notifier**: Alert and notification system
- **reverse-proxy**: Ingress and routing layer
- **cs-manager**: Central management service
- **public-api**: External API gateway
- **kube-manager**: Kubernetes management service

### Runtime Components

- **runtime-monitor**: eBPF-based runtime monitoring agent (Tetragon)
- **event-processor**: Real-time event processing pipeline

### Infrastructure Components (Optional)

- **postgresql**: Primary data store
- **redis**: Cache and session store
- **rabbitmq**: Message broker for event streaming
- **clickhouse**: Analytics database for historical events

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- PV provisioner support in the underlying infrastructure (if persistence is enabled)
- Sufficient resources for running all components

## Installation

### Quick Start

Install the chart with default configuration:

```bash
helm install runtime-radar -n runtime-radar ./install/helm \
  --set global.keys.encryption=<YOUR_ENCRYPTION_KEY> \
  --set global.keys.publicAccessTokenSalt=<YOUR_SALT> \
  --set global.ownCsUrl=https://your-domain.com \
  --set global.administrator.username=admin \
  --set global.administrator.password=<YOUR_PASSWORD> \
  --create-namespace
```

**Important**: Replace placeholder values with secure, randomly generated strings. Do not use the example values in production.

### Installing with External Databases

To use external databases instead of deploying them in the cluster:

```bash
helm install runtime-radar -n runtime-radar ./install/helm \
  --set postgresql.deploy=false \
  --set postgresql.externalHost=postgres.example.com \
  --set global.postgresql.auth.existingSecret=my-postgresql-secret \
  --set redis.deploy=false \
  --set redis.externalHost=redis.example.com \
  --set global.redis.auth.existingSecret=my-redis-secret \
  --set rabbitmq.deploy=false \
  --set rabbitmq.externalHost=rabbitmq.example.com \
  --set global.rabbitmq.auth.existingSecret=my-rabbitmq-secret \
  --set clickhouse.deploy=false \
  --set clickhouse.externalHost=clickhouse.example.com \
  --set global.clickhouse.auth.existingSecret=my-clickhouse-secret \
  --set global.keys.encryption=<YOUR_ENCRYPTION_KEY> \
  --set global.keys.publicAccessTokenSalt=<YOUR_SALT> \
  --set global.ownCsUrl=https://your-domain.com \
  --set global.administrator.username=admin \
  --set global.administrator.password=<YOUR_PASSWORD> \
  --create-namespace
```

Each `global.<svc>.auth.existingSecret` must point to an operator-managed Secret carrying AUTH credentials only (`<SVC>_USER`, `<SVC>_PASSWORD`, and `<SVC>_DB` for postgresql/clickhouse). Connection metadata (`<SVC>_ADDR`/`<SVC>_SSL_*`/`<SVC>_TLS_*`) is supplied by the chart-owned `cs-<svc>-config` ConfigMaps and must not be placed in the Secret.

### Installing with Custom Values File

Create a `custom-values.yaml` file:

```yaml
global:
  keys:
    encryption: "<YOUR_ENCRYPTION_KEY>"
    publicAccessTokenSalt: "<YOUR_SALT>"
  ownCsUrl: "https://your-domain.com"
  administrator:
    username: admin
    password: "<YOUR_PASSWORD>"

auth-center:
  replicas: 2

reverse-proxy:
  ingress:
    enabled: true
    class: nginx
    hostname: runtime-radar.your-domain.com

postgresql:
  persistence:
    size: 10Gi
    storageClass: fast-ssd

redis:
  persistence:
    size: 5Gi

rabbitmq:
  persistence:
    size: 10Gi

clickhouse:
  persistence:
    size: 50Gi
```

Install with the custom values:

```bash
helm install runtime-radar -n runtime-radar ./install/helm -f custom-values.yaml --create-namespace
```

## Upgrading

To upgrade an existing installation:

```bash
helm upgrade runtime-radar -n runtime-radar ./install/helm -f custom-values.yaml
```

### Breaking changes

The following breaking changes apply when upgrading from earlier versions:

- **Administrator credentials moved to `global.administrator.*`.** The chart values `auth-api.administrator.username` / `auth-api.administrator.password` / `auth-api.administrator.existingSecret` (the last of which was a never-wired no-op) and the never-wired `auth-center.administrator.*` have been replaced by `global.administrator.username` / `global.administrator.password` / `global.administrator.existingSecret`. There is no auto-migration shim — update your environment overlays and `--set` flags before upgrading, or `helm upgrade` will fail with a `required` validation error.
- **Single `cs-account` Secret replaces `auth-api-account` and `auth-center-account`.** The chart now creates one umbrella-owned Secret named `cs-account` (or uses the operator-supplied Secret named in `global.administrator.existingSecret`). After a `helm upgrade`, the previously created `auth-api-account` and `auth-center-account` Secrets become orphans — they are no longer referenced by any workload and can be deleted safely:

  ```bash
  kubectl delete secret auth-api-account auth-center-account -n runtime-radar --ignore-not-found
  ```
- **Connection metadata moved from Secrets to ConfigMaps.** The chart-owned Secrets `postgresql`/`redis`/`rabbitmq`/`clickhouse` no longer carry `<SVC>_ADDR`/`<SVC>_SSL_*`/`<SVC>_TLS_*` keys — those moved to new ConfigMaps `cs-postgresql-config`/`cs-redis-config`/`cs-rabbitmq-config`/`cs-clickhouse-config`. Operators supplying their own auth Secret via `<svc>.auth.existingSecret` or the new `global.<svc>.auth.existingSecret` must scope it to AUTH credentials only (`<SVC>_USER`, `<SVC>_PASSWORD`, `<SVC>_DB` where applicable). Including ADDR/SSL keys in the Secret no longer has any effect; consumers read those from the ConfigMaps.
- **Cross-subchart secret-name propagation requires `global.<svc>.auth.existingSecret`.** Setting only the top-level `postgresql.auth.existingSecret` propagates to sub-chart but NOT to consumer deployments (auth-api, history-api, etc.). To make consumer pods read from your external Secret, use `global.postgresql.auth.existingSecret` (and the corresponding global knobs for redis/rabbitmq/clickhouse). For external services where subchart isn't deployed (`<svc>.deploy=false`), only the global knob is needed.
- **`global.keys.existingSecret` now honored by every consumer.** `cluster-manager` and `public-api` previously hardcoded `cs-keys` for `PUBLIC_ACCESS_TOKEN_SALT_KEY` / `ACCESS_TOKEN_SALT` and silently ignored `global.keys.existingSecret`. They now resolve the keys-secret name through the `common.cs.keys.secretName` helper. Operators using a custom `global.keys.existingSecret` must ensure the Secret contains the `publicAccessTokenSalt` key (in addition to `encryption` and `token`).

## Uninstalling

To uninstall/delete the `runtime-radar` deployment:

```bash
helm uninstall runtime-radar -n runtime-radar
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

**Note**: Persistent Volume Claims are not deleted automatically. To delete them:

```bash
kubectl delete pvc -n runtime-radar -l app.kubernetes.io/instance=runtime-radar
```

## Configuration

### Security Keys

The following keys must be set for the system to function:

#### `global.keys.encryption`

Encryption key for secrets stored in database. Must be a 64-character hexadecimal string (32 bytes)

Can be generated with command:
```sh
openssl rand -hex 32
```

#### `global.keys.token`

Encryption key for authentication tokens. Must be a 64-character hexadecimal string (32 bytes)

Can be generated with command:
```sh
openssl rand -hex 32
```

#### `global.keys.publicAccessTokenSalt`

Salt for public API tokens. Must be a 128-character hexadecimal string (64 bytes)

Can be generated with command:
```sh
openssl rand -hex 64
```

#### Auto-generated Keys

You can let Helm generate these keys automatically by setting the value to `INIT-DO-NOT-USE`:

```sh
helm install runtime-radar -n runtime-radar ./install/helm \
  --set global.keys.encryption=INIT-DO-NOT-USE \
  --set global.keys.publicAccessTokenSalt=INIT-DO-NOT-USE \
  --set global.ownCsUrl=https://your-domain.com \
  --set global.administrator.username=admin \
  --set global.administrator.password=<YOUR_PASSWORD> \
  --create-namespace
```

**Warning**: This approach should be used with extreme caution:
- Keys are generated during installation and stored in Kubernetes secrets
- On subsequent `helm upgrade` operations, if you continue to use `INIT-DO-NOT-USE`, new keys may be regenerated
- Regenerated keys will make existing encrypted data **unrecoverable**, resulting in **data loss**
- This feature is intended for development and testing environments only
- For production environments, always generate and manage keys securely outside of Helm

**Important**: These keys should be securely generated and stored. Changing them after initial deployment may result in data loss.

### TLS Configuration

The chart supports TLS configuration in multiple ways:

#### Auto-generated Self-signed Certificates

```yaml
tls:
  autoGenerated: true
```

#### Existing Secret

```yaml
tls:
  autoGenerated: false
global:
  tls:
    existingSecret: "runtime-radar-tls-secret"
```

The secret should contain:
- `tls.crt`: TLS certificate
- `tls.key`: TLS private key
- `ca.crt`: CA certificate (optional)

#### Inline Certificates

```yaml
tls:
  autoGenerated: false
  cert: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  certKey: |
    -----BEGIN PRIVATE KEY-----
    ...
    -----END PRIVATE KEY-----
  certCA: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
```

### Ingress Configuration

To expose the Runtime Radar web interface:

```yaml
reverse-proxy:
  ingress:
    enabled: true
    class: nginx
    hostname: runtime-radar.your-domain.com
    tls:
      autoGenerated: false
      existingSecret: "cs-ingress-tls"
```

### Multi-cluster Setup

#### Primary Cluster

```yaml
global:
  ownCsUrl: "https://runtime-radar-primary.your-domain.com"
  isChildCluster: false
```

#### Child Cluster

```yaml
global:
  ownCsUrl: "https://runtime-radar-child.your-domain.com"
  centralCsUrl: "https://runtime-radar-primary.your-domain.com"
  isChildCluster: true
```

### Resource Requirements

Default resource limits can be adjusted based on your workload:

```yaml
runtime-monitor:
  resources:
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 200m
      memory: 256Mi

event-processor:
  resources:
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 200m
      memory: 256Mi
```

### High Availability

For production deployments, increase replica counts:

```yaml
auth-center:
  replicas: 3

policy-enforcer:
  replicas: 3

history-api:
  replicas: 3

event-processor:
  replicas: 3

public-api:
  replicas: 3
```

### Monitoring (Metrics, Prometheus & Grafana)

Runtime Radar ships an optional, self-contained observability stack. It has three layers that work together:

1. **Metrics exposure** — make components expose Prometheus-format `/metrics`.
2. **Prometheus** — scrape and store those metrics.
3. **Grafana** — visualize them with the bundled dashboards.

Each layer can be enabled independently and can also point at infrastructure you already run.

#### Quick start: full in-cluster stack

To deploy the complete bundled stack (metrics + Prometheus + Grafana), set:

```yaml
# 1. Expose application + infrastructure metrics
metrics:
  enabled: true
postgresql:
  metrics:
    enabled: true
rabbitmq:
  metrics:
    enabled: true
clickhouse:
  metrics:
    enabled: true

# 2. Deploy Prometheus to scrape them
prometheus:
  deploy: true

# 3. Deploy Grafana to visualize them
grafana:
  deploy: true
  auth:
    username: admin
    password: "<YOUR_GRAFANA_PASSWORD>"
```

Once running, Grafana is exposed through the reverse-proxy at `https://<your-ownCsUrl>/grafana` (see [Accessing Grafana](#accessing-grafana)).

> **How the wiring works:** the umbrella chart auto-generates a `cs-metrics` ConfigMap (Prometheus scrape config) **only when Prometheus is deployed and at least one metrics flag is enabled**, and a `grafana-datasources` Secret that points Grafana's `Prometheus` datasource at the in-cluster (or external) Prometheus. You do not need to write scrape configs or datasources by hand.

#### Enabling metrics

Metrics are **off by default**. Turning a flag on only makes a component *expose* metrics — something still has to scrape them (an in-cluster or external Prometheus).

| Flag                         | What it does                                                                                                              |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `metrics.enabled`            | Runtime Radar application services expose `/metrics`. Also wires the web UI's "open in Grafana" links (`GRAFANA_URL`).   |
| `global.metrics.enabled`     | Same as above; useful when sharing one value across child clusters.                                                       |
| `postgresql.metrics.enabled` | Starts a PostgreSQL Prometheus exporter (scraped on port `9187`).                                                         |
| `rabbitmq.metrics.enabled`   | Exposes RabbitMQ metrics (scraped on port `9419`).                                                                        |
| `clickhouse.metrics.enabled` | Exposes ClickHouse metrics (scraped on port `8001`).                                                                      |

For an **external** datastore (`<svc>.deploy=false`), set `<svc>.metrics.externalHost` so Prometheus scrapes the right target, e.g.:

```yaml
postgresql:
  deploy: false
  externalHost: postgres.example.com
  metrics:
    enabled: true
    externalHost: postgres-exporter.example.com:9187
```

The generated scrape config also collects Tetragon runtime metrics (`runtime-monitor`, port `2112`) and reverse-proxy/Caddy metrics (port `9090`) automatically.

#### Prometheus

```yaml
prometheus:
  deploy: true
  replicaCount: 1
  persistence:
    enabled: true
    size: 20Gi
    storageClass: fast-ssd
```

- The scrape configuration is read from the ConfigMap named in `prometheus.scrapeConfigmap` (default `cs-metrics`), which the chart renders for you from the metrics flags above. Enabling Prometheus without any metrics flag produces an empty scrape config and nothing to graph.
- Adjust retention/storage with `prometheus.persistence.*`.

**Use an existing Prometheus** instead of deploying one — leave `prometheus.deploy=false` and point the stack at it. Grafana's datasource is configured from this value:

```yaml
prometheus:
  deploy: false
  externalHost: https://prometheus.monitoring.svc.cluster.local:9090
```

#### Grafana

```yaml
grafana:
  deploy: true
  auth:
    username: admin
    password: "<YOUR_GRAFANA_PASSWORD>"
  persistence:
    enabled: true
    size: 5Gi
```

- **Dashboards** — the chart packs every dashboard under `install/helm/dashboards/*.json` (runtime, Tetragon runtime, gRPC, Go app, cluster status, public-api, reverse-proxy) into the `cs-dashboards` ConfigMap and loads them automatically through the dashboard provider (`grafana.dashboardsProvider.enabled`, on by default). Add your own by listing extra ConfigMaps in `grafana.dashboardsConfigMaps`.
- **Datasource** — the `Prometheus` datasource is wired automatically to the deployed (or external) Prometheus; no manual setup needed.

##### Accessing Grafana

By default `grafana.subPath: grafana` serves Grafana behind the reverse-proxy, so it is reachable at:

```
https://<global.ownCsUrl>/grafana
```

The Runtime Radar web UI links here directly when `metrics.enabled` is set.

**Use an existing Grafana** instead — leave `grafana.deploy=false` and set `grafana.externalHost`. The UI's Grafana links then redirect to that host:

```yaml
grafana:
  deploy: false
  externalHost: https://grafana.example.com
```

### Node Affinity

Deploy components to specific nodes:

```yaml
runtime-monitor:
  nodeSelector:
    workload: monitoring

postgresql:
  nodeSelector:
    workload: database

clickhouse:
  nodeSelector:
    workload: analytics
```

## Parameters

### Global parameters

| Name                                    | Description                                                                                                                                                                                                                                                                                                                                                                                                     | Value         |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| `global.imageRegistry`                  | Global Docker image registry                                                                                                                                                                                                                                                                                                                                                                                    | `""`          |
| `global.imageTag`                       | Global Docker image tag to use for RR components                                                                                                                                                                                                                                                                                                                                                                | `""`          |
| `global.logLevel`                       | Logging level for components                                                                                                                                                                                                                                                                                                                                                                                    | `INFO`        |
| `global.tls.enabled`                    | Enable TLS for RR components                                                                                                                                                                                                                                                                                                                                                                                    | `true`        |
| `global.tls.existingSecret`             | Name of the existing secret with TLS certificates                                                                                                                                                                                                                                                                                                                                                               | `""`          |
| `global.auth.enabled`                   | Enable authentication for RR components                                                                                                                                                                                                                                                                                                                                                                         | `true`        |
| `global.keys.existingSecret`            | Existing secret name with keys `encryption`, `token`, and `publicAccessTokenSalt`                                                                                                                                                                                                                                                                                                                               | `""`          |
| `global.keys.encryption`                | Encryption key for secrets stored in database. Must be a 64-character hexadecimal string (32 bytes)                                                                                                                                                                                                                                                                                                             | `""`          |
| `global.keys.token`                     | Encryption key for authentication tokens. Must be a 64-character hexadecimal string (32 bytes)                                                                                                                                                                                                                                                                                                                  | `""`          |
| `global.keys.publicAccessTokenSalt`     | Salt for public API tokens. Must be a 128-character hexadecimal string (64 bytes)                                                                                                                                                                                                                                                                                                                               | `""`          |
| `global.administrator.existingSecret`   | Name of an existing secret with administrator credentials (must contain keys `username` and `password`). When empty, the chart creates a Secret named `cs-account` from `username` and `password` below.                                                                                                                                                                                                        | `""`          |
| `global.administrator.username`         | Administrator username. Ignored when `existingSecret` is set.                                                                                                                                                                                                                                                                                                                                                   | `""`          |
| `global.administrator.password`         | Administrator password. Ignored when `existingSecret` is set.                                                                                                                                                                                                                                                                                                                                                   | `""`          |
| `global.postgresql.auth.existingSecret` | Name of an existing secret with PostgreSQL auth credentials (must contain `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`). When empty, the chart creates a Secret named `postgresql`. NOTE: when `postgresql.deploy=true` (default), you must also set `postgresql.auth.existingSecret` to the same value so the sub-chart reads from it. For external PostgreSQL, prefer `postgresql.deploy=false`.       | `""`          |
| `global.postgresql.tls.enabled`         | Enable TLS traffic support (overrides `tls.enabled`)                                                                                                                                                                                                                                                                                                                                                            | `true`        |
| `global.postgresql.tls.verify`          | Verify TLS connection to the service (overrides `tls.verify`)                                                                                                                                                                                                                                                                                                                                                   | `true`        |
| `global.postgresql.tls.existingSecret`  | Name of an existing secret that contains the certificates (overrides `tls.existingSecret`)                                                                                                                                                                                                                                                                                                                      | `""`          |
| `global.redis.auth.existingSecret`      | Name of an existing secret with Redis auth credentials (must contain `REDIS_USER`, `REDIS_PASSWORD`). When empty, the chart creates a Secret named `redis`. NOTE: when `redis.deploy=true` (default), you must also set `redis.auth.existingSecret` to the same value so the sub-chart reads from it.                                                                                                           | `""`          |
| `global.redis.tls.enabled`              | Enable TLS traffic support (overrides `tls.enabled`)                                                                                                                                                                                                                                                                                                                                                            | `true`        |
| `global.redis.tls.verify`               | Verify TLS connection to the service (overrides `tls.verify`)                                                                                                                                                                                                                                                                                                                                                   | `true`        |
| `global.redis.tls.existingSecret`       | Name of an existing secret that contains the certificates (overrides `tls.existingSecret`)                                                                                                                                                                                                                                                                                                                      | `""`          |
| `global.rabbitmq.auth.existingSecret`   | Name of an existing secret with RabbitMQ auth credentials (must contain `RABBIT_USER`, `RABBIT_PASSWORD`). When empty, the chart creates a Secret named `rabbitmq`. NOTE: when `rabbitmq.deploy=true` (default), you must also set `rabbitmq.auth.existingSecret` to the same value so the sub-chart reads from it.                                                                                             | `""`          |
| `global.clickhouse.auth.existingSecret` | Name of an existing secret with ClickHouse auth credentials (must contain `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB`). When empty, the chart creates a Secret named `clickhouse`. NOTE: when `clickhouse.deploy=true` (default), you must also set `clickhouse.auth.existingSecret` to the same value so the sub-chart reads from it. For external ClickHouse, prefer `clickhouse.deploy=false`. | `""`          |
| `global.clickhouse.tls.enabled`         | Enable TLS traffic support (overrides `tls.enabled`)                                                                                                                                                                                                                                                                                                                                                            | `true`        |
| `global.clickhouse.tls.verify`          | Verify TLS connection to the service (overrides `tls.verify`)                                                                                                                                                                                                                                                                                                                                                   | `true`        |
| `global.clickhouse.tls.existingSecret`  | Name of an existing secret that contains the certificates (overrides `tls.existingSecret`)                                                                                                                                                                                                                                                                                                                      | `""`          |
| `global.grafana.tls.enabled`            | Enable TLS traffic support (overrides `tls.enabled`)                                                                                                                                                                                                                                                                                                                                                            | `true`        |
| `global.grafana.tls.verify`             | Verify TLS connection to the service (overrides `tls.verify`)                                                                                                                                                                                                                                                                                                                                                   | `true`        |
| `global.grafana.tls.existingSecret`     | Name of an existing secret that contains the certificates (overrides `tls.existingSecret`)                                                                                                                                                                                                                                                                                                                      | `""`          |
| `global.imagePullSecrets`               | Names of the secrets of the global container registry as an array                                                                                                                                                                                                                                                                                                                                               | `["regcred"]` |
| `global.ownCsUrl`                       | URL of primary installation                                                                                                                                                                                                                                                                                                                                                                                     | `""`          |
| `global.centralCsUrl`                   | URL of primary installation                                                                                                                                                                                                                                                                                                                                                                                     | `""`          |
| `global.isChildCluster`                 | Is this a child cluster                                                                                                                                                                                                                                                                                                                                                                                         | `false`       |

### Common RR parameters

| Name                       | Description                                                                | Value           |
| -------------------------- | -------------------------------------------------------------------------- | --------------- |
| `fullnameOverride`         | String to fully override common.fullname                                   | `runtime-radar` |
| `imagePullSecret.name`     | Name of the secret with container registry credentials                     | `regcred`       |
| `imagePullSecret.username` | Container registry username                                                | `""`            |
| `imagePullSecret.password` | Container registry password                                                | `""`            |
| `serviceAccount.create`    | Create a service account                                                   | `true`          |
| `serviceAccount.name`      | Service account name                                                       | `runtime-radar` |
| `tls.autoGenerated`        | Generate automatically self-signed TLS certificates if nothing is provided | `true`          |
| `tls.verify`               | Verify connection to external cluster                                      | `false`         |
| `tls.cert`                 | TLS certificate                                                            | `""`            |
| `tls.certKey`              | TLS certificate key                                                        | `""`            |
| `tls.certCA`               | TLS certificate CA                                                         | `""`            |
| `metrics.enabled`          | Enable CS metrics                                                          | `false`         |

### Auth-center component parameters

| Name                       | Description                                                | Value |
| -------------------------- | ---------------------------------------------------------- | ----- |
| `auth-center.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `auth-center.replicas`     | Number of replicas for the auth-center component           | `2`   |

### Policy-enforcer component parameters

| Name                           | Description                                                | Value |
| ------------------------------ | ---------------------------------------------------------- | ----- |
| `policy-enforcer.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `policy-enforcer.replicas`     | Number of replicas for the policy-enforcer component       | `2`   |

### History-api component parameters

| Name                            | Description                                                | Value            |
| ------------------------------- | ---------------------------------------------------------- | ---------------- |
| `history-api.nodeSelector`      | Template to specify the labels of nodes for pod assignment | `{}`             |
| `history-api.replicas`          | Number of replicas for the history-api component           | `2`              |
| `history-api.retentionInterval` | Interval to retain history data for                        | `8760h`          |
| `history-api.rabbitmq.queue`    | RabbitMQ queue name                                        | `history_events` |

### Container-registry-integrator component parameters

| Name                           | Description                                                | Value |
| ------------------------------ | ---------------------------------------------------------- | ----- |
| `cluster-manager.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `cluster-manager.replicas`     | Number of replicas for the cluster-manager component       | `2`   |

### Notifier component parameters

| Name                    | Description                                                | Value |
| ----------------------- | ---------------------------------------------------------- | ----- |
| `notifier.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `notifier.replicas`     | Number of replicas for the notifier component              | `2`   |

### Reverse-proxy component parameters

| Name                                       | Description                                                                | Value       |
| ------------------------------------------ | -------------------------------------------------------------------------- | ----------- |
| `reverse-proxy.nodeSelector`               | Template to specify the labels of nodes for pod assignment                 | `{}`        |
| `reverse-proxy.replicas`                   | Number of replicas for the reverse-proxy component                         | `2`         |
| `reverse-proxy.ingress.enabled`            | Enable ingress for RR                                                      | `false`     |
| `reverse-proxy.ingress.class`              | Ingress class                                                              | `""`        |
| `reverse-proxy.ingress.hostname`           | Hostname of ingress                                                        | `""`        |
| `reverse-proxy.ingress.tls.autoGenerated`  | Generate automatically self-signed TLS certificates if nothing is provided | `true`      |
| `reverse-proxy.ingress.tls.existingSecret` | Name of an existing secret that contains the certificates                  | `""`        |
| `reverse-proxy.ingress.tls.cert`           | Certificate value                                                          | `""`        |
| `reverse-proxy.ingress.tls.certKey`        | Certificate key value                                                      | `""`        |
| `reverse-proxy.ingress.tls.certCA`         | CA Certificate value                                                       | `""`        |
| `reverse-proxy.service.type`               | Type of reverse-proxy service                                              | `ClusterIP` |
| `reverse-proxy.service.nodePorts`          | Node ports which should be exposed outside                                 | `{}`        |

### CS-manager component parameters

| Name                           | Description                                                | Value |
| ------------------------------ | ---------------------------------------------------------- | ----- |
| `cs-manager.nodeSelector`      | Template to specify the labels of nodes for pod assignment | `{}`  |
| `cs-manager.replicas`          | Number of replicas for the cs-manager component            | `1`   |
| `cs-manager.registrationToken` | Token for cluster registration                             | `""`  |

### Kube-manager component parameters

| Name                        | Description                                                | Value |
|-----------------------------| ---------------------------------------------------------- | ----- |
| `kube-manager.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `kube-manager.replicas`     | Number of replicas for the cs-manager component            | `1`   |

### Runtime-monitor component parameters

| Name                                                | Description                                                     | Value                                                                                                                                 |
| --------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `runtime-monitor.nodeSelector`                      | Template to specify the labels of nodes for pod assignment      | `{}`                                                                                                                                  |
| `runtime-monitor.configUpdateInterval`              | Interval to update the stored configuration                     | `30s`                                                                                                                                 |
| `runtime-monitor.dnsPolicy`                         | Configuration of the DNS policy for runtime monitoring          | `ClusterFirstWithHostNet`                                                                                                             |
| `runtime-monitor.containerPorts.http`               | Port that HTTP server should be listening on                    | `9000`                                                                                                                                |
| `runtime-monitor.containerPorts.grpc`               | Port that GRPC server should be listening on                    | `8000`                                                                                                                                |
| `runtime-monitor.containerPorts.gops`               | Port that gops agent should be listening on                     | `7000`                                                                                                                                |
| `runtime-monitor.tetragon.enableProcessCred`        | Enable visibility of capabilities in the exec and kprobe events | `true`                                                                                                                                |
| `runtime-monitor.tetragon.enableProcessNs`          | Enable visibility of namespaces in the exec and kprobe events   | `true`                                                                                                                                |
| `runtime-monitor.tetragon.enableMsgHandlingLatency` | Enable latency monitoring in message handling                   | `true`                                                                                                                                |
| `runtime-monitor.tetragon.exportAllowList`          | Allowlist for JSON export                                       | `{"pod_regex":["deathstar"],"event_set":["PROCESS_EXEC", "PROCESS_EXIT", "PROCESS_KPROBE", "PROCESS_UPROBE", "PROCESS_TRACEPOINT"]}
` |
| `runtime-monitor.tetragon.grpc.address`             | Set address of Tetragon grpc connection in host:port format     | `localhost:54321`                                                                                                                     |
| `runtime-monitor.tetragon.prometheus.enabled`       | Whether to enable exposing Tetragon metrics.                    | `true`                                                                                                                                |
| `runtime-monitor.tetragon.prometheus.port`          | The port at which to expose metrics.                            | `2112`                                                                                                                                |
| `runtime-monitor.tetragon.resources`                | Resource configuration for tetragon container                   | `{}`                                                                                                                                  |
| `runtime-monitor.rabbitmq.queue`                    | RabbitMQ queue name                                             | `runtime_events`                                                                                                                      |
| `runtime-monitor.resources`                         | Resource configuration for runtime-monitor container            | `{}`                                                                                                                                  |

### Event-processor component parameters

| Name                                          | Description                                                | Value            |
| --------------------------------------------- | ---------------------------------------------------------- | ---------------- |
| `event-processor.nodeSelector`                | Template to specify the labels of nodes for pod assignment | `{}`             |
| `event-processor.replicas`                    | Number of replicas for the component                       | `2`              |
| `event-processor.configUpdateInterval`        | Interval to update the stored configuration                | `30s`            |
| `event-processor.rabbitmq.runtimeEventsQueue` | RabbitMQ runtime events queue name                         | `runtime_events` |
| `event-processor.rabbitmq.historyEventsQueue` | RabbitMQ history events queue name                         | `history_events` |
| `event-processor.resources`                   | Resource configuration for event-processor container       | `{}`             |

### Public-api component parameters

| Name                      | Description                                                | Value |
| ------------------------- | ---------------------------------------------------------- | ----- |
| `public-api.nodeSelector` | Template to specify the labels of nodes for pod assignment | `{}`  |
| `public-api.replicas`     | Number of replicas for the public-api component            | `2`   |

### Postgresql installation configuration

| Name                                        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Value               |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------- |
| `postgresql.deploy`                         | Deploy component                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `true`              |
| `postgresql.externalHost`                   | External host with PostgreSQL. Requires setting `postgresql.deploy` to `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `""`                |
| `postgresql.fullnameOverride`               | String to fully override common.names.fullname template                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | `postgresql`        |
| `postgresql.tls.autoGenerated`              | Generate automatically self-signed TLS certificates if nothing is provided                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `true`              |
| `postgresql.tls.cert`                       | Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `""`                |
| `postgresql.tls.certKey`                    | Certificate key value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `""`                |
| `postgresql.tls.certCA`                     | CA Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `""`                |
| `postgresql.auth.existingSecret`            | Name of an existing secret with PostgreSQL credentials, read by the sub-chart when `postgresql.deploy=true`. The `auth.postgresPassword`, `auth.password`, and `auth.replicationPassword` values will be ignored and taken from this secret. The secret must carry AUTH credentials only — `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` (connection metadata `POSTGRES_ADDR`/`POSTGRES_SSL_*` lives in the always-created `cs-postgresql-config` ConfigMap and must not be placed in this Secret). NOTE: To make the same secret visible to all consumer services (auth-api, history-api, etc.), set `global.postgresql.auth.existingSecret` to the same value. The secret might also contain the `ldap-password` key if LDAP is enabled. | `postgresql`        |
| `postgresql.auth.username`                  | Name of custom user to be created                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `runtime-radar`     |
| `postgresql.auth.password`                  | Password of custom user to be created. Ignored if `auth.existingSecret` is set.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `""`                |
| `postgresql.auth.database`                  | Name of custom database to be created                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `runtime-radar`     |
| `postgresql.auth.existingSecretPasswordKey` | Name of the key in the existing secret with PostgreSQL credentials. Only used if `auth.existingSecret` is set.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `POSTGRES_PASSWORD` |
| `postgresql.nodeSelector`                   | Labels of nodes for primary PostgreSQL pod assignment                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `{}`                |
| `postgresql.resources`                      | Resource configuration for PostgreSQL container                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `{}`                |
| `postgresql.persistence.enabled`            | Enable data persistence for primary PostgreSQL using PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | `true`              |
| `postgresql.persistence.storageClass`       | Persistent volume storage class for primary PostgreSQL                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `""`                |
| `postgresql.persistence.size`               | Persistent volume size for PostgreSQL                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `1Gi`               |
| `postgresql.persistence.existingClaim`      | Name of an existing PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | `""`                |
| `postgresql.persistence.selector`           | Template to specify an existing persistent volume                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `{}`                |
| `postgresql.metrics.enabled`                | Start a prometheus exporter                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `false`             |
| `postgresql.metrics.externalHost`           | PostgreSQL metrics external host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `""`                |

### Redis installation configuration

| Name                                   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Value            |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `redis.deploy`                         | Deploy component                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `true`           |
| `redis.externalHost`                   | External host with Redis. Requires setting `redis.deploy` to `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `""`             |
| `redis.fullnameOverride`               | String to fully override common.names.fullname                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `redis`          |
| `redis.tls.autoGenerated`              | Generate automatically self-signed TLS certificates if nothing is provided                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `true`           |
| `redis.tls.cert`                       | Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `""`             |
| `redis.tls.certKey`                    | Certificate key value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `""`             |
| `redis.tls.certCA`                     | CA Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | `""`             |
| `redis.auth.existingSecret`            | Name of an existing secret with Redis credentials, read by the sub-chart when `redis.deploy=true`. When set, the `auth.password` parameter is ignored. The secret must carry AUTH credentials only — `REDIS_USER`, `REDIS_PASSWORD` (connection metadata `REDIS_ADDR`/`REDIS_TLS_*` lives in the always-created `cs-redis-config` ConfigMap and must not be placed in this Secret). NOTE: To make the same secret visible to all consumer services (auth-api, history-api, etc.), set `global.redis.auth.existingSecret` to the same value. | `redis`          |
| `redis.auth.username`                  | Redis username                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `runtime-radar`  |
| `redis.auth.password`                  | Redis password                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `""`             |
| `redis.auth.existingSecretPasswordKey` | Password key to retrieve from the existing secret                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `REDIS_PASSWORD` |
| `redis.replicaCount`                   | Number of Redis master instances to deploy (experimental, requires additional configuring)                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `1`              |
| `redis.nodeSelector`                   | Labels of nodes for Redis master pod assignment                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `{}`             |
| `redis.resources`                      | Resource configuration for Redis container                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `{}`             |
| `redis.persistence.enabled`            | Enable persistence for Redis master nodes using PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | `false`          |
| `redis.persistence.storageClass`       | Persistent volume storage class                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `""`             |
| `redis.persistence.size`               | Persistent volume size                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `1Gi`            |
| `redis.persistence.existingClaim`      | Use an existing PVC created manually                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | `""`             |
| `redis.persistence.selector`           | Template to specify additional labels for PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | `{}`             |

### RabbitMQ installation configuration

| Name                                      | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Value             |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `rabbitmq.deploy`                         | Deploy component                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | `true`            |
| `rabbitmq.externalHost`                   | External host with RabbitMQ                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | `""`              |
| `rabbitmq.fullnameOverride`               | String to fully override rabbitmq.fullname template                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `rabbitmq`        |
| `rabbitmq.auth.username`                  | RabbitMQ application username                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `runtime-radar`   |
| `rabbitmq.auth.password`                  | RabbitMQ application password                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `""`              |
| `rabbitmq.auth.existingSecret`            | Name of an existing secret with RabbitMQ credentials, read by the sub-chart when `rabbitmq.deploy=true`. When set, the `auth.password` parameter is ignored. The secret must carry AUTH credentials only — `RABBIT_USER`, `RABBIT_PASSWORD` (connection metadata `RABBIT_ADDR` lives in the always-created `cs-rabbitmq-config` ConfigMap and must not be placed in this Secret). NOTE: To make the same secret visible to all consumer services (event-processor, history-api, etc.), set `global.rabbitmq.auth.existingSecret` to the same value. | `rabbitmq`        |
| `rabbitmq.auth.existingSecretPasswordKey` | Password key to be retrieved from existing secret                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `RABBIT_PASSWORD` |
| `rabbitmq.nodeSelector`                   | Template to specify the labels of nodes for pod assignment                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | `{}`              |
| `rabbitmq.resources`                      | Resource configuration for RabbitMQ container                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `{}`              |
| `rabbitmq.persistence.enabled`            | Enable RabbitMQ data persistence using PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | `true`            |
| `rabbitmq.persistence.storageClass`       | Persistent volume storage class for RabbitMQ                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | `""`              |
| `rabbitmq.persistence.size`               | Persistent volume size for RabbitMQ                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `1Gi`             |
| `rabbitmq.persistence.existingClaim`      | Name of an existing PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `""`              |
| `rabbitmq.persistence.selector`           | Template to specify an existing persistent volume                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `{}`              |
| `rabbitmq.metrics.enabled`                | Enable exposing RabbitMQ metrics to be gathered by Prometheus                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `false`           |
| `rabbitmq.metrics.externalHost`           | RabbitMQ metrics external host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | `""`              |

### Clickhouse installation configuration

| Name                                        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Value                 |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| `clickhouse.deploy`                         | Deploy component                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `true`                |
| `clickhouse.externalHost`                   | External host with ClickHouse. Requires setting `clickhouse.deploy` to `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `""`                  |
| `clickhouse.fullnameOverride`               | String to fully override common.names.fullname                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `clickhouse`          |
| `clickhouse.nodeSelector`                   | Labels of nodes for ClickHouse pod assignment                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | `{}`                  |
| `clickhouse.replicaCount`                   | Number of ClickHouse replicas to deploy per shard                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `1`                   |
| `clickhouse.resources`                      | Resource configuration for Clickhouse container                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `{}`                  |
| `clickhouse.persistence.enabled`            | Enable persistence using PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | `true`                |
| `clickhouse.persistence.storageClass`       | Persistent volume storage class                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `""`                  |
| `clickhouse.persistence.size`               | Data volume size                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `5Gi`                 |
| `clickhouse.persistence.existingClaim`      | Name of an existing PVC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | `""`                  |
| `clickhouse.persistence.selector`           | Template to specify an existing persistent volume                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `{}`                  |
| `clickhouse.tls.autoGenerated`              | Generate automatically self-signed TLS certificates if nothing is provided                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | `true`                |
| `clickhouse.tls.cert`                       | Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `""`                  |
| `clickhouse.tls.certKey`                    | Certificate key value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | `""`                  |
| `clickhouse.tls.certCA`                     | CA Certificate value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | `""`                  |
| `clickhouse.auth.username`                  | ClickHouse administrator name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | `runtime-radar`       |
| `clickhouse.auth.password`                  | ClickHouse administartor password                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `""`                  |
| `clickhouse.auth.existingSecret`            | Name of an existing secret with ClickHouse credentials, read by the sub-chart when `clickhouse.deploy=true`. When set, the `auth.password` parameter is ignored. The secret must carry AUTH credentials only — `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB` (connection metadata `CLICKHOUSE_ADDR`/`CLICKHOUSE_SSL_*` lives in the always-created `cs-clickhouse-config` ConfigMap and must not be placed in this Secret). NOTE: To make the same secret visible to all consumer services (history-api, event-processor, etc.), set `global.clickhouse.auth.existingSecret` to the same value. | `clickhouse`          |
| `clickhouse.auth.existingSecretPasswordKey` | Name of the key stored in the existing secret                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | `CLICKHOUSE_PASSWORD` |
| `clickhouse.auth.database`                  | Name of the ClickHouse database                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | `runtime-radar`       |
| `clickhouse.metrics.enabled`                | Enable the export of Prometheus metrics                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | `false`               |
| `clickhouse.metrics.externalHost`           | ClickHouse metrics external host                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `""`                  |

### Grafana installation configuration

| Name                                 | Description                                                                  | Value                 |
| ------------------------------------ | ---------------------------------------------------------------------------- | --------------------- |
| `grafana.deploy`                     | Deploy component                                                             | `false`               |
| `grafana.externalHost`               | External host with Grafana. Requires setting `grafana.deploy` to `false`.    | `""`                  |
| `grafana.fullnameOverride`           | String to fully override common.names.fullname                               | `grafana`             |
| `grafana.nodeSelector`               | Node labels for pod assignment                                               | `{}`                  |
| `grafana.replicaCount`               | Number of Grafana nodes                                                      | `1`                   |
| `grafana.resources`                  | Resource configuration for Clickhouse container                              | `{}`                  |
| `grafana.tls.autoGenerated`          | Generate automatically self-signed TLS certificates if nothing is provided   | `true`                |
| `grafana.tls.cert`                   | Certificate value                                                            | `""`                  |
| `grafana.tls.certKey`                | Certificate key value                                                        | `""`                  |
| `grafana.tls.certCA`                 | CA Certificate value                                                         | `""`                  |
| `grafana.auth.username`              | Grafana administrator name                                                   | `runtime-radar`       |
| `grafana.auth.password`              | Grafana administartor password                                               | `""`                  |
| `grafana.datasourcesSecretName`      | The name of an externally-managed secret containing custom datasource files. | `grafana-datasources` |
| `grafana.dashboardsProvider.enabled` | Enable the use of a Grafana dashboard provider                               | `true`                |
| `grafana.dashboardsConfigMaps`       | Array with the names of a series of ConfigMaps containing dashboards files   | `{}`                  |
| `grafana.subPath`                    | Use sub path for grafana for exposing it via reverse proxy                   | `grafana`             |

### Prometheus installation configuration

| Name                                   | Description                                                                     | Value        |
| -------------------------------------- | ------------------------------------------------------------------------------- | ------------ |
| `prometheus.deploy`                    | Deploy component                                                                | `false`      |
| `prometheus.externalHost`              | External host with Prometheus. Requires setting `prometheus.deploy` to `false`. | `""`         |
| `prometheus.fullnameOverride`          | String to fully override common.names.fullname                                  | `prometheus` |
| `prometheus.replicaCount`              | Number of Prometheus replicas to deploy                                         | `1`          |
| `prometheus.resources`                 | Resource configuration for Clickhouse container                                 | `{}`         |
| `prometheus.persistence.enabled`       | Enable persistence using PVC                                                    | `true`       |
| `prometheus.persistence.storageClass`  | Persistent volume storage class                                                 | `""`         |
| `prometheus.persistence.size`          | Data volume size                                                                | `5Gi`        |
| `prometheus.persistence.existingClaim` | Name of an existing PVC                                                         | `""`         |
| `prometheus.persistence.selector`      | Template to specify an existing persistent volume                               | `{}`         |
| `prometheus.scrapeConfigmap`           | ConfigMap which contains scrape config files                                    | `cs-metrics` |
