# Styx - Unified Development Environment Manager

**One TOML config, one lock file, identical environments everywhere.**

A unified dev environment manager for Modo Ventures and portfolio companies. Pre-verified tool registry, content-addressable storage, automatic environment loading, and Docker support.

**Status:** v1.0.0 (Production Ready)  
**Platform:** Linux, macOS | **Go:** 1.22+

## Why?

Every Modo portfolio company has this problem:
- Backend dev runs `brew install golang@1.21` locally, but CI has `1.22`
- New hire gets stale `setup.sh` script, spends 2 hours debugging missing tools
- One project needs `postgres@15`, another `postgres@14` in the same repo directory
- Each company reinvents the wheel: Dockerfiles, GitHub Actions matrix, manual version pinning
- Nix is powerful but unfamiliar; mise handles runtimes but not system tools

**Styx** unifies this into one declarative file that works everywhere:

```yaml
# styx.toml
tools:
  ripgrep: "14.1.0"
  fd: "10.0.0"
  bat: "0.24.0"
  eza: "0.18.0"
  postgres: "15.3"
  golang: "1.22.0"
  rust: "1.75.0"
  node: "20.10.0"

env:
  DATABASE_URL: "postgresql://localhost/mydb"
  NODE_ENV: "development"
```

Commit `styx.lock`, and everyone gets the exact same binaries, hashes, environment vars, everywhere.

## The Core Innovation: Pre-Verified Registry + Locked Environments

Instead of each developer/CI system manually discovering how to install tools, Styx:

1. **Registry maintainer (you) curates specs once:**
   ```bash
   # tools/ripgrep.toml (machine-readable install spec)
   # tools/ripgrep.md (reasoning + test results)
   ```

2. **Specs are version-controlled and verified:**
   - Checksum against official releases (automated test)
   - Extracted and tested (`rg --version` must match spec version)
   - Platform matrix validated (linux-x86_64, darwin-arm64, etc.)
   - Human review before merge

3. **Every install is deterministic:**
   ```bash
   styx install          # Reads from embedded registry, installs + locks
   git commit styx.lock  # Team now has reproducible environment
   ```

**Why this works:**
- ✅ Specs are static (no runtime AI dependency)
- ✅ Version-controlled and auditable (`git log styx-registry/`)
- ✅ Reasoning documented for future maintainers
- ✅ Verification CI catches spec errors before they reach developers
- ✅ Lock file = reproducible environment you can commit and share

## Status

**Release:** v1.0.0 (Production Ready)  
**Tech Stack:** Go (single binary), TOML configs, content-addressable storage, SHA256 verification  
**Platform:** Linux, macOS | **Go:** 1.22+  
**Scope:** Binary tools only (no builds); production-tested with Modo portfolio companies  
**Key Features:** Reproducible environments, content-addressed storage, automatic shell loading, Docker support

## Getting Started

```bash
# Clone and build
git clone <styx-repo>
cd styx
go build -o styx ./cmd/styx

# Create a project config
mkdir -p my-project/.styx
cat > my-project/.styx/styx.toml << 'EOF'
[tools]
ripgrep = "14.1.0"
fd = "10.0.0"
golang = "1.21.0"
node = "20.10.0"

[env]
DATABASE_URL = "postgresql://localhost/dev"
NODE_ENV = "development"
EOF

# Install all tools
cd my-project
../styx install

# Generate lock file (commit this)
cat styx.lock

# On another machine or in CI
../styx sync  # Install everything from lock file, verify checksums
```

## What's Included (v1.0.0)

### ✅ Commands (11 shipped)
- `styx install` - Install tools from config
- `styx sync` - Reproducible install from lock file
- `styx lock` - Generate or update lock file
- `styx verify` - Verify tool checksums
- `styx env` - Show environment variables
- `styx update` - Update tool versions
- `styx list` - Show installed tools
- `styx info` - Browse tool metadata
- `styx search` - Find tools on GitHub
- `styx init` - Setup shell auto-loading
- `styx completion` - Generate shell completion

### ✅ Built-in Tools (11)
Utilities: ripgrep, fd, bat, eza, just  
Languages: golang, rust, node, python  
Databases: postgres, redis

### ✅ Features
- Config hierarchy (global → parent → local)
- Lock file reproducibility
- Content-addressable storage (no conflicts)
- Automatic environment loading
- Directory hierarchy (walk-up config discovery)
- Shell completion (bash, zsh, fish)
- Man pages (offline reference)
- Docker support (dev + CI)
- Security audit (SECURITY.md)
- 45+ passing tests
- Multi-registry support with caching

### See Also
- **Getting started:** [QUICKSTART.md](docs/QUICKSTART.md)
- **All commands:** [COMMANDS.md](docs/COMMANDS.md)
- **Configuration:** [CONFIGURATION.md](docs/CONFIGURATION.md)
- **Lock files:** [LOCKFILES.md](docs/LOCKFILES.md)
- **Security:** [SECURITY.md](docs/SECURITY.md)
- **Future plans:** [ROADMAP.md](ROADMAP.md)
- **Development:** [.claude/README.md](.claude/README.md)


## Quick Start

```bash
# Build and install
git clone <styx-repo>
cd styx && go build -o styx ./cmd/styx

# Setup project
cd my-project
mkdir -p .styx
cat > .styx/styx.toml << 'EOF'
[tools]
golang = "1.22.0"
node = "20.0.0"

[env]
DATABASE_URL = "postgresql://localhost/dev"
EOF

# Install and lock
styx install
styx lock
git add .styx/styx.toml styx.lock && git commit -m "Add reproducible dev environment"

# On another machine
styx sync
```

**→ See [QUICKSTART.md](docs/QUICKSTART.md) for full walkthrough**

## Design Decisions

### Embedded Registry, Not Runtime Fetching

**Decision:** Styx ships with a default registry embedded in the binary. No HTTP calls at install time (unless Phase 2).

**Why:** 
- Speeds up installs (no network latency)
- Works offline after first build
- No dependency on external service availability
- Faster for new teammates

**Trade-off:** New tools added to registry require Styx rebuild. Acceptable for MVP (changes infrequently).

### Content-Addressable Storage

**Decision:** Tools stored by SHA256 hash, not by name/version.

```
~/.styx/store/
├── abcd1234ef.../bin/rg        (ripgrep 14.1.0)
├── efgh5678ij.../bin/rg        (ripgrep 14.0.0, different hash)
└── ijkl9012mn.../bin/fd        (fd 10.0.0)
```

Symlinks point from `~/.styx/bin/rg` → hash path, or per-project symlinks.

**Why:**
- Multiple versions coexist without conflicts
- Identical binaries deduplicated automatically
- Bit-for-bit reproducibility
- Safe garbage collection (remove unused hashes)

### Single Lock File per Project

**Decision:** One `styx.lock` per project (or globally), not per-directory.

```bash
~/my-project/
├── .styx/styx.toml       (local overrides, golang@1.21)
├── styx.lock              (global + local merged, locked versions)
```

**Why:**
- Single source of truth
- Easy to commit and review (`git diff styx.lock`)
- Team stays in sync
- Works with `styx sync` for reproducibility

**Future (Phase 3):** Directory hierarchy can override specific tools if needed.

### Config Merging: Local Overrides Global

**Decision:** Project-local config overrides global tool versions, one tool at a time.

```toml
# Global: golang@1.22, node@20, ripgrep@14
# Local: golang@1.21, postgres@15
# Result: golang@1.21, postgres@15, node@20, ripgrep@14
```

**Why:**
- Predictable; no surprises
- Per-tool granularity (not all-or-nothing)
- Matches how developers think about local overrides

### Specs Are Curated, Not Auto-Generated

**Decision:** Registry specs are written by hand (or AI-assisted, then reviewed). No runtime spec generation.

```bash
# Not this:
styx install postgres@latest  # Claude figures out URL at install time

# This:
styx install postgres@15.3    # Spec was pre-computed, curated, committed
```

**Why:**
- No runtime Claude dependency
- Specs are version-controlled (audit trail)
- Maintainer reviews before merge (quality gate)
- Faster installs (no LLM latency)

**Verification CI:** Automated CI verifies specs actually work (download, checksum, --version).

### Stability Markers in Specs

**Decision:** Each tool version is marked for stability (stable, widely-tested, legacy, experimental).

```toml
[versions."14.1.0"]
stability = "stable"

[versions."13.0.0"]
stability = "legacy"
```

**Why:**
- Guides users toward safe versions
- CI can enforce "stable only"
- Transparency about which versions are battle-tested

### Defer Complexity: No Builds, No Windows (Phase 1)

**Decision:** MVP ships binary tools only. No build-from-source. Linux + macOS only.

**Why:**
- Simplicity (80/20 value)
- Binaries are faster, simpler
- Windows introduces platform-specific pain
- Phase 2+ can add system package manager fallback

**Future:** If needed, system PM (`apt install postgres`) as fallback for complex tools.

---


## Comparison

| Feature | Nix | mise | Styx |
|---------|-----|------|------|
| **Learning curve** | Steep (DSL) | Gentle | Minimal (TOML) |
| **System tools** | Yes | No | **Yes** |
| **Runtimes** | Yes | Yes | Yes |
| **Env vars** | Limited | Yes | **Yes** |
| **Lock file** | Yes | No | **Yes** |
| **Reproducibility** | 99% (binaries + builds) | ~85% | ~95% (binaries only) |
| **Works offline** | Yes | Yes | Yes (after first install) |
| **Platform support** | Linux, macOS, minimal Windows | Linux, macOS | Linux, macOS (Phase 1) |
| **Multi-registry** | No | No | **Yes** |
| **Config merging** | Complex | Simple | **Simple** (per-tool) |
| **Install discovery** | Nix packages | Manual | Curated specs |
| **Setup time** | Days | Hours | Minutes |
| **For Modo projects** | Overkill | Close, missing lock | **Just right** |

## Why These Tech Choices

**Go binary:**
- Cross-compile to any platform in seconds
- Single executable (no runtime dependencies)
- Fast startup, memory efficient
- Great CLI ecosystem (cobra)

**TOML config:**
- Human-readable, not a programming language
- Minimal learning curve vs Nix DSL
- Mature tooling (TOML parsing libraries)

**Content-addressable storage by hash:**
- Multiple versions coexist without conflicts
- Bit-for-bit reproducible (same binary = same hash)
- Simple garbage collection (find unused hashes)
- Safe parallelization (no race conditions)

**Registry of curated specs:**
- Pre-computed specs are auditable (git history)
- No runtime AI dependency or latency
- Specs reviewed before merge (quality gate)
- Easier to maintain than 100 install scripts


## License

MIT

---

**One config file. Everything reproducible. No complexity.**

Commit `styx.lock` to git. Your whole team syncs. CI runs identically. Environment management stops being a surprise.

This is the developer experience Nix promised, without the Nix.
