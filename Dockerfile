# Styx: Unified Dev Environment Manager
# Docker image for CI/CD pipelines and containerized environments

FROM --platform=linux/amd64 alpine:3.19 as builder

# Install build dependencies
RUN apk add --no-cache go git ca-certificates

# Copy source code
COPY . /src
WORKDIR /src

# Build styx binary
RUN go build -ldflags="-s -w" -o /usr/local/bin/styx ./cmd/styx && \
    chmod +x /usr/local/bin/styx

# Verify binary works
RUN styx --version

# Runtime image: minimal base with styx
FROM --platform=linux/amd64 alpine:3.19

LABEL maintainer="Modo Ventures <noreply@killallservers.dev>"
LABEL description="Styx: Unified dev environment manager for portfolio companies"
LABEL version="0.1.0"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    tar \
    gzip \
    bash \
    git \
    && rm -rf /var/cache/apk/*

# Copy styx binary from builder
COPY --from=builder /usr/local/bin/styx /usr/local/bin/styx

# Create styx directories
RUN mkdir -p /root/.styx/bin /root/.styx/store /root/.styx/cache

# Set working directory
WORKDIR /workspace

# Make styx available in PATH
ENV PATH="/root/.styx/bin:${PATH}"

# Default command: show help
ENTRYPOINT ["styx"]
CMD ["--help"]

# Health check: verify styx works
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
    CMD styx --version || exit 1
