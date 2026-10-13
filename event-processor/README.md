# event-processor
Event Processor consumes Runtime Monitor events and enriches them for downstream services.

When `kube-manager.autoscaler` is enabled, kube-manager scales this deployment from the backlog of
the `runtime_events` queue (see [the autoscaler](../kube-manager/README.md#autoscaler)) and stamps
the Deployment with the `runtime-radar.io/autoscaler-last-scaled-at` annotation.
