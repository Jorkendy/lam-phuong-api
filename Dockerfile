# syntax=docker/dockerfile:1

#########################
# 1. BUILD STAGE
#########################
FROM golang:1.24.5-alpine AS builder

# Tuỳ, nhưng nên fixed cho reproducible
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Copy go.mod trước để cache dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source vào container build
COPY . .

# (OPTIONAL) nếu muốn lấy version/commit từ git để embed vào binary
# Tạo 1 package ví dụ internal/buildinfo với:
#   var Version = "dev"
#   var Commit  = "dev"
# rồi chỉnh lại lệnh go build bên dưới để set -X (mình ghi comment sẵn):

# Lấy version & commit từ git (nếu fail thì fallback "dev")
# RUN VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
#     COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev") && \
#     go build -ldflags "-s -w \
#       -X 'your/module/internal/buildinfo.Version=${VERSION}' \
#       -X 'your/module/internal/buildinfo.Commit=${COMMIT}'" \
#       -o server ./cmd/server

# Nếu chưa cần embed version thì build bình thường:
RUN go build -o server ./cmd/server
# Nếu main.go ở root:
# RUN go build -o server .

#########################
# 2. RUNTIME STAGE
#########################
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy CA certs nếu app gọi HTTPS tới service khác
# (alpine base đã có, copy sang runtime image)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary đã build
COPY --from=builder /app/server .

# Port API lắng nghe
EXPOSE 8080

# Chạy dưới user không phải root
USER nonroot:nonroot

# Chạy app
ENTRYPOINT ["./server"]
