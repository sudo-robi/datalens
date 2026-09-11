# Multi-stage build for Go ingestion service
FROM golang:1.24-alpine AS build

# Install ca-certificates and git for fetching dependencies
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o server .

# Production image — minimal attack surface
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/server .
RUN chown app:app server

USER app

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8081/health || exit 1

CMD ["./server"]
