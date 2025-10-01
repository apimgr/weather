# Multi-stage build for Go weather service
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with optimization
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o weather .

# Final stage - minimal runtime image
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy templates and static files
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static

# Copy the binary
COPY --from=builder /app/weather /weather

# Set working directory
WORKDIR /

# Expose port
EXPOSE 3000

# Health check using wget (included in scratch via static binary)
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD ["/weather", "healthcheck"]

# Run as non-root user (nobody)
USER 65534:65534

# Start the application
ENTRYPOINT ["/weather"]
