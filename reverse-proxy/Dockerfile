ARG GO_VERSION=1.26.0

FROM golang:${GO_VERSION} AS tools

RUN CGO_ENABLED=0 GOBIN=/usr/bin go install github.com/go-task/task/v3/cmd/task@v3.38.0

FROM golang:${GO_VERSION} AS builder

COPY --from=tools /usr/bin/task /usr/bin
COPY . /app
WORKDIR /app

RUN task srv

FROM caddy:2.8.4-alpine AS linter
WORKDIR /app
COPY --from=tools /usr/bin/task /usr/bin
COPY --from=builder /app/*.pem /
COPY Taskfile.yml ./
COPY Caddyfile ./
RUN task lint

FROM caddy:2.8.4-alpine

LABEL org.opencontainers.image.source=https://github.com/runtime-radar/runtime-radar
LABEL org.opencontainers.image.description="reverse-proxy serves Runtime Radar UI and proxies requests to Runtime Radar components"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=linter /app/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/srv /srv
COPY --from=builder /app/*.pem /

ENV TLS=false

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
