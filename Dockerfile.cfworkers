# Multi-stage build for myCart Cloudflare Container Workers
# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy source code
COPY . .

# Build binary (no CGO, embedded assets included)
# -s: Strip debug info, -w: Strip DWARF symbols
RUN go build -ldflags="-s -w" -o mycart ./cmd

# Stage 2: Runtime (distroless for minimal attack surface)
FROM gcr.io/distroless/static-debian12

# Copy compiled binary from builder
COPY --from=builder /build/mycart /mycart

# Cloudflare Container Workers requirement
EXPOSE 8080

# Start server on port 8080
ENTRYPOINT ["/mycart", "serve", "--http", ":8080"]
