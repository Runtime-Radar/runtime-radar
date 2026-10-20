# runtime-monitor
Runtime Monitor observes Kubernetes workloads and emits runtime security events.

## Tetragon connection

In Kubernetes the service talks to the Tetragon agent over a Unix socket:
`unix:///var/run/tetragon/tetragon.sock` (the `TETRAGON_ADDR` / `--tetragonAddr` default in
`pkg/config/config.go`, and `tetragon.grpc.address` in `.helm/values.yaml`). TCP is deliberately
**not** supported there: the Tetragon DaemonSet runs with `hostNetwork: true`, so a TCP listener is
reachable without authentication from any hostNetwork pod on the node, which is CVE-2026-65960.

Tetragon creates the socket as `root:root` with mode `0660`, while the app container runs as uid
65532, so the chart sets `containerSecurityContext.runAsGroup: 0` plus
`podSecurityContext.supplementalGroups: [0]` (the fallback for devMode deploys, where the
container-level securityContext is not rendered) and mounts the `tetragon-run` hostPath read-only at
`/var/run/tetragon`. The socket directory is fixed at `/var/run/tetragon`: it is baked into the
`tetragon-run` hostPath and into the mount paths of both containers. Only the socket file name can
be changed via `tetragon.grpc.address`.

`docker-compose.yml` deliberately diverges: it passes `--tetragonAddr=tetragon:54321` explicitly, so
the dev stack keeps using TCP between containers and never exercises the Go default.
