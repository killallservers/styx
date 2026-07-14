# Styx Configuration Guide

Complete reference for configuring Styx projects.

## Overview

Styx uses TOML config files to declare development tools and environment variables.

Two config files work together:

1. **Global** (`~/.local/share/styx/styx.toml`) - Optional system-wide defaults
2. **Local** (`./.styx/styx.toml` or `./styx.toml`) - Project-specific overrides

Local settings override global settings per-tool.

## File Locations

### Global Config

**Path:** `~/.local/share/styx/styx.toml`

**Purpose:** System-wide defaults shared across all projects

**Example:**
```toml
[tools]
ripgrep = "14.1.0"
fd = "10.0.0"
golang = "1.23.1"
node = "20.10.0"

[env]
RUST_BACKTRACE = "1"
LOG_LEVEL = "info"
```

**Usage:** Create manually if you want system defaults. Not required.

### Local Config

**Paths (checked in order):**
1. `.styx/styx.toml` (preferred)
2. `./styx.toml` (fallback)

**Purpose:** Project-specific tool versions and environment

**Example:**
```toml
[tools]
golang = "1.22.0"              # Override: different from global
postgres = "15.3"              # New: not in global
# Inherits from global: ripgrep, fd, node, RUST_BACKTRACE, LOG_LEVEL

[env]
DATABASE_URL = "postgresql://localhost/dev"
API_PORT = "3000"
```

## Config Format

### Tools Section

Declare development tools by name and version.

```toml
[tools]
ripgrep = "14.1.0"
fd = "10.0.0"
bat = "0.24.0"
eza = "0.18.0"
golang = "1.23.1"
rust = "1.81.0"
node = "20.10.0"
python = "3.12.0"
```

**Rules:**
- Tool names are case-sensitive (match registry exactly)
- Versions must exist in embedded registry (or HTTP registry in Phase 2)
- Version strings are semantic (e.g., "1.23.1", not "1.23" or "latest")
- All tools are optional (omit if not needed)

**Available tools (MVP):**
- `ripgrep` - Fast grep
- `fd` - Fast find
- `bat` - Cat with syntax highlighting
- `eza` - Modern ls (replacement for exa)
- `golang` - Go compiler
- `rust` - Rust compiler
- `node` - Node.js runtime
- `python` - Python runtime

Request additional tools via GitHub issues.

### Environment Section

Declare environment variables available to your project.

```toml
[env]
DATABASE_URL = "postgresql://localhost/dev"
NODE_ENV = "development"
RUST_BACKTRACE = "1"
LOG_LEVEL = "info"
DEBUG = "true"
MY_VAR = "my_value"
```

**Rules:**
- Keys must be valid environment variable names (alphanumeric + underscore)
- Values are strings (no quoting required for simple values)
- Special characters in values: quote them (`"value with spaces"`)
- Environment variables are case-sensitive

**Automatic variables (added by Styx):**

For each configured tool, Styx automatically adds `TOOL_NAME_PATH`:

```
RIPGREP_PATH=/home/user/.styx/bin/rg
FD_PATH=/home/user/.styx/bin/fd
GOLANG_PATH=/home/user/.styx/bin/go
RUST_PATH=/home/user/.styx/bin/rustc
NODE_PATH=/home/user/.styx/bin/node
PYTHON_PATH=/home/user/.styx/bin/python
```

View these with `styx env`.

### Comments

Use `#` for comments:

```toml
# Global development tools
[tools]
ripgrep = "14.1.0"  # Fast grep replacement

# Project-specific settings
[env]
DATABASE_URL = "postgresql://localhost/dev"  # Dev database
NODE_ENV = "development"
```

## Merging Behavior

When both global and local configs exist, they merge as follows:

### Tools

Local overrides global per-tool, and adds new tools:

```toml
# Global config
[tools]
golang = "1.23.1"      # ← Global default
node = "20.10.0"       # ← Inherited
ripgrep = "14.1.0"     # ← Inherited

# Local config
[tools]
golang = "1.22.0"      # ← Overrides global
postgres = "15.3"      # ← New tool

# Effective config
golang = "1.22.0"      # From local (overrides)
node = "20.10.0"       # From global (inherited)
ripgrep = "14.1.0"     # From global (inherited)
postgres = "15.3"      # From local (new)
```

### Environment Variables

Local overrides global by key, and adds new variables:

```toml
# Global config
[env]
RUST_BACKTRACE = "1"   # ← Global default
LOG_LEVEL = "info"     # ← Inherited

# Local config
[env]
LOG_LEVEL = "debug"    # ← Overrides global
DATABASE_URL = "postgresql://localhost/dev"  # ← New

# Effective config
RUST_BACKTRACE = "1"   # From global (inherited)
LOG_LEVEL = "debug"    # From local (overrides)
DATABASE_URL = "postgresql://localhost/dev"  # From local (new)
```

## Examples

### Minimal Config

A single tool:

```toml
[tools]
ripgrep = "14.1.0"
```

Result: `ripgrep@14.1.0` installed, no env vars.

### Backend Service

Node.js backend with database:

```toml
[tools]
nodejs = "20.10.0"
postgres = "15.3"

[env]
DATABASE_URL = "postgresql://localhost/dev"
NODE_ENV = "development"
API_PORT = "3000"
```

### Go Project

Go + support tools:

```toml
[tools]
golang = "1.23.1"
ripgrep = "14.1.0"
fd = "10.0.0"

[env]
GOPATH = "${HOME}/.go"
```

### Full Stack (Node + Go)

Multi-language project:

```toml
[tools]
nodejs = "20.10.0"
golang = "1.23.1"
postgres = "15.3"
redis = "7.2.0"

[env]
# Backend (Go)
GOPATH = "${HOME}/.go"

# Frontend (Node)
NODE_ENV = "development"

# Database
DATABASE_URL = "postgresql://localhost/dev"
REDIS_URL = "redis://localhost:6379"

# General
DEBUG = "true"
RUST_BACKTRACE = "1"
```

### Python Project

Python + tools:

```toml
[tools]
python = "3.12.0"
ripgrep = "14.1.0"

[env]
PYTHONUNBUFFERED = "1"
DATABASE_URL = "postgresql://localhost/dev"
```

### Platform-Specific (Not Yet Supported)

Future feature (Phase 2+):

```toml
# Per-platform tool versions
[tools.linux-x86_64]
golang = "1.23.1"

[tools.darwin-arm64]
golang = "1.23.1"

[tools.darwin-x86_64]
golang = "1.22.0"  # Older version for older Intel Macs
```

Currently: Global config applies to all platforms.

## Global + Local Patterns

### Pattern 1: Standardized Global, Project Overrides

**Global (`~/.local/share/styx/styx.toml`):**
```toml
[tools]
ripgrep = "14.1.0"
fd = "10.0.0"
golang = "1.23.1"
node = "20.10.0"

[env]
RUST_BACKTRACE = "1"
```

**Local (`.styx/styx.toml`):**
```toml
[tools]
golang = "1.22.0"  # Project needs older version

[env]
DATABASE_URL = "postgresql://localhost/my_project"
```

**Result:** All tools from global, except golang@1.22.0. Both env vars active.

### Pattern 2: Minimal Global, Mostly Local

**Global:**
```toml
[tools]
ripgrep = "14.1.0"  # Shared across all projects
```

**Local:**
```toml
[tools]
golang = "1.23.1"
node = "20.10.0"
postgres = "15.3"

[env]
DATABASE_URL = "postgresql://localhost/dev"
```

**Result:** Each project declares its own tools + env. Global ripgrep inherited.

### Pattern 3: No Global, Only Local

**No global config exists.**

**Local:**
```toml
[tools]
golang = "1.23.1"
node = "20.10.0"

[env]
DATABASE_URL = "postgresql://localhost/dev"
```

**Result:** Only these tools + vars are available.

## Validation

Styx validates configs before using them. Errors you might see:

### Missing Tool

```
error: tool "foobar" not found in registry
available tools: ripgrep, fd, bat, eza, golang, rust, node, python
```

**Fix:** Check spelling, or request the tool.

### Malformed TOML

```
error: failed to parse config: expected '=', found 'EOF' at line 5
```

**Fix:** Check TOML syntax (missing brackets, quotes, etc.).

### Duplicate Tool

```
error: tool "golang" specified in both global and local config
```

**Note:** This is actually OK—local overrides global. The error might mean something else.

## Workflow: Setup → Commit → Share

### Step 1: Create Local Config

```bash
mkdir -p .styx
cat > .styx/styx.toml << 'EOF'
[tools]
golang = "1.23.1"
node = "20.10.0"

[env]
DATABASE_URL = "postgresql://localhost/dev"
EOF
```

### Step 2: Install

```bash
styx install
```

Outputs:
```
✓ golang@1.23.1 installed to /home/you/.styx/bin/golang
✓ node@20.10.0 installed to /home/you/.styx/bin/node
```

### Step 3: Generate Lock

```bash
styx lock
```

Creates `styx.lock` with exact versions and checksums.

### Step 4: Commit

```bash
git add .styx/styx.toml styx.lock
git commit -m "Set up reproducible dev environment"
git push
```

### Step 5: Team Syncs

Teammate pulls and installs:

```bash
git pull
styx install  # Uses config, verifies against styx.lock
```

Everyone has identical tools and checksums.

## Troubleshooting

### "tool X not found"

Check available tools:

```bash
styx env | grep _PATH
```

If tool is missing, request it via GitHub.

### "version Y not available for tool X"

Check the registry (embedded in binary). Request new versions.

### Config not being read

Styx checks these paths (in order):
1. `.styx/styx.toml`
2. `./styx.toml`
3. `~/.local/share/styx/styx.toml` (global)

Verify the file exists and is readable:

```bash
cat .styx/styx.toml
```

### Merge confusion

See which config is active:

```bash
styx env  # Shows all loaded variables
```

This tells you what's actually merged.

### Accidentally committed secrets

If you committed a secret (API key, password) in `styx.toml`:

```bash
# Remove from git history (careful!)
git rm --cached .styx/styx.toml
echo ".styx/styx.toml" >> .gitignore
git commit -m "Remove styx.toml from git (contains secrets)"

# Edit config to remove secret
vim .styx/styx.toml

# In future, use env var substitution or external secrets tool
```

## Best Practices

1. **Commit `styx.toml` and `styx.lock`** - Team gets identical environments

2. **Keep global config minimal** - Only truly system-wide defaults

3. **Use local config for project-specific settings** - Database URLs, API ports

4. **Don't commit secrets** - Use external secrets management, or git-ignored files

5. **Document unusual versions** - Leave comments if project needs older tool version

   ```toml
   [tools]
   golang = "1.22.0"  # Project requires <1.23 (issue: https://github.com/...)
   ```

6. **Pin exact versions** - Never use ranges like `"1.23.*"` (not supported anyway)

7. **Regular updates** - Review and update tool versions quarterly

   ```bash
   # Update a tool
   sed -i 's/golang = "1.23.1"/golang = "1.24.0"/' .styx/styx.toml
   styx lock
   git diff styx.lock  # Review changes
   git commit -am "Update golang to 1.24.0"
   ```

## Next: Lock Files

Once you have a config, generate a lock file:

```bash
styx lock
```

See [Styx Lock Files](LOCKFILES.md) for details. (Or see COMMANDS.md for `styx lock` reference.)

---

**Config format is intentionally simple: just TOML with `[tools]` and `[env]` sections.**

**One config file per project. One lock file per project. Reproducible environments everywhere.**
