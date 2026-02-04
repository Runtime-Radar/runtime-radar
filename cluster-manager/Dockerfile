FROM golang:1.25.0 AS tools

RUN CGO_ENABLED=0 GOBIN=/usr/bin go install github.com/go-task/task/v3/cmd/task@v3.38.0

FROM golang:1.25.0 AS builder

ARG BUILD_RELEASE
ARG BUILD_BRANCH
ARG BUILD_COMMIT

WORKDIR /go/src/repo
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY pkg/tools/tools.go pkg/tools/

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" \
    mkdir -p bin && \
    GOBIN=/go/src/repo/bin go install github.com/google/gops

COPY --from=tools /usr/bin/task /usr/bin
COPY . /go/src/repo

RUN task build

FROM debian:bullseye AS runner

LABEL org.opencontainers.image.source=https://github.com/runtime-radar/runtime-radar
LABEL org.opencontainers.image.description="This is a Cluster Manager service which controls child clusters from the main cluster"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ARG APP_NAME=cluster-manager

COPY --from=builder /go/src/repo/cmd/${APP_NAME}/${APP_NAME} /app
COPY --from=builder /go/src/repo/cmd/${APP_NAME}/*.pem /
COPY --from=builder /go/src/repo/bin/gops /usr/bin/gops

EXPOSE 8000 9000
ENTRYPOINT ["/app"]
