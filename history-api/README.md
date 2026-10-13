# history-api
History API stores and retrieves historical Runtime Radar security data.

When `kube-manager.autoscaler` is enabled, kube-manager scales this deployment from the backlog of
the `history_events` queue (see [the autoscaler](../kube-manager/README.md#autoscaler)) and stamps
the Deployment with the `runtime-radar.io/autoscaler-last-scaled-at` annotation.
