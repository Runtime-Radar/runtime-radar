ARG GO_VERSION=1.26.0

FROM golang:${GO_VERSION} AS tools

RUN CGO_ENABLED=0 GOBIN=/usr/bin go install github.com/go-task/task/v3/cmd/task@v3.38.0

FROM golang:${GO_VERSION} AS builder

ARG BUILD_RELEASE
ARG BUILD_BRANCH
ARG BUILD_COMMIT

WORKDIR /go/src/repo
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY pkg/tools/tools.go pkg/tools/

ENV GOCACHE=/root/.cache/go-build

COPY --from=tools /usr/bin/task /usr/bin
COPY runtime-monitor/ runtime-monitor/
COPY lib/ lib/

RUN --mount=type=cache,target="/root/.cache/go-build" task -d runtime-monitor build

# Runs as root: runtime-monitor talks to Tetragon, which runs as root in the same pod.
FROM gcr.io/distroless/static-debian13 AS runner

LABEL org.opencontainers.image.source=https://github.com/runtime-radar/runtime-radar
LABEL org.opencontainers.image.description="Runtime monitor monitors runtime events in k8s"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ARG APP_NAME=runtime-monitor

COPY --from=builder /go/src/repo/${APP_NAME}/cmd/${APP_NAME}/${APP_NAME} /app
COPY --from=builder /go/src/repo/${APP_NAME}/cmd/${APP_NAME}/*.pem /

EXPOSE 8000 9000
ENTRYPOINT ["/app"]
