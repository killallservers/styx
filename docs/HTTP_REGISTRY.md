# HTTP Registry Specification (Phase 2)

## Overview

The HTTP Registry enables remote tool management without requiring binary rebuilds. Styx fetches specs from a central registry server, with intelligent fallback to embedded specs for offline reliability.

## Architecture

```
HTTP Registry Server (registry.styx.sh)
    ↓
registry.json (full registry export)
    ↓ (client requests)
Styx client (with fallback chain)
    ├─ Try: HTTP fetch (24hr cache validation)
    ├─ Try: Fresh cache (< 24 hours)
    ├─ Try: Stale cache (any age, with warning)
    └─ Try: Embedded registry (always succeeds)
```

## JSON Format

The HTTP registry serves `registry.json` with the following structure:

```json
{
  "version": "1.0.0",
  "generated": "2026-07-14",
  "tools": {
    "golang": {
      "name": "golang",
      "repository": "github:golang/go",
      "versions": {
        "1.26.0": {
          "released": "2024-11-01",
          "stability": "stable",
          "methods": [
            {
              "type": "binary",
              "platforms": {
                "linux-x86_64": "go1.26.0.linux-amd64.tar.gz",
                "darwin-arm64": "go1.26.0.darwin-arm64.tar.gz",
                "darwin-x86_64": "go1.26.0.darwin-amd64.tar.gz"
              },
              "checksums": {
                "linux-x86_64": "e8f0a1b2c3d4e5f6...",
                "darwin-arm64": "d7e8f0a1b2c3d4...",
                "darwin-x86_64": "e8f0a1b2c3d4e5..."
              },
              "executable": "go"
            }
          ]
        },
        "1.25.0": { ... }
      }
    },
    "kubectl": { ... },
    "terraform": { ... }
  }
}
```

## Fallback Chain

### 1. HTTP Fetch (Fresh)
- **Trigger:** On `styx install`, `styx add`, etc.
- **Request:** `GET https://registry.styx.sh/registry.json`
- **Timeout:** 30 seconds
- **Cache:** Save to `~/.styx/cache/registry.cache.json`
- **Success:** Use fetched specs

### 2. Cache Check (Fresh)
- **File:** `~/.styx/cache/registry.cache.json`
- **Freshness:** Modification time < 24 hours
- **Use when:** HTTP fetch fails, but cache exists and is fresh
- **Success:** Use cached specs (no network call)

### 3. Cache Fallback (Stale)
- **File:** `~/.styx/cache/registry.cache.json`
- **Age:** Any age
- **Warning:** Print to stderr: "Using stale cached registry (fresh fetch failed: ...)"
- **Use when:** HTTP fails and cache is stale (old)
- **Success:** Use stale cache (better than nothing)

### 4. Embedded Fallback
- **Specs:** Baked into binary at compile time
- **Use when:** All network/cache attempts fail
- **Always succeeds** (no network dependency)

## Migration Path (v1.0 → v1.1 → v2.0)

### v1.0.0 (Current)
```
Embedded only
└─ No HTTP, no cache
```

### v1.1.0 (This Session)
```
Foundation for HTTP:
├─ styx registry export (generates registry.json)
├─ HTTP fallback code (skeleton)
└─ Cache infrastructure (partial)
```

### v2.0.0 (Next)
```
Full HTTP support:
├─ HTTP fetch with timeout
├─ Cache validation (24hr freshness)
├─ Fallback chain
├─ registry.styx.sh deployment
└─ Documentation for self-hosting
```

## Server Setup (Phase 2+)

### Option 1: Official Server (registry.styx.sh)
- Hosted by Modo Ventures
- Automated updates from repo releases
- Served via CloudFront CDN
- Built-in status dashboard

### Option 2: Self-Hosted
Users can run their own registry server:

```bash
# 1. Build custom registry
styx registry export > registry.json

# 2. Serve static file
python3 -m http.server 8000 --bind 0.0.0.0

# 3. Configure client
[[registries]]
url = "http://internal-registry.company.com"
version = "1.0"
```

## Adding Tools to HTTP Registry

### For Official Server
1. Add TOML spec to repo: `pkg/registry/specs/newtool.toml`
2. Submit PR with spec
3. Repo maintainer merges PR
4. Next release rebuilds registry.json
5. `registry.styx.sh` updates (CDN invalidates cache)
6. All clients automatically get the tool in next sync

### For Self-Hosted
Same process, but on your own registry:
```bash
# Clone official specs repo
git clone https://github.com/killallservers/styx-registry
cd styx-registry

# Add or modify specs
cp specs/newtool.toml specs/

# Generate and serve
styx registry export > registry.json
# (upload to server)
```

## Cache Management

### Cache Location
```
~/.styx/cache/registry.cache.json
```

### Cache Lifetime
- **Valid:** < 24 hours old
- **Stale:** ≥ 24 hours old (usable, but warning shown)
- **Refreshed:** Automatically on successful HTTP fetch

### Manual Cache Clear
```bash
# Clear all caches
rm -rf ~/.styx/cache/registry*

# Next styx command will fetch fresh or use embedded
```

## Spec Format for HTTP

Registry JSON uses the same format as TOML specs (via struct tags):

```go
type ToolSpec struct {
  Name       string                `json:"name"`
  Repository string                `json:"repository"`
  Versions   map[string]VersionSpec `json:"versions"`
}

type VersionSpec struct {
  Released  string           `json:"released"`
  Stability string           `json:"stability"`
  Methods   []InstallMethod  `json:"methods"`
}

type InstallMethod struct {
  Type       string            `json:"type"`
  Platforms  map[string]string `json:"platforms"`
  Checksums  map[string]string `json:"checksums"`
  Executable string            `json:"executable"`
}
```

## Validation

HTTP registry specs use the same validation as embedded:

```go
ValidateToolSpec(spec) // 8 checks:
  1. Tool has name
  2. Tool has versions
  3. All versions have methods
  4. All methods have platforms
  5. All methods have checksums
  6. Platforms and checksums match exactly
  7. All executables specified
  8. Checksums are valid SHA256
```

Any spec failing validation is rejected with error message naming the tool and issue.

## Security Considerations

### Checksum Verification
- All binaries verified against SHA256 checksums in registry
- Tampering detected immediately at install time
- No version of a tool is installable without matching checksum

### HTTPS-Only (Phase 2+)
- All HTTP requests use HTTPS only
- No fallback to HTTP (enforced by client)
- Registry URL must be HTTPS

### Signature Verification (Phase 3+)
Future enhancement:
- Registry JSON signed with Ed25519
- Client verifies signature before using specs
- Protection against server compromise

## Deployment Timeline

| Phase | Status | Features |
|-------|--------|----------|
| v1.0 | ✓ Complete | Embedded only |
| v1.1 | 🔄 This session | Foundation + export + list commands |
| v2.0 | 📅 Next | Full HTTP support, registry.styx.sh |
| v2.1 | 📅 Future | HTTPS + signatures |
| v3.0 | 📅 Planned | Windows support, more platforms |

## Testing

### Unit Tests (Needed)
- JSON parsing
- Fallback chain logic
- Cache validation
- Spec validation

### Integration Tests (Needed)
- Fetch from mock HTTP server
- Cache behavior (fresh, stale, missing)
- Fallback to embedded
- No network scenario

### Performance Benchmarks
- Fresh cache hit: < 10ms
- Cache miss (embedded): < 5ms
- HTTP fetch (worst case): < 30s + fallback

## Monitoring

### Metrics to Track
- HTTP fetch success rate
- Cache hit rate (fresh vs stale vs miss)
- Fallback frequency
- Tools downloaded per day
- Most popular tools

### Health Checks
- `registry.styx.sh` availability
- CDN cache hit ratio
- JSON validity (schema validation)
- Checksum mismatches

## Future Enhancements

### Planned (Post v2.0)
1. **Registry UI** - Browse tools, versions, docs
2. **Audit logging** - Track who installed what when
3. **Analytics** - Trending tools, usage patterns
4. **Tool ratings** - Community feedback on tools
5. **Custom registries** - Organization-specific tools
6. **Registry federation** - Link multiple registries

### Proposed (Post v3.0)
1. **Build caching** - Pre-compiled binary variants
2. **A/B testing** - Canary deployments of tools
3. **Staged rollout** - Beta versions before stable
4. **Compliance checking** - License/security scanning
5. **Integration with package managers** - apt, homebrew, cargo

## References

- [REGISTRY.md](REGISTRY.md) - TOML spec format
- [Code](../pkg/registry/) - Implementation
- [Config](../pkg/config/) - Config loading
