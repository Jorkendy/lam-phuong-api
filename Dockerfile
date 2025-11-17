# syntax=docker/dockerfile:1

#########################
# 1. BUILD STAGE
#########################
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Bật GOOS env cho build static
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy dependency files trước
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Lấy version & commit từ git (Coolify clone repo nên .git tồn tại)
RUN VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev") && \
    BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) && \
    echo "Building version=$VERSION commit=$COMMIT" && \
    go build -ldflags "-s -w \
      -X 'lam-phuong-api/internal/buildinfo.Version=${VERSION}' \
      -X 'lam-phuong-api/internal/buildinfo.Commit=${COMMIT}' \
      -X 'lam-phuong-api/internal/buildinfo.BuildTime=${BUILD_TIME}'" \
      -o server ./cmd/server

#########################
# 2. RUNTIME STAGE
#########################
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy CA certificates (Go TLS cần để call HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/server .

# Expose API port
EXPOSE 8080

# Run as nonroot user
USER nonroot:nonroot

ENTRYPOINT ["./server"]
