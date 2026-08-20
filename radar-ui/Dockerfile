FROM node:20.19.6-alpine AS builder

RUN --mount=type=cache,target=/var/cache/apk \
    apk update && \
    apk add --no-cache \
        jq \
        git \
        openssh \
        bash \
        libc6-compat \
        openssl \
        curl \
        gpgsm \
        gpg-agent && \
    npm install -g pnpm@8

WORKDIR /build
COPY . .

RUN jq -cr .version package.json > version.json && \
    yarn install && \
    yarn build

FROM scratch AS exporter
COPY --from=builder /build/dist/apps/runtime-radar /dist/apps/runtime-radar
