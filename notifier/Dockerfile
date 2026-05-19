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
ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" \
    mkdir -p bin && \
    GOBIN=/go/src/repo/bin go install github.com/google/gops

COPY --from=tools /usr/bin/task /usr/bin
COPY notifier/ notifier/
COPY lib/ lib/
COPY history-api/api/ history-api/api/
COPY history-api/pkg/model/ history-api/pkg/model/
COPY policy-enforcer/api/ policy-enforcer/api/
COPY policy-enforcer/pkg/model/ policy-enforcer/pkg/model/
COPY event-processor/api/ event-processor/api/

RUN --mount=type=cache,target="/root/.cache/go-build" task -d notifier build

FROM debian:12.0-slim AS runner

LABEL org.opencontainers.image.source=https://github.com/runtime-radar/runtime-radar
LABEL org.opencontainers.image.description="Notifier service is responsible for sending notifications"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ARG APP_NAME=notifier

COPY --from=builder /go/src/repo/${APP_NAME}/cmd/${APP_NAME}/${APP_NAME} /app
COPY --from=builder /go/src/repo/${APP_NAME}/cmd/${APP_NAME}/*.pem /
COPY --from=builder /go/src/repo/notifier/templates /templates
COPY --from=builder /go/src/repo/bin/gops /usr/bin/gops

EXPOSE 8000 9000
ENTRYPOINT ["/app"]
