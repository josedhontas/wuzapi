# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
ENV CGO_ENABLED=1
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o wuzapi

FROM debian:bookworm-slim

# Install runtime dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    netcat-openbsd \
    postgresql-client \
    openssl \
    curl \
    ffmpeg \
    tzdata \
    && apt-get clean

ENV TZ="America/Sao_Paulo"
WORKDIR /app

COPY --from=builder /app/wuzapi         /app/
COPY --from=builder /app/static         /app/static/
COPY --from=builder /app/wuzapi.service /app/wuzapi.service

RUN chmod +x /app/wuzapi && \
    chmod -R 755 /app && \
    chown -R root:root /app

ENTRYPOINT ["/app/wuzapi", "--logtype=console", "--color=true"]
