# kube-manager
The Kube Manager service collects information about entities in the K8S cluster and provides interfaces for obtaining this data.

## Autoscaler

Kube Manager can scale the deployments that consume the RabbitMQ event queues — `event-processor` and
`history-api` — from the backlog of the queue each one reads. The loop is disabled by default and is
turned on with `AUTOSCALER_ENABLED`.

### Signal

Every `AUTOSCALER_INTERVAL` the service reads the number of messages ready for delivery in each
target queue with a passive AMQP queue inspect (`lib/rabbit`, `MessageBroker.InspectQueue`). Every
target probes through its own message broker, bound to that target's queue. The same inspect also
returns the consumer count of the queue, which is exported and warned about but never scaled on.

Each broker is opened by the leading replica when it takes its first sample, not while the service
starts: an unreachable RabbitMQ therefore delays the scaling instead of stopping Kube Manager, whose
other duties — the pod and node inventory, `KillPod` — need no broker at all.
The connection is retried on every sample and the loop starts scaling as soon as one succeeds.

Consumers acknowledge a message as soon as they read it, so messages already in flight are not
counted and the backlog reports saturation with a lag. That is acceptable for scaling: consumers
that cannot keep up leave messages in the queue, and a sustained overload always shows up as a
growing backlog.

### Rule

Per target, on every sample:

- backlog above `SCALE_UP_BACKLOG` for `SCALE_UP_WINDOW` consecutive samples → `+1` replica;
- backlog below `SCALE_DOWN_BACKLOG` for `SCALE_DOWN_WINDOW` consecutive samples → `-1` replica;
- a backlog between the two thresholds is the neutral band: it resets both counters;
- the result is clamped to `[MIN_REPLICAS, MAX_REPLICAS]`.

The loop holds off in two cases. While the deployment is unhealthy (`status.availableReplicas` <
`spec.replicas`) or mid-rollout (`status.observedGeneration` < `metadata.generation`) it adds no
replicas: the scale-up and the correction up to `MIN_REPLICAS` wait, while the scale-down and the
correction down to `MAX_REPLICAS` still go through — replicas a stuck deployment cannot start are
the ones it needs least. The counters keep counting meanwhile, so a window that filled up during a
rollout fires on the tick right after it. And it takes no action within `AUTOSCALER_COOLDOWN` of the
previous one: the time of the last action is stored in the deployment annotation
`runtime-radar.io/autoscaler-last-scaled-at`, which is the only persistent state of the autoscaler,
so the cooldown survives a restart or a leadership change.

A failed queue probe — a broker that is down, or one that was never reachable to begin with —
resets the counters and skips the tick; a deployment changed by someone else between the read and
the update skips the tick too, and the next sample re-reads and decides again.

Replicas outside `[MIN_REPLICAS, MAX_REPLICAS]` are corrected before the windows are considered. A
Helm upgrade resets `spec.replicas` to the chart value, so `minReplicas` should equal the target's
`replicas` in the chart: the deployment drops to that floor on upgrade and the loop climbs back from
there.

The same correction applies to a manual `kubectl scale`: while the autoscaler is enabled, a target
scaled below `MIN_REPLICAS` — including down to zero — is clamped back to `MIN_REPLICAS` on the next
tick, unless the deployment is unhealthy or mid-rollout, in which case the correction waits for it
to settle. `MIN_REPLICAS` itself cannot be lowered to zero, it is rejected at startup. To scale a target
to zero for maintenance, either turn the whole loop off with `AUTOSCALER_ENABLED=false`, or — as long
as the other target stays enabled — set that one target's `ENABLED=false`. Turning both targets off
while `AUTOSCALER_ENABLED` is set is a startup failure, not a way to switch the loop off.

### Leader election

All Kube Manager replicas take part in the election, only the winner runs the loop. The lock is a
Kubernetes Lease named `kube-manager-autoscaler` in `CS_NAMESPACE`; the identity of a replica is its
pod hostname. The lease is released on shutdown, so a surviving replica takes over at once instead
of waiting out the lease duration.

### Configuration

The autoscaler needs `CS_NAMESPACE` (the namespace holding the target deployments and the lease) and
the RabbitMQ connection parameters.

```
  --autoscalerEnabled || AUTOSCALER_ENABLED environment variable
        Set to enable queue driven autoscaling of the event consuming deployments.
  --autoscalerInterval || AUTOSCALER_INTERVAL environment variable
        Interval between queue backlog samples. (default 10s)
  --autoscalerCooldown || AUTOSCALER_COOLDOWN environment variable
        Minimum time between two scaling actions on the same target. (default 60s)
  --rabbitAddr string || RABBIT_ADDR environment variable
        Set RabbitMQ address in host[:port] format. (default "rabbitmq.default:5672")
  --rabbitUser string || RABBIT_USER environment variable
        Set RabbitMQ user. (default "guest")
  --rabbitPassword string || RABBIT_PASSWORD environment variable
        Set RabbitMQ password. (default "guest")
```

Each target is configured with its own set of variables, prefixed with
`AUTOSCALER_EVENT_PROCESSOR_` and `AUTOSCALER_HISTORY_API_` respectively (command line equivalents
are `--autoscalerEventProcessor*` and `--autoscalerHistoryAPI*`). Both targets are enabled by
default, and are only sampled when `AUTOSCALER_ENABLED` is set.

| Variable suffix      | event-processor  | history-api      | Meaning                                                            |
|----------------------|------------------|------------------|--------------------------------------------------------------------|
| `ENABLED`            | `true`           | `true`           | is scaling of this target enabled?                                  |
| `DEPLOYMENT`         | `event-processor`| `history-api`    | name of the deployment to scale                                     |
| `QUEUE`              | `runtime_events` | `history_events` | queue whose backlog drives the scaling                              |
| `MIN_REPLICAS`       | `1`              | `1`              | lower replica bound                                                 |
| `MAX_REPLICAS`       | `20`             | `20`             | upper replica bound                                                 |
| `SCALE_UP_BACKLOG`   | `1000`           | `20000`          | backlog above which the target scales up                            |
| `SCALE_UP_WINDOW`    | `4`              | `4`              | consecutive samples above `SCALE_UP_BACKLOG` required to scale up   |
| `SCALE_DOWN_BACKLOG` | `10`             | `200`            | backlog below which the target scales down                          |
| `SCALE_DOWN_WINDOW`  | `10`             | `10`             | consecutive samples below `SCALE_DOWN_BACKLOG` required to scale down |

Settings are validated at startup and the service refuses to start on an invalid combination. With
`AUTOSCALER_ENABLED` set:

- `CS_NAMESPACE` must not be empty;
- `AUTOSCALER_INTERVAL` must be > 0, `AUTOSCALER_COOLDOWN` >= 0;
- at least one target must be enabled — switch the loop off with `AUTOSCALER_ENABLED=false` rather
  than by turning off all of its targets;

and for every enabled target:

- `DEPLOYMENT` and `QUEUE` must not be empty;
- `MIN_REPLICAS` must be >= 1 — zero is not a legal floor;
- `MAX_REPLICAS` must be >= `MIN_REPLICAS` and must fit in an int32;
- `SCALE_UP_WINDOW` and `SCALE_DOWN_WINDOW` must be >= 1;
- `SCALE_UP_BACKLOG` must be >= 0 and `SCALE_DOWN_BACKLOG` >= 1 — the rule tests `backlog <
  SCALE_DOWN_BACKLOG`, so a floor of zero could never be reached;
- `SCALE_DOWN_BACKLOG` must be less than `SCALE_UP_BACKLOG`.

Only the settings are checked that early. Whether RabbitMQ answers on `RABBIT_ADDR`, and whether the
target deployments exist, is found out by the loop while it runs and is reported in the logs and the
metrics, never by refusing to start.

### Enabling in Helm

The umbrella chart exposes the same settings under `kube-manager.autoscaler`, off by default:

```yaml
kube-manager:
  autoscaler:
    enabled: true
    interval: 10s
    cooldown: 60s
    eventProcessor:
      queue: runtime_events
      minReplicas: 1
      maxReplicas: 20
    historyApi:
      queue: history_events
      minReplicas: 1
      maxReplicas: 20
```

Each queue name must match the two chart values the producer and the consumer of that queue are
configured with: `runtime_events` is `runtime-monitor.rabbitmq.queue` and
`event-processor.rabbitmq.runtimeEventsQueue`, `history_events` is
`event-processor.rabbitmq.historyEventsQueue` and `history-api.rabbitmq.queue`. `minReplicas` should
be kept equal to the target's `replicas`. The RabbitMQ connection parameters are not exposed as
autoscaler values: the chart mounts the shared `rabbitmq` Secret and ConfigMap into kube-manager, so
`RABBIT_ADDR`, `RABBIT_USER` and `RABBIT_PASSWORD` are already set.

### RBAC

The autoscaler only ever touches the release namespace, so its permissions are a namespaced Role,
rendered next to the pre-existing ClusterRole whenever `serviceAccount.create` is set — regardless
of `autoscaler.enabled`, so turning the autoscaler on through `AUTOSCALER_ENABLED` alone (e.g.
`kubectl set env`) needs no chart change. It grants, in the release namespace:

- `apps` / `deployments`: `get`, `update`, narrowed by `resourceNames` to the two target deployments
  (on top of the cluster-wide `list` and `watch` the cache already needs);
- `coordination.k8s.io` / `leases`: `get`, `create`, `update` — for the leader election lease;
- core / `events`: `create`, `patch` — for the Kubernetes Events of the scaling actions.

With `serviceAccount.create=false` neither the Role nor the RoleBinding is rendered: grant the same
rules in the release namespace to the ServiceAccount you pre-created, or the leader election and
every tick fail with `forbidden`.

### Observability

Metrics on `/metrics` of the instrumentation server:

- `autoscaler_queue_backlog{target,queue}` — ready messages in the target queue at the last sample.
  A failed probe leaves the gauge at the last value it read rather than reporting an empty queue, so
  alert on it together with the `Can't probe queue` warning;
- `autoscaler_queue_consumers{target,queue}` — consumers attached to the target queue. The broker
  declares a queue before inspecting it, so a `QUEUE` that matches nothing reads as a permanently
  empty queue instead of an error: a count of zero next to a running deployment means the name is
  wrong, and is logged as a warning too;
- `autoscaler_replicas{target}` — replicas of the target deployment;
- `autoscaler_scale_actions_count{target,direction}` — scaling actions, `direction` is `up` or `down`;
- `autoscaler_is_leader` — `1` on the replica that holds the lease, `0` on the others.

The three per-target gauges describe what the leader sees and are dropped when a replica leaves the
lease, so a replica exports them only while `autoscaler_is_leader` is `1`.

Every scaling action is logged at `INFO`, as is this replica taking, losing and releasing the lease;
a replica noting that another one holds the lease logs at `DEBUG`. Each action also emits a `Normal`
Kubernetes Event with reason `Scaled` on the scaled deployment, from the `kube-manager-autoscaler`
component.

### Verifying

```
kubectl -n <namespace> describe deploy event-processor
```

shows the current `Replicas`, the `runtime-radar.io/autoscaler-last-scaled-at` annotation and the
`Scaled` events with the backlog that triggered them. The current leader is in

```
kubectl -n <namespace> get lease kube-manager-autoscaler -o yaml
```

as `spec.holderIdentity` — the hostname of the Kube Manager pod running the loop.
