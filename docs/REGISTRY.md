# Styx Registry

The Styx registry is a declarative collection of tool specifications that define how to download, verify, and install development tools.

## Registry Architecture

```
pkg/registry/specs/         ← Source of truth (TOML files)
    ├── ripgrep.toml
    ├── golang.toml
    ├── node.toml
    └── ... (16 tools total)
    
    ↓ (compile time)
    
Go embed.FS                 ← Baked into binary
    
    ↓ (runtime)
    
LoadEmbeddedRegistry()      ← Parse TOML → ToolSpec objects
    
    ↓
    
map[string]*ToolSpec       ← In-memory registry
```

## Tool Spec Format

Each tool has a TOML file defining name, repository, and versions.

### Example: `golang.toml`

```toml
name = "golang"
repository = "github:golang/go"

[[versions]]
version = "1.26.0"
released = "2024-11-01"
stability = "stable"

[versions.methods.binary]
type = "binary"
executable = "go"

[versions.methods.binary.platforms]
linux-x86_64 = "go1.26.0.linux-amd64.tar.gz"
darwin-arm64 = "go1.26.0.darwin-arm64.tar.gz"
darwin-x86_64 = "go1.26.0.darwin-amd64.tar.gz"

[versions.methods.binary.checksums]
linux-x86_64 = "e8f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d"
darwin-arm64 = "d7e8f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6"
darwin-x86_64 = "e8f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c"
```

### Schema

**Top-level fields:**
- `name` (string, required) - Tool identifier (matches filename without .toml)
- `repository` (string, required) - Repository reference (e.g., `github:golang/go`)

**Versions array:**
- `version` (string, required) - Semantic version (e.g., `1.26.0`)
- `released` (string, required) - Release date (ISO 8601: `YYYY-MM-DD`)
- `stability` (string) - Stability level: `stable`, `lts`, `beta` (default: `stable`)

**Methods table (supports multiple install methods):**
- Currently only `binary` is supported
- Key becomes method name (e.g., `[versions.methods.binary]`)

**Method fields:**
- `type` (string, required) - Install method type: `binary`
- `executable` (string, required) - Binary name on disk (e.g., `go`, `rg`, `node`)
- `platforms` (table, required) - Platform → filename mapping
- `checksums` (table, required) - Platform → SHA256 checksum mapping

**Supported platforms:**
- `linux-x86_64` - Linux on x86-64
- `darwin-arm64` - macOS on Apple Silicon
- `darwin-x86_64` - macOS on Intel

## Multiple Versions

Tools can define multiple versions. Later versions (higher in file) override earlier ones for "latest" detection.

```toml
[[versions]]
version = "1.26.0"
# ... latest version (detected first by styx add)

[[versions]]
version = "1.25.0"
# ... stable LTS version

[[versions]]
version = "1.22.0"
# ... older version
```

## Adding a New Tool

### 1. Create spec file

Create `pkg/registry/specs/<toolname>.toml`:

```toml
name = "example"
repository = "github:user/example"

[[versions]]
version = "1.0.0"
released = "2024-11-01"
stability = "stable"

[versions.methods.binary]
type = "binary"
executable = "example"

[versions.methods.binary.platforms]
linux-x86_64 = "example-1.0.0-linux-x86_64.tar.gz"
darwin-arm64 = "example-1.0.0-arm64-apple-darwin.tar.gz"
darwin-x86_64 = "example-1.0.0-x86_64-apple-darwin.tar.gz"

[versions.methods.binary.checksums]
linux-x86_64 = "abcd1234..."
darwin-arm64 = "efgh5678..."
darwin-x86_64 = "ijkl9012..."
```

### 2. Get real checksums

```bash
# Download binaries and compute SHA256
sha256sum example-1.0.0-linux-x86_64.tar.gz
# abcd1234... example-1.0.0-linux-x86_64.tar.gz
```

### 3. Build and test

```bash
go build -o styx ./cmd/styx/
styx add example@1.0.0
```

### 4. Commit

```bash
git add pkg/registry/specs/example.toml
git commit -m "feat: add example tool to registry"
```

## Validation

The registry loader validates all specs on startup:

```go
ValidateToolSpec(spec) // Checks:
  - Tool has name and versions
  - All versions have methods
  - All methods have platforms and checksums
  - Platforms and checksums match exactly
  - All executables are specified
```

Validation errors are fatal and reported at build time if using Go build, or at load time if using HTTP registry.

## Future Extensions

### Phase 2: HTTP Registry

Same TOML format served as JSON at `https://registry.styx.sh/registry.json`:

```bash
# Fallback chain:
1. Try HTTP fetch (with 24hr cache)
2. Try fresh cache (< 24 hours old)
3. Try stale cache (any age)
4. Fall back to embedded (always succeeds)
```

### Phase 2+: Additional Install Methods

Future support for:
- `apt` / `homebrew` - Package managers
- `cargo install` - Rust projects
- `source` - Build from source
- `docker` - Container images

Each would add a new method section:

```toml
[versions.methods.apt]
type = "apt"
package = "ripgrep"

[versions.methods.cargo]
type = "cargo"
crate = "ripgrep"
```

## Current Registry (16 tools, 20+ versions)

### Development Tools
- **ripgrep** (14.1.0) - Fast text search
- **fd** (10.1.0) - Fast file finder
- **bat** (0.24.0) - Syntax highlighting cat
- **eza** (0.18.0) - Modern ls
- **just** (1.25.2) - Command runner
- **git** (2.47.0, 2.45.0) - Version control
- **curl** (8.10.0, 8.6.0) - HTTP client
- **jq** (1.7.1, 1.6.0) - JSON processor
- **tmux** (3.4, 3.3) - Terminal multiplexer

### Runtimes
- **golang** (1.26.0, 1.25.0, 1.22.0) - Go runtime
- **node** (20.10.0, 18.17.0) - JavaScript/Node.js
- **python** (3.12.0, 3.11.0) - Python
- **rust** (1.75.0) - Rust toolchain

### Databases & Services
- **postgres** (15.3) - PostgreSQL
- **redis** (7.2.0) - Redis cache
- **docker** (27.0.0, 25.0.0) - Container runtime

## Commands That Use Registry

| Command | Purpose | Uses Registry |
|---------|---------|---------------|
| `styx add` | Add tool to config | ✓ Lookup + validate |
| `styx install` | Install all tools | ✓ Lookup + download |
| `styx update` | Update version | ✓ Lookup + validate |
| `styx lock` | Generate checksums | ✓ Lookup + resolve |
| `styx info` | Tool metadata | ✓ Browse versions |
| `styx search` | Find on GitHub | ✗ (GitHub API) |

## Design Principles

1. **Declarative** - Specs describe what, not how
2. **Embeddable** - Baked into binary at compile time
3. **Validating** - Errors caught early, not at install time
4. **Versionable** - Full history in Git
5. **Extensible** - Foundation for HTTP + package managers
6. **Maintainable** - One TOML file per tool (easy to read, easy to review)

## Contributing

1. Pick a popular tool not yet in registry
2. Create TOML spec following the schema
3. Test locally: `go build && styx add <tool>`
4. Open PR with new spec file
5. Maintainer builds + releases new binary

Registry contributions are low-friction: just add a TOML file, no Go code changes needed.
