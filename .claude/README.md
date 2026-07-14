# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**📖 Start here:** Read the sections below in order. For complete Claude Code setup and reference materials, see `.claude/` directory structure and contents.

## .claude/ Directory Reference

The `.claude/` directory contains all Claude Code configuration, agents, patterns, and project setup. Start with these files:

### Core Project Definition
- **README.md** (this file) - All developer guidance for working with Styx
- **CLAUDE.md** - You are here; codebase instructions and reference
- **.claude/rules/default.md** - Security rules and permissions

### Custom Agents
See `.claude/agents/` for specialized workers:
- **architect.md** - Design reviews and architecture decisions
- **code-reviewer.md** - Go code review, config/registry/storage audits

### Workflows & Orchestration
See `.claude/workflows/` and `.claude/rules/patterns/` for multi-agent workflows and quality patterns:
- **audit-codebase.js** - Find issues, adversarially verify, report
- **migrate-in-parallel.js** - Parallel refactors with isolated worktrees
- **research-question.js** - Multi-source research with cross-checking
- **loop-until-converged.js** - Iterative discovery with deduplication
- **judge-panel.js** - Multi-angle decision synthesis

### Quality Patterns
See `.claude/rules/patterns/` for reusable orchestration techniques:
- **adversarial-verify.md** - Spawn skeptics to refute findings
- **perspective-diverse.md** - Multi-lens review (correctness, security, performance)
- **dedup.md** - Deduplication against seen set for iterative discovery
- **judge-panel.md** - Draft from angles, judge, synthesize winner
- **cost-aware.md** - Scale to available token budget
- **phase-orchestration.md** - Pipeline (no barrier) vs parallel (barrier)

### Memory & Session Context
See `.claude/memory/` for persistent session memory:
- **MEMORY.md** - Index of remembered facts across sessions
- **\*.md** - Individual memory files (user preferences, feedback, project state)

### Skills & Technology Reference
See `.claude/skills/` for stack-specific guidance (docs on frameworks, libraries, patterns):
- Not included in this project's .claude/ yet; can be added as needed

### Hooks & Automation
See `.claude/hooks/` for automated scripts:
- Currently empty; can add pre-commit, post-save, or other event handlers

---

## Project Overview

**Styx** is a unified dev environment manager for Modo Ventures and portfolio companies. One TOML config file, one lock file, identical environments everywhere—no Nix language, no runtime AI dependency.

**Target:** Internal use first (Modo portfolio), public release after validation  
**Scope:** Linux + macOS only; binary tools only (no builds); internal Modo use first  
**Timeline:** MVP in 2-3 weeks (Phase 1)

The core innovation: pre-verified, registry-based installation specs (stored as `.toml` files) instead of maintaining install scripts or discovering URLs at runtime. Specs are curated and tested once, then embedded in the binary.

**Architecture:**
- **styx/** (this repo) - Single Go binary with embedded registry (8 MVP tools)
- **styx-registry/** (future) - Separate repo for registry specs if we expand beyond embedded

## Essential Commands

### Build & Run
```bash
go build -o styx ./cmd/styx        # Build binary to current dir
./styx install                     # Install from styx.toml (requires registry)
./styx sync --from-lock            # Reproducible install from styx.lock
```

### Testing
```bash
go test ./...                       # Run all tests
go test -bench=. -benchmem ./...    # Run benchmarks
go test ./pkg/config -run TestName  # Single test
go test -v ./...                    # Verbose output
```

### Development
```bash
go mod tidy                         # Update go.mod/go.sum
go fmt ./...                        # Format code
go vet ./...                        # Static analysis
```

### Linting
No linter configured by default; consider `golangci-lint` for CI.

## Architecture & Key Modules

### Project Structure

```
styx/
├── go.mod / go.sum
├── README.md
├── Makefile
├── main.go                    # Entry point
├── cmd/
│   └── styx/
│       ├── install.go         # styx install command
│       ├── lock.go            # styx lock command
│       ├── verify.go          # styx verify command
│       ├── env.go             # styx env command
│       └── main.go
├── pkg/
│   ├── config/
│   │   ├── parser.go          # Parse .styx/styx.toml + ~/.styx/styx.toml
│   │   ├── merger.go          # Merge global + local (local wins)
│   │   └── validator.go       # Validate tool + env keys
│   ├── registry/
│   │   ├── spec.go            # Registry spec data structure
│   │   └── embedded.go        # Embedded default registry
│   ├── installer/
│   │   ├── installer.go       # Main install orchestration
│   │   └── downloader.go      # Download + verify binaries
│   ├── storage/
│   │   ├── store.go           # Content-addressable store (~/.styx/store/{sha256})
│   │   └── symlink.go         # Symlink management
│   ├── platform/
│   │   └── detect.go          # Platform detection + override
│   └── lock/
│       ├── lock.go            # Lock file data structure
│       ├── generator.go       # Generate styx.lock
│       └── parser.go          # Parse styx.lock
├── embedded/
│   └── registry/              # Shipped with binary
│       ├── ripgrep.toml
│       ├── fd.toml
│       ├── bat.toml
│       ├── eza.toml
│       ├── golang.toml
│       ├── rust.toml
│       ├── node.toml
│       └── python.toml
├── tests/
│   ├── config_test.go
│   ├── platform_test.go
│   ├── storage_test.go
│   ├── lock_test.go
│   └── integration_test.go
└── examples/
    ├── simple.styx.toml
    └── with-env.styx.toml
```

### Core Package Structure

**`pkg/config/`** - Config parsing & merging
- `parser.go`: Parse `.styx/styx.toml` (project-local) and `~/.styx/styx.toml` (global)
- `merger.go`: Merge global + local configs, local overrides global per-tool
- `validator.go`: Validate config structure and tool names

**`pkg/registry/`** - Registry specs (Phase 1: embedded only)
- `spec.go`: Registry spec data structure (platform, URL, checksum, executable)
- `embedded.go`: Load embedded default registry from binary

**`pkg/installer/`** - Tool installation orchestration
- `installer.go`: Main orchestration (resolve spec → download → verify → extract → symlink)
- `downloader.go`: Download binaries with parallel support, retry logic, checksum verification

**`pkg/storage/`** - Content-addressable storage
- `store.go`: Manage `~/.styx/store/{sha256}/bin/{tool}` layout
- `symlink.go`: Create/update symlinks to tools, manage PATH integration

**`pkg/platform/`** - Platform detection & normalization
- `detect.go`: Auto-detect `linux-x86_64`, `darwin-arm64`, `darwin-x86_64`; user override support

**`pkg/lock/`** - Lock file generation & parsing
- `lock.go`: Lock file data structure (tools, checksums, registry version, env vars)
- `generator.go`: Generate `styx.lock` from resolved tools
- `parser.go`: Parse and validate `styx.lock`

### Key Design Patterns

**Config Hierarchy:**
```
~/.local/share/styx/styx.toml (global)
  + ./styx.toml (local, per-project)
  = Merged result (local overrides specific tools)
```

**Registry-Based Discovery:**
1. One-time (offline): Registry maintainer uses Claude to generate tool specs
   - Reads GitHub release pages, docs
   - Outputs `tools/ripgrep.toml` (machine-readable spec) + `tools/ripgrep.md` (reasoning)
   - Commits to Styx-registry
2. Every install: Tool reads pre-computed spec from registry (no LLM call at runtime)

**Content-Addressable Storage:**
- Tools stored by version: `~/.styx/store/ripgrep/14.1.0/bin/rg`
- No version conflicts; multiple versions coexist
- Lock file records checksums for integrity

**Lock File Strategy:**
- `styx.lock` captures exact versions, URLs, checksums, registry version used
- Enables reproducible `styx sync --from-lock` on CI or another machine
- Commit to git for team reproducibility

### Installation Methods (Per Tool Spec)

1. **Binary** - Download precompiled binary from GitHub releases (fastest)
2. **Build** - Fetch source, resolve build dependencies, compile
3. **System PM** - Use system package manager (apt, brew, etc.), prompt user for approval

Tool spec `.toml` files list all available methods per version, indexed by platform.

## Development Phases

### Phase 1: Core Config + Registry Foundation (2-3 weeks)
**Ship minimum viable product to Modo portfolio companies.**

**Week 1: Core Infrastructure**
- Go project setup (cobra CLI)
- TOML config parser + tests
- Config merging logic (global + local override)
- Platform detection (linux-x86_64, darwin-arm64, darwin-x86_64)
- Content-addressable storage (write, read, verify by SHA256)
- Symlink management
- **Tests:** config parsing, platform detection, storage ops

**Week 2: Install + Lock**
- Embedded registry (8 MVP tools binary specs)
- `styx install` command (resolve → download → verify → symlink)
- Lock file generation + parsing
- `styx verify` command (checksum verification)
- `styx lock` command (regenerate from config)
- Error handling + helpful messages
- **Tests:** Parallel download, lock file integrity, end-to-end

**Week 3: Polish + Dogfood**
- `styx env` command (show loaded env vars)
- Documentation + examples
- Test on 2-3 Modo projects
- Spec verification CI (download, checksum, --version)
- Edge case fixes from real usage

**MVP Commands:** `styx install`, `styx lock`, `styx verify`, `styx env`

**Definition of Done:**
```bash
styx install           # Works locally (Linux + macOS)
cat styx.lock          # Valid JSON
styx verify            # Checksums match
styx env               # Shows loaded variables
styx sync              # Installs from lock on fresh machine
```

### Phase 2: Registry Growth & Polish (4-8 weeks)
**Expand toolkit based on Modo project needs.**
- Registry fetching (HTTP fallback to embedded)
- More tool specs (postgres, redis, make, etc.)
- `styx update` for selective version updates
- Performance optimization (parallel downloads)
- System PM fallback (optional: apt/brew if binary unavailable)

### Phase 3: Environment Loading & Shell Integration (8-12 weeks)
- Load env vars from config + lock
- Walk directory tree for hierarchical overrides
- Shell spawning with loaded environment
- direnv compatibility

### Phase 4: Polish & Distribution
- GitHub Actions integration
- GoReleaser + Homebrew formula
- Man pages, shell completion (bash, zsh, fish)
- Performance profiling
- Security audit

### Not in Scope (Indefinitely)
- Windows support
- Build-from-source (too complex; use system PM if needed)
- Nix-style full environment reproducibility
- Multi-registry complexity

**Phase 5+ (Polish, CI/CD, Discovery):**
- Multiple registries, caching strategies
- GitHub Actions integration, Docker support
- Tool discovery (`styx search`, `styx info`)
- Distribution (Homebrew, GoReleaser)

## Config Format

### Global Config
```toml
# ~/.local/share/styx/styx.toml
[[registries]]
url = "github:styx-community/styx-registry"
version = "latest"

[tools]
golang = "1.22.0"
node = "20.10.0"

[env]
RUST_BACKTRACE = "1"

[shells]
default = ["golang", "node"]
```

### Local Override
```toml
# ./styx.toml (in project)
[tools]
golang = "1.21.0"           # Override global
postgres = "15.3"           # New tool

[env]
DATABASE_URL = "..."        # Merged with global
```

Result: golang@1.21.0 (local), postgres@15.3 (local), node@20.10.0 (inherited from global).

### Tool Spec (In Registry)
```toml
# styx-registry/tools/ripgrep.toml
name = "ripgrep"
repository = "github:BurntSushi/ripgrep"

[versions."14.1.0"]
released = "2024-11-20"
stability = "stable"

[[versions."14.1.0".methods]]
type = "binary"
platforms.linux-x86_64 = "ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz"
[versions."14.1.0".methods.checksums]
"linux-x86_64" = "abcd1234..."

[versions."14.1.0".llm_reasoning]
timestamp = "2026-01-15T14:30:00Z"
model = "claude-opus-4.6"
reasoning = """
GitHub releases provide official musl binaries for all platforms...
"""
```

### Lock File
```json
{
  "version": "1.0",
  "generated_at": "2026-01-12T14:30:00Z",
  "registry_snapshot": {
    "url": "github:styx-community/styx-registry",
    "version": "2026-01-15"
  },
  "tools": [
    {
      "name": "ripgrep",
      "version": "14.1.0",
      "install_method": "binary",
      "storage_path": "~/.styx/store/ripgrep/14.1.0",
      "binary_hash_sha256": "abcd1234...",
      "executable": "rg",
      "source_config": "global"
    }
  ],
  "env": { "RUST_BACKTRACE": "1" }
}
```

## CLI Commands (Reference)

Implemented or planned:
- `styx install [tool]` - Discover & download from registry specs
- `styx sync` - Reproducible install from lock file
- `styx lock` - Generate lock file
- `styx verify` - Verify checksums against lock
- `styx update [tool]` - Update tool(s) to latest
- `styx env` - Show loaded env vars
- `styx shell` - Start shell with env loaded
- `styx list` - Show installed tools + versions
- `styx search <tool>` - Find tool on GitHub
- `styx info <tool>` - Show metadata (versions, platforms, stability)

## Integration Points

**LLM (One-time Registry Generation, Offline)**
- Claude: Read GitHub release pages → output TOML specs + reasoning
- Registry maintainer: Review, test, commit specs
- Result: Deterministic installs from pre-computed specs (no runtime LLM cost)

**Platform Detection**
- Auto-detect: `linux-x86_64`, `darwin-arm64`, etc.
- Allow user override via config
- Fail gracefully if binary not available

**Dependency Resolution**
- Recursive: Tool X needs library Y → check store → check system PM → prompt user
- Build: Resolve build dependencies transitively
- Lock: Record all resolved dependencies

**Storage & Verification**
- SHA256 checksums in specs + lock file
- Verify downloads before extraction
- Clean up incomplete/corrupted downloads

## Testing Strategy

- **Unit tests**: Config parsing, merging, platform detection, storage ops
- **Integration tests**: Registry fetching, downloads, environment loading
- **Fixtures**: Mock GitHub releases, test tool configs
- **Benchmarks**: Install speed, config merge performance

Key areas needing tests:
- Config merging edge cases (multiple overrides, inheritance)
- Platform matching (fuzzy name matching for binaries)
- Checksum verification (corrupted files, retries)
- Lock file reproducibility (same inputs → same lock)

## Technical Challenges

1. **LLM Registry Generation** - Claude must reliably parse GitHub release formats; output must be valid TOML
2. **Platform Mapping** - Normalize platform names across tools (`x86_64-linux` vs `x86-64-unknown-linux-musl`)
3. **Dependency Resolution** - Handle circular/conflicting versions gracefully
4. **Shell Integration** - Support bash, zsh, fish, PowerShell with consistent behavior
5. **Security** - Verify checksums, detect supply chain compromise
6. **Performance** - Parallel downloads, lazy evaluation of specs, efficient caching

## Notable Design Decisions

**Embedded Registry, Not Runtime Fetching**
- Styx ships with default registry embedded in binary (8 MVP tools)
- No HTTP calls at install time (Phase 2 adds optional HTTP)
- Speeds up installs, works offline, no external service dependency
- Trade-off: new tools require Styx rebuild (acceptable for MVP)

**Content-Addressable Storage**
- Tools stored by SHA256 hash: `~/.styx/store/{sha256}/bin/{tool}`
- Multiple versions coexist without conflicts
- Identical binaries deduplicated automatically
- Bit-for-bit reproducibility, safe garbage collection

**Single Lock File per Project**
- One `styx.lock` per project, not per-directory
- Single source of truth, easy to commit and review
- Team stays in sync, works with `styx sync` for reproducibility
- Future (Phase 3): directory hierarchy can override specific tools if needed

**Config Merging: Local Overrides Global**
- Project-local config overrides global tool versions, one tool at a time
- Example: global has golang@1.22 + node@20, local specifies golang@1.21 + postgres@15
- Result: golang@1.21, postgres@15, node@20 (inherited)
- Predictable, per-tool granularity, matches developer expectations

**Specs Are Curated, Not Auto-Generated**
- Registry specs written by hand (or AI-assisted, then reviewed), not runtime-generated
- No runtime Claude dependency
- Specs pre-verified (download, checksum, test `--version`)
- Verification CI catches errors before developers use them

**Linux + macOS Only, Binary Tools Only**
- Windows support deferred (not in scope)
- Build-from-source deferred (too complex; use system PM if needed)
- Focus on binary distribution for speed and simplicity

## Common Tasks

**Adding a new tool to registry:**
1. Run Claude: "Generate spec for `newlib@X.Y.Z`" (reads GitHub, outputs TOML + reasoning)
2. Review & test the generated spec
3. Commit `tools/newlib.toml` + `tools/newlib.md` to styx-registry

**Debugging install failures:**
- Check: `styx list` (what's installed?)
- Check: Registry spec correctness (platforms, checksums, URLs)
- Check: Platform detection (`styx status` shows detected platform)
- Check: Lock file (expected versions, checksums match?)

**Improving config merging:**
- Ensure local tool versions override global (not append)
- Test inheritance: local doesn't specify tool X → should use global
- Test edge cases: empty local, empty global, multiple overrides

**Performance optimization:**
- Profile with `go test -bench=. -benchmem`
- Focus: config parsing, platform detection, storage ops
- Consider lazy evaluation for large tool sets

## Agents, Workflows, and Patterns

Custom agents, reusable workflows, and quality patterns for orchestrating multi-agent work.

### Available Agents

- **architect** - Systems architect for Styx design reviews and architecture decisions
- **code-reviewer** - Senior code reviewer for Go implementations, config parsing, registry clients, installer logic

See `.claude/agents/` for detailed configuration.

### Workflow Templates

Reusable orchestration scripts for common multi-agent patterns. Stored in `.claude/workflows/`.

**audit-codebase.js** - Find issues, verify findings, report confirmed results
```
Pattern: Find → Audit in parallel → Verify each finding → Report confirmed
Use when: Security audits, compliance checks, code standards sweeps
```

**migrate-in-parallel.js** - Discover files, transform each in isolated worktree, verify
```
Pattern: Find → Migrate each in isolated tree → Verify migrations
Use when: Refactors, library upgrades, large API changes
```

**research-question.js** - Fan out web searches, fetch sources, cross-check, synthesize
```
Pattern: Search (3 angles) → Fetch sources → Verify claims → Synthesize
Use when: Competitive analysis, trend research, tech evaluation
```

**loop-until-converged.js** - Iteratively find issues, dedup, stop when dry
```
Pattern: Search → Dedup → Loop until dry
Use when: Flaky test discovery, bug sweeps, finding edge cases
```

**judge-panel.js** - Draft solutions from different angles, judge each, synthesize
```
Pattern: Draft (MVP/Risk/Cost/User angles) → Judge each → Synthesize
Use when: Architecture decisions, design choices, trade-off analysis
```

### Quality Patterns

Reusable orchestration techniques that encode best practices. Stored in `.claude/rules/patterns/`.

**Adversarial Verification** - Spawn N skeptics to refute each finding; report only if majority survives
- **Best for:** High-stakes findings (security, breaking changes)
- **Cost:** 3× per finding
- **File:** `patterns/adversarial-verify.md`

**Perspective-Diverse Review** - Each reviewer examines via different lens (correctness, security, perf)
- **Best for:** Multi-faceted quality (code, design, architecture)
- **Cost:** N lenses per finding
- **File:** `patterns/perspective-diverse.md`

**Deduplication Against Seen Set** - Maintain Set<key>, filter out already-processed items
- **Best for:** Iterative discovery (loop-until-dry)
- **Cost:** Negligible (O(1) per check)
- **File:** `patterns/dedup.md`

**Judge Panel with Synthesis** - Draft from N angles, judge each, combine winner + best ideas
- **Best for:** Decisions with no obvious right answer
- **Cost:** N drafts + N judges + synthesis
- **File:** `patterns/judge-panel.md`

**Cost-Aware Scaling** - Scale agent count and rounds based on available token budget
- **Best for:** Open-ended discovery with budget constraints
- **Cost:** Adapts to user's token limit
- **File:** `patterns/cost-aware.md`

**Phase-Based Orchestration** - Use `pipeline()` (no barrier) for independent stages, `parallel()` (barrier) for cross-item logic
- **Best for:** Structuring multi-stage workflows
- **Cost:** Optimizes wall-clock time
- **File:** `patterns/phase-orchestration.md`

### When to Use Workflows vs Subagents

| | Subagents | Workflows |
|---|-----------|-----------|
| **What** | Single specialized worker | Multi-agent orchestration script |
| **Who decides next step** | Claude, turn-by-turn | Script, executing deterministically |
| **Results** | Land in context | Stay in script variables |
| **Scale** | Few per turn | Dozens to hundreds |
| **Use when** | Offloading a side task | Building repeatable multi-phase work |

**Example usage:**
- ✅ Audit 100 files for issues → use workflow (parallelism, deterministic phases)
- ✅ Research a question across sources → use workflow (multi-phase, cross-check)
- ❌ Quick code review → use subagent (simpler, context stays in session)
