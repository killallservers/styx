# Local Development: Styx Registry

Test the registry server locally before deploying to production.

---

## Quick Start

### Option 1: Local Binary (Fastest)

Build and run the registry server locally:

```bash
# Build registry server
go build -o bin/styx-registry ./cmd/registry

# Run server (serves specs from pkg/registry/specs/)
./bin/styx-registry -port 7506 -specs pkg/registry/specs

# Test it
curl http://localhost:7506/health
curl http://localhost:7506/registry.json | jq '.tools | length'

# Expected: "26" (tools count)
```

Server runs on `http://localhost:7506`.

---

### Option 2: Docker Compose (Production-Like)

Run the exact production setup locally:

```bash
cd infra/compose

# Build images
docker compose -f compose.dev.yml build

# Start services
docker compose -f compose.dev.yml up -d

# Watch logs
docker compose -f compose.dev.yml logs -f

# Test
curl http://localhost:7506/health
curl http://localhost:7506/registry.json | jq '.tools | length'

# Stop
docker compose -f compose.dev.yml down
```

This runs:
- Registry server (port 7506, internal)
- Caddy reverse proxy (ports 80, 443 with self-signed cert)

---

## Testing Registry Endpoints

### Health Check

```bash
curl -s http://localhost:7506/health | jq
```

Expected response:
```json
{
  "status": "ok",
  "tools": 26,
  "lastBuild": "2026-07-14T17:14:07Z"
}
```

### Full Registry

```bash
curl -s http://localhost:7506/registry.json | jq '.tools | keys'
```

Expected: Array of 26 tool names (ripgrep, fd, bat, golang, node, etc.)

### Specific Tool

```bash
curl -s http://localhost:7506/registry.json | jq '.tools.golang'
```

Expected:
```json
{
  "name": "golang",
  "versions": [
    {
      "version": "1.26.0",
      "platforms": {...},
      ...
    }
  ]
}
```

---

## Testing the Client

The styx CLI client should fetch from the HTTP registry:

```bash
# With registry running locally
export STYX_REGISTRY_URL=http://localhost:7506

styx search golang
styx add golang@1.26.0
styx lock
```

---

## Caddy Testing

If using Docker Compose with Caddy:

```bash
# HTTPS with self-signed cert (dev only)
curl -k https://localhost/registry.json | jq '.tools | length'

# Or via Caddy container
docker exec styx-caddy curl http://localhost:7506/registry.json | jq '.tools | length'
```

---

## Adding Tools to Local Registry

To test adding new tools:

1. Create a new spec file:
   ```bash
   cp pkg/registry/specs/ripgrep.toml pkg/registry/specs/my-tool.toml
   ```

2. Edit the file with your tool details

3. Restart registry:
   ```bash
   # If using binary:
   pkill styx-registry
   ./bin/styx-registry -port 7506 -specs pkg/registry/specs

   # If using Docker:
   docker compose -f infra/compose/compose.dev.yml restart registry
   ```

4. Verify:
   ```bash
   curl http://localhost:7506/registry.json | jq '.tools | keys'
   ```

---

## Testing TOML Spec Validation

The registry validates all specs on load. To test a malformed spec:

```bash
# Edit a spec file to make it invalid
nano pkg/registry/specs/test-tool.toml

# Try to start registry
./bin/styx-registry -port 7506 -specs pkg/registry/specs

# Should fail with validation error
```

---

## Ports

| Service | Port | URL |
|---------|------|-----|
| Registry server | 7506 | http://localhost:7506 |
| Caddy (HTTP) | 80 | http://localhost (dev) |
| Caddy (HTTPS) | 443 | https://localhost (dev, self-signed) |

---

## Logs

### Binary Mode

```bash
# Stdout (where you ran the binary)
# Watch the terminal output
```

### Docker Mode

```bash
# All services
docker compose -f infra/compose/compose.dev.yml logs -f

# Registry only
docker compose -f infra/compose/compose.dev.yml logs -f registry

# Caddy only
docker compose -f infra/compose/compose.dev.yml logs -f caddy
```

---

## Cleanup

```bash
# Binary mode (just Ctrl+C in terminal)

# Docker mode
docker compose -f infra/compose/compose.dev.yml down
docker compose -f infra/compose/compose.dev.yml down -v  # Remove volumes too
```

---

## Troubleshooting

### "Port already in use"

```bash
# Find what's using port 7506
lsof -i :7506
kill -9 <PID>

# Or use different port
./bin/styx-registry -port 7507
```

### TOML parsing error

Check your spec file syntax:
```bash
go run ./cmd/registry -specs pkg/registry/specs
# Will print which spec file failed to parse
```

### Caddy certificate issues

Dev uses self-signed certificates. Ignore warnings:
```bash
curl -k https://localhost/health
# -k skips certificate verification (dev only!)
```

---

## Next Steps

- [Production deployment](DEPLOY.md)
- [GitHub Actions setup](GITHUB_SECRETS_SETUP.md)
