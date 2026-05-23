# Chart

## Parameters

### Global parameters

| Name                                  | Description                                                                                 | Value |
| ------------------------------------- | ------------------------------------------------------------------------------------------- | ----- |
| `global.imageRegistry`                | Global Docker image registry                                                                | `""`  |
| `global.imagePullSecrets`             | Global Docker registry secret names as an array                                             | `[]`  |
| `global.storageClass`                 | Global StorageClass for Persistent Volume(s)                                                | `""`  |
| `global.rabbitmq.auth.existingSecret` | Name of an existing secret that contains the certificates (overrides `auth.existingSecret`) | `""`  |

### Common parameters

| Name                | Description                                                                            | Value           |
| ------------------- | -------------------------------------------------------------------------------------- | --------------- |
| `nameOverride`      | String to partially override common.fullname template (will maintain the release name) | `""`            |
| `fullnameOverride`  | String to fully override common.fullname template                                      | `""`            |
| `namespaceOverride` | String to fully override common.namespace                                              | `""`            |
| `clusterDomain`     | Kubernetes Cluster Domain                                                              | `cluster.local` |
| `labels`            | Add labels to all the deployed resources                                               | `{}`            |
| `annotations`       | Add annotations to all the deployed resources                                          | `{}`            |
| `imagePullSecrets`  | Global Docker registry secret names as an array                                        | `[]`            |

### RabbitMQ parameters

| Name                             | Description                                                                                                              | Value                                             |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------- |
| `image.registry`                 | RabbitMQ image registry                                                                                                  | `""`                                              |
| `image.repository`               | RabbitMQ image repository                                                                                                | `rabbitmq`                                        |
| `image.tag`                      | RabbitMQ image tag (immutable tags are recommended)                                                                      | `4.1.1-management`                                |
| `image.digest`                   | RabbitMQ image digest in the way sha256:aa.... Please note this parameter, if set, will override the tag                 | `""`                                              |
| `image.pullPolicy`               | RabbitMQ image pull policy                                                                                               | `IfNotPresent`                                    |
| `auth.username`                  | RabbitMQ application username                                                                                            | `user`                                            |
| `auth.password`                  | RabbitMQ application password                                                                                            | `""`                                              |
| `auth.existingSecret`            | Existing secret with RabbitMQ credentials (must contain a value for `rabbitmq-password` key)                             | `""`                                              |
| `auth.existingSecretPasswordKey` | Password key to be retrieved from existing secret                                                                        | `rabbitmq-password`                               |
| `resources`                      | Set container requests and limits for different resources like CPU or memory                                             | `{}`                                              |
| `logs`                           | Path of the RabbitMQ server's Erlang log file. Value for the `RABBITMQ_LOGS` environment variable                        | `-`                                               |
| `ulimitNofiles`                  | RabbitMQ Max File Descriptors                                                                                            | `65536`                                           |
| `plugins`                        | List of default plugins to enable (should only be altered to remove defaults; for additional plugins use `extraPlugins`) | `rabbitmq_management rabbitmq_peer_discovery_k8s` |
| `configuration`                  | RabbitMQ Configuration file content: required cluster configuration                                                      | `""`                                              |
| `existingConfigmap`              | The name of an existing ConfigMap with your custom configuration                                                         | `""`                                              |
| `volumeMounts`                   | Optionally specify extra list of additional volumeMounts                                                                 | `[]`                                              |
| `volumes`                        | Optionally specify extra list of additional volumes .                                                                    | `[]`                                              |

### Statefulset parameters

| Name                                                | Description                                                                             | Value            |
| --------------------------------------------------- | --------------------------------------------------------------------------------------- | ---------------- |
| `replicaCount`                                      | Number of RabbitMQ replicas to deploy                                                   | `1`              |
| `podLabels`                                         | RabbitMQ Pod labels. Evaluated as a template                                            | `{}`             |
| `podAnnotations`                                    | RabbitMQ Pod annotations. Evaluated as a template                                       | `{}`             |
| `affinity`                                          | Affinity for pod assignment. Evaluated as a template                                    | `{}`             |
| `nodeSelector`                                      | Node labels for pod assignment. Evaluated as a template                                 | `{}`             |
| `tolerations`                                       | Tolerations for pod assignment. Evaluated as a template                                 | `[]`             |
| `terminationGracePeriodSeconds`                     | Default duration in seconds k8s waits for container to exit before sending kill signal. | `120`            |
| `extraEnvVars`                                      | Extra environment variables to add to RabbitMQ pods                                     | `[]`             |
| `containerPorts.amqp`                               | AMQP                                                                                    | `5672`           |
| `containerPorts.dist`                               | dist                                                                                    | `25672`          |
| `containerPorts.manager`                            | manager                                                                                 | `15672`          |
| `containerPorts.epmd`                               | EPMD                                                                                    | `4369`           |
| `containerPorts.metrics`                            | metrics                                                                                 | `9419`           |
| `podSecurityContext.fsGroupChangePolicy`            | Set filesystem group change policy                                                      | `Always`         |
| `podSecurityContext.sysctls`                        | Set kernel settings using the sysctl interface                                          | `[]`             |
| `podSecurityContext.supplementalGroups`             | Set filesystem extra groups                                                             | `[]`             |
| `podSecurityContext.fsGroup`                        | Set RabbitMQ pod's Security Context fsGroup                                             | `1001`           |
| `containerSecurityContext.seLinuxOptions`           | Set SELinux options in container                                                        | `nil`            |
| `containerSecurityContext.runAsUser`                | Set RabbitMQ containers' Security Context runAsUser                                     | `1001`           |
| `containerSecurityContext.runAsGroup`               | Set RabbitMQ containers' Security Context runAsGroup                                    | `1001`           |
| `containerSecurityContext.runAsNonRoot`             | Set RabbitMQ container's Security Context runAsNonRoot                                  | `true`           |
| `containerSecurityContext.allowPrivilegeEscalation` | Set container's privilege escalation                                                    | `false`          |
| `containerSecurityContext.readOnlyRootFilesystem`   | Set container's Security Context readOnlyRootFilesystem                                 | `true`           |
| `containerSecurityContext.capabilities.drop`        | Set container's Security Context runAsNonRoot                                           | `["ALL"]`        |
| `containerSecurityContext.seccompProfile.type`      | Set container's Security Context seccomp profile                                        | `RuntimeDefault` |
| `livenessProbe.initialDelaySeconds`                 | Initial delay seconds for livenessProbe                                                 | `120`            |
| `livenessProbe.timeoutSeconds`                      | Timeout seconds for livenessProbe                                                       | `20`             |
| `livenessProbe.periodSeconds`                       | Period seconds for livenessProbe                                                        | `30`             |
| `livenessProbe.failureThreshold`                    | Failure threshold for livenessProbe                                                     | `6`              |
| `livenessProbe.successThreshold`                    | Success threshold for livenessProbe                                                     | `1`              |
| `readinessProbe.initialDelaySeconds`                | Initial delay seconds for readinessProbe                                                | `10`             |
| `readinessProbe.timeoutSeconds`                     | Timeout seconds for readinessProbe                                                      | `20`             |
| `readinessProbe.periodSeconds`                      | Period seconds for readinessProbe                                                       | `30`             |
| `readinessProbe.failureThreshold`                   | Failure threshold for readinessProbe                                                    | `3`              |
| `readinessProbe.successThreshold`                   | Success threshold for readinessProbe                                                    | `1`              |
| `startupProbe.initialDelaySeconds`                  | Initial delay seconds for startupProbe                                                  | `10`             |
| `startupProbe.timeoutSeconds`                       | Timeout seconds for startupProbe                                                        | `20`             |
| `startupProbe.periodSeconds`                        | Period seconds for startupProbe                                                         | `30`             |
| `startupProbe.failureThreshold`                     | Failure threshold for startupProbe                                                      | `3`              |
| `startupProbe.successThreshold`                     | Success threshold for startupProbe                                                      | `1`              |

### Service parameters

| Name                           | Description                                  | Value       |
| ------------------------------ | -------------------------------------------- | ----------- |
| `service.type`                 | Kubernetes Service type                      | `ClusterIP` |
| `service.ports.amqp`           | Amqp service port                            | `5672`      |
| `service.ports.dist`           | Erlang distribution service port             | `25672`     |
| `service.ports.manager`        | RabbitMQ Manager service port                | `15672`     |
| `service.ports.metrics`        | RabbitMQ Prometheues metrics service port    | `9419`      |
| `service.ports.epmd`           | EPMD Discovery service port                  | `4369`      |
| `service.annotations`          | Service annotations. Evaluated as a template | `{}`        |
| `service.headless.annotations` | Annotations for the headless service.        | `{}`        |

### Persistence Parameters

| Name                                               | Description                                                             | Value               |
| -------------------------------------------------- | ----------------------------------------------------------------------- | ------------------- |
| `persistence.enabled`                              | Enable persistence using Persistent Volume Claims                       | `true`              |
| `persistence.existingClaim`                        | Name of an existing PVC to use                                          | `""`                |
| `persistence.storageClass`                         | Storage class of backing PVC                                            | `""`                |
| `persistence.labels`                               | Persistent Volume Claim labels                                          | `{}`                |
| `persistence.annotations`                          | Persistent Volume Claim annotations                                     | `{}`                |
| `persistence.accessModes`                          | Persistent Volume Access Modes                                          | `["ReadWriteOnce"]` |
| `persistence.size`                                 | Size of data volume                                                     | `8Gi`               |
| `persistence.selector`                             | Selector to match an existing Persistent Volume for ClickHouse data PVC | `{}`                |
| `persistence.dataSource`                           | Custom PVC data source                                                  | `{}`                |
| `persistence.persistentVolumeClaimRetentionPolicy` | PostgreSQL Persistent Volume Claim Retention Policy                     | `{}`                |

### Other Parameters

| Name                          | Description                                                                                | Value  |
| ----------------------------- | ------------------------------------------------------------------------------------------ | ------ |
| `serviceAccount.create`       | Enable creation of ServiceAccount for RabbitMQ pods                                        | `true` |
| `serviceAccount.name`         | Name of the created serviceAccount                                                         | `""`   |
| `serviceAccount.automount`    | Auto-mount the service account token in the pod                                            | `true` |
| `serviceAccount.annotations`  | Annotations for service account. Evaluated as a template. Only used if `create` is `true`. | `{}`   |
| `networkPolicy.enabled`       | Specifies whether a NetworkPolicy should be created                                        | `true` |
| `networkPolicy.allowExternal` | Don't require server label for connections                                                 | `true` |
| `networkPolicy.extraIngress`  | Add extra ingress rules to the NetworkPolicy                                               | `[]`   |
| `networkPolicy.extraEgress`   | Add extra ingress rules to the NetworkPolicy                                               | `[]`   |

### Metrics Parameters

| Name                     | Description                                                        | Value                 |
| ------------------------ | ------------------------------------------------------------------ | --------------------- |
| `metrics.enabled`        | Enable exposing RabbitMQ metrics to be gathered by Prometheus      | `false`               |
| `metrics.externalHost`   | RabbitMQ metrics external host                                     | `""`                  |
| `metrics.plugins`        | Plugins to enable Prometheus metrics in RabbitMQ                   | `rabbitmq_prometheus` |
| `metrics.podAnnotations` | Annotations for enabling prometheus to access the metrics endpoint | `{}`                  |
