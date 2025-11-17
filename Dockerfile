# syntax=docker/dockerfile:1

### Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy go.mod trước để cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
# nếu main.go ở root thì: go build -o server .

### Run stage (image nhỏ, ít bug)
FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080
# nếu API listen port khác thì sửa lại & EXPOSE đúng
USER nonroot:nonroot

CMD ["./server"]
