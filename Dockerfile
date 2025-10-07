# Multi-stage build - NO interpretation allowed
FROM alpine:latest AS builder

# Install requirements
RUN apk add --no-cache bash curl make go

# Copy source
WORKDIR /build
COPY . .

# Build binary
RUN make build

# Runtime stage
FROM scratch

# Copy binary only
COPY --from=builder /build/binaries/weather-linux-amd64 /weather

# Metadata labels (required)
LABEL org.opencontainers.image.source="https://github.com/apimgr/weather"
LABEL org.opencontainers.image.description="weather server"
LABEL org.opencontainers.image.licenses="MIT"

# Expose port (informational only)
EXPOSE 80

# Run
ENTRYPOINT ["/weather"]
