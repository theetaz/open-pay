# Dev image for Open Pay Go services — hot reload via air.
# One image, shared by all services; the SERVICE env var selects which to build/run.
# Source is bind-mounted at runtime (see docker-compose.dev.yml), so this image
# only needs the toolchain + a warmed module cache.

FROM golang:1.25-alpine

RUN apk add --no-cache git ca-certificates wget

# Hot-reload tool
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Warm the module cache into the image layer (persists in /go/pkg/mod, which is
# NOT bind-mounted at runtime, so every service container reuses it for free).
COPY go.mod go.sum ./
RUN go mod download

# Per-service air launcher (reads $SERVICE / $PORT at runtime)
COPY scripts/air-entrypoint.sh /usr/local/bin/air-entrypoint.sh
RUN chmod +x /usr/local/bin/air-entrypoint.sh

EXPOSE 8080

CMD ["air-entrypoint.sh"]
