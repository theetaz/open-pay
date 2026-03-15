# Multi-stage Dockerfile for any Go microservice
# Usage: docker build --build-arg SERVICE=payment -t openpay-payment .
# Must be run from the project root directory

FROM golang:1.24-alpine AS builder

ARG SERVICE

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the specific service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./services/${SERVICE}/cmd/

# Runtime image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
