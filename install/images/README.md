# Base images

Minimal, hardened [Wolfi](https://github.com/wolfi-dev)-based base images built with [apko](https://github.com/chainguard-dev/apko). These are reusable runner bases that services can move onto instead of `debian:12.0-slim`. This is build-tooling only — no service Dockerfile is wired to them yet (see [Follow-up](#follow-up)).

## Images

| Config | Image | Extras |
|---|---|---|
| `base.apko.yaml` | `runtime-radar-base` | — (no shell) |
| `base-busybox.apko.yaml` | `runtime-radar-base-busybox` | `busybox` (`/bin/sh` + coreutils) |

Every base contains `wolfi-baselayout`, `ca-certificates-bundle`, and `tzdata`. All run **non-root** as username `nonroot` (uid/gid `65532`, `run-as: 65532`) and have **no `ENTRYPOINT`** — each service sets its own. Built for `amd64` only.

> **`runtime-radar-base` contains no shell** (no `busybox`, no `/bin/sh`, no `bash`). For any downstream Dockerfile with `FROM runtime-radar-base`:
> - `RUN` instructions **cannot be used** — there is no shell to execute them (the build will fail). Do all package/file preparation in an earlier builder stage and `COPY` the result in.
> - `CMD` / `ENTRYPOINT` **must use exec form** (JSON array, e.g. `ENTRYPOINT ["/app/myservice"]`). Shell form (`ENTRYPOINT /app/myservice`) won't work because it relies on `/bin/sh`.
>
> Use `runtime-radar-base-busybox` only for services that genuinely need a shell at runtime.

## Building

```bash
task images:build   # build all
```

By default the images are tagged with the current build date in `YYYYMMDD` form (e.g. `runtime-radar-base:20261013`). Override the tag and registry like the service Taskfiles:

```bash
IMAGE_TAG=v1.2.3 IMAGE_REGISTRY=my.local.registry task images:build
```

With no `IMAGE_REGISTRY` the images are tagged locally (`runtime-radar-base:<date>`, etc.).

Push the built images to a registry (build them first; pushing requires `IMAGE_REGISTRY`):

```bash
IMAGE_REGISTRY=my.local.registry task images:push   # push all
```

Point the service Dockerfiles at a base tag (build → push → patch, sharing one `IMAGE_TAG`):

```bash
IMAGE_TAG=v1.2.3 task images:patch   # bump all service Dockerfiles to :v1.2.3
```

`patch` rewrites the `:<tag>` of the `FROM ...runtime-radar-base*... AS runner` line in every service Dockerfile to `IMAGE_TAG` (default: today's date), leaving the registry prefix and each file's base variant (`runtime-radar-base` / `-busybox`) untouched. It discovers the Dockerfiles with `git grep`, so new services are picked up automatically. Re-running with the same tag is a no-op.

## Prerequisites

- **Docker.** The Taskfile runs apko via `docker run` and `docker load`s the result — no Dockerfile or BuildKit features required.
- **Package repository.** Packages come from the public Wolfi repository `https://packages.wolfi.dev/os`. The build host must reach it.
- **Signing key.** Each config references the Wolfi signing key by URL (`keyring: https://packages.wolfi.dev/os/wolfi-signing.rsa.pub`), so apko fetches it at build time — nothing is committed. If Wolfi rotates its key, apko fails with a signature-verification error; no local change is needed once the new key is served.
- **apko image / `ghcr.io` reachability.** The Taskfile runs `ghcr.io/wolfi-dev/apko@sha256:a1b1b11f504122eaa61da63eb9d07ca58d4287938adb99bb18c5135ba0ada99a` (apko v1.2.17). The repository publishes no version tags, only a rolling `latest`, so the image is pinned by digest to keep the build reproducible. To upgrade, resolve the digest of the new `latest` and check `docker run --rm <ref> version`. The build host must reach `ghcr.io`.

## Notes

The Taskfile runs apko directly (`docker run … apko build <config> <ref> /out/image.tar`), mounting the config dir read-only at `/work` and writing the tar + SBOM to a temp dir (the configs reference the signing key by URL, so nothing else needs mounting). apko appends the target arch when writing the tar, so a config built as `<name>:<tag>` loads as **`<name>:<tag>-amd64`**. The Taskfile then **re-tags** that loaded image to the final `IMAGE_REGISTRY`/`IMAGE_TAG` ref (e.g. `runtime-radar-base:20261013`) and removes the intermediate. Each image has a distinct intermediate tag, so the builds are independent (the aggregate `build` runs them in sequence for deterministic output).

After the build, the Taskfile verifies the produced image runs as the non-root uid `65532` (`docker inspect` guard) and fails the build otherwise. The guard inspects the intermediate tag *before* re-tagging to the final ref, so a rejected (root) image is never tagged with the authoritative name.

**Reproducibility caveat.** Only the top-level packages listed in each `*.apko.yaml` are version-pinned. The transitive dependency closure is resolved by apko against the live repository at build time, so builds are **not fully bit-reproducible** — a transitive package can change between builds. If full reproducibility is later required, generate an apko lockfile (`apko lock`) and build from it; that is out of scope here because it changes the build flow.

## Follow-up

Migrating the service Dockerfiles onto these bases (adjusting per-service `COPY`s and K8s `securityContext`/`fsGroup` for non-root) and pushing the bases to the registry are separate, out-of-scope tasks.
