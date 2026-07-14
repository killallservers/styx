# Docker Support

Styx provides Docker images for containerized environments and CI/CD pipelines.

## Quick Start

### Build Docker Image

```bash
docker build -t styx:latest .
```

### Run Styx in Container

```bash
# Show help
docker run --rm styx:latest --help

# Run styx commands
docker run --rm -v $HOME/.styx:/root/.styx -v $(pwd):/workspace styx:latest install

# Interactive shell with styx available
docker run --rm -it -v $(pwd):/workspace styx:latest /bin/bash
```

## Docker Compose

Two services for different use cases:

### Development: styx-dev

Interactive shell with styx available:

```bash
docker-compose run --rm styx-dev
# Inside container:
> styx install
> styx list
```

Features:
- Mounts host `.styx` directory (persistent configuration)
- Mounts project directory (read-write)
- Interactive bash shell
- Volume persistence between runs

### CI/CD: styx-ci

Non-interactive service for CI pipelines:

```bash
docker-compose run --rm styx-ci styx verify
```

Features:
- Read-only project mount (prevents accidental modifications)
- Ephemeral store (no persisted data)
- Suitable for GitHub Actions, GitLab CI, etc.
- Lightweight and fast

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build with Styx

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    container:
      image: styx:latest
    steps:
      - uses: actions/checkout@v4
      - name: Install tools
        run: styx sync
      - name: Verify installation
        run: styx verify
      - name: Run build
        run: go build ./cmd/styx
```

### GitLab CI Example

```yaml
stages:
  - build

build:
  stage: build
  image: styx:latest
  script:
    - styx sync
    - styx verify
    - go build ./cmd/styx
  artifacts:
    paths:
      - styx
```

## Configuration

### Environment Variables

Inside Docker container:

```bash
HOME=/root          # Home directory
PATH=/root/.styx/bin:/usr/bin:...  # Styx tools in PATH
CI=true             # Set in CI service
```

### Mount Points

**Development:**
- `/root/.styx` - Styx data directory (persistent)
- `/workspace` - Project directory (read-write)

**CI:**
- `/workspace` - Project directory (read-only)
- `/tmp/styx-store` - Ephemeral tool storage

## Image Details

### Base Image

- Alpine 3.19 (minimal, ~7 MB)
- Multi-stage build (reduces image size)

### Installed Packages

- ca-certificates (HTTPS support)
- curl (downloads)
- tar, gzip (extraction)
- bash (shell)
- git (version control)

### Styx Binary

- Built from source (`go build`)
- ~10 MB uncompressed
- Statically linked (portable)
- Embedded registry included

### Final Image Size

~50 MB (alpine base + styx + dependencies)

## Customization

### Build Custom Image

```dockerfile
# Dockerfile.custom
FROM styx:latest

# Add additional tools
RUN apk add --no-cache make llvm

# Override entry point
ENTRYPOINT ["/bin/bash"]
```

```bash
docker build -f Dockerfile.custom -t styx-extended:latest .
```

### Use Different Base Image

```dockerfile
FROM ubuntu:22.04

# Install build tools
RUN apt-get update && apt-get install -y \
    ca-certificates curl tar gzip bash git

# Copy styx binary (built separately)
COPY styx /usr/local/bin/styx

WORKDIR /workspace
ENTRYPOINT ["styx"]
CMD ["--help"]
```

## Troubleshooting

### Permission Denied

```bash
# Run with host user ID
docker run --rm --user $(id -u):$(id -g) -v $(pwd):/workspace styx:latest install
```

### Container Can't Access Network

```bash
# Use host network (Linux only)
docker run --rm --network host -v $(pwd):/workspace styx:latest install
```

### Volume Mount Issues

```bash
# Use absolute paths
docker run --rm -v $(pwd):/workspace styx:latest install

# Or use docker-compose (handles paths automatically)
docker-compose run --rm styx-dev
```

## Security Considerations

- Container runs as root (no privilege isolation)
- Alpine base minimizes attack surface
- No setuid/setgid binaries in image
- Mount project directory as read-only in CI when possible
- Use ephemeral volumes for CI (no persistent state)

## Building for Multiple Platforms

```bash
# Build for ARM64 (Apple Silicon)
docker buildx build --platform linux/arm64 -t styx:latest .

# Build for both AMD64 and ARM64
docker buildx build --platform linux/amd64,linux/arm64 -t styx:latest .
```

## Pushing to Registry

```bash
# Tag image
docker tag styx:latest killallservers/styx:0.1.0
docker tag styx:latest killallservers/styx:latest

# Push to Docker Hub
docker push killallservers/styx:0.1.0
docker push killallservers/styx:latest
```

## See Also

- [README.md](../README.md) - Main documentation
- [CLAUDE.md](../CLAUDE.md) - Architecture documentation
- [docker-compose.yml](../docker-compose.yml) - Docker Compose configuration
