# Styx Commands Reference

## styx install

Install all tools from configuration.

### Usage

```bash
styx install [tools...]
```

### Description

Resolves tools from the embedded registry, downloads binaries, verifies checksums, stores in content-addressable storage, and creates symlinks.

Reads from global config (`~/.local/share/styx/styx.toml`) merged with local config (`./.styx/styx.toml` or `./styx.toml`).

### Examples

**Install all configured tools:**
```bash
styx install
```

**Install specific tools only:**
```bash
styx install golang ripgrep
```

### Output

```
✓ ripgrep@14.1.0 installed to /home/erik/.styx/bin/ripgrep
✓ golang@1.23.1 installed to /home/erik/.styx/bin/golang
✓ node@20.10.0 installed to /home/erik/.styx/bin/node
```

### Notes

- Creates `~/.styx/store/`, `~/.styx/cache/`, `~/.styx/bin/` if missing
- Downloads are cached in `~/.styx/cache/`
- Binaries stored by SHA256 hash
- Symlinks created in `~/.styx/bin/` pointing to store
- Generates `styx.lock` on success

---

## styx lock

Generate or update lock file.

### Usage

```bash
styx lock [flags]
```

### Flags

- `--from-lock` - Reproduce from existing lock file (not yet implemented)

### Description

Reads configuration, installs all tools (resolving versions and checksums), and writes deterministic `styx.lock` file.

The lock file can be committed to git for team reproducibility.

### Examples

**Generate lock file:**
```bash
styx lock
```

Output:
```
✓ Lock file generated: styx.lock (3 tools)
```

### Output Format

Lock file is JSON with this structure:

```json
{
  "version": "1.0",
  "generated_at": "2026-07-13T20:00:00Z",
  "registry_snapshot": {
    "url": "embedded",
    "version": "0.1.0"
  },
  "tools": [
    {
      "name": "ripgrep",
      "version": "14.1.0",
      "install_method": "binary",
      "storage_path": "~/.styx/store/abcd1234.../bin/rg",
      "binary_hash_sha256": "abcd1234...",
      "executable": "rg",
      "source_config": "global"
    }
  ],
  "env": {
    "DATABASE_URL": "postgresql://localhost/dev"
  }
}
```

### Usage in CI/CD

Commit `styx.lock` to git:

```bash
git add styx.lock .styx/styx.toml
git commit -m "Update dev environment"
```

Then use `styx install` with lock file (or `styx sync` when implemented).

---

## styx verify

Verify installed tools match checksums in lock file.

### Usage

```bash
styx verify [tools...]
```

### Description

Loads `styx.lock`, calculates SHA256 hashes of installed binaries, and compares against recorded checksums.

Useful for detecting corrupted downloads, tampered binaries, or incomplete installations.

### Examples

**Verify all tools:**
```bash
styx verify
```

Output:
```
✓ All 3 tools verified successfully
```

**Verify specific tools:**
```bash
styx verify golang ripgrep
```

### Error Output

If verification fails:

```
✗ Verification failed:
  - golang: checksum mismatch (expected: abc123..., got: def456...)
  - ripgrep: file not found
```

### Notes

- Requires `styx.lock` to exist
- Checks each tool's binary against SHA256 in lock file
- Reports mismatches clearly
- Does NOT fix mismatches (use `styx install` to reinstall)

---

## styx env

Show loaded environment variables and tool paths.

### Usage

```bash
styx env [variables...]
```

### Description

Loads and merges global + local configs, displays all resolved environment variables.

Automatically adds `TOOL_NAME_PATH` variables for each tool (e.g., `RIPGREP_PATH=/home/erik/.styx/bin/rg`).

### Examples

**Show all environment variables:**
```bash
styx env
```

Output:
```
DATABASE_URL=postgresql://localhost/dev
GOLANG_PATH=/home/erik/.styx/bin/golang
NODE_ENV=development
NODE_PATH=/home/erik/.styx/bin/node
RIPGREP_PATH=/home/erik/.styx/bin/ripgrep
RUST_BACKTRACE=1
```

**Show specific variables:**
```bash
styx env GOLANG_PATH DATABASE_URL
```

Output:
```
DATABASE_URL=postgresql://localhost/dev
GOLANG_PATH=/home/erik/.styx/bin/golang
```

**Use in shell:**
```bash
# Print as export statements
styx env | grep PATH

# Set in current shell (manually)
eval "export $(styx env | grep DATABASE_URL)"
```

### Notes

- Displays in sorted alphabetical order (KEY=VALUE format)
- If lock file exists, uses actual executable names from there
- Otherwise uses configured tool names
- Filters by variable name if arguments provided
- Returns nothing if no variables found

---

## styx sync

Reproducible install from lock file ✅ (Implemented in Phase 2)

### Usage

```bash
styx sync [tools...]
```

### Description

Installs all tools exactly as specified in `styx.lock` file.

Reproduces the exact environment recorded in lock:
- Same tool versions
- Same checksums (SHA256 verified)
- Bit-for-bit reproducible across machines

Essential for CI/CD pipelines and new team members getting identical environments.

### Examples

**Install all tools from lock:**
```bash
styx sync
```

Output:
```
Installing ripgrep@14.1.0... ✓
Installing golang@1.23.1... ✓
Installing node@20.10.0... ✓

✓ Synced 3 tools from lock file
```

**Sync specific tools:**
```bash
styx sync golang ripgrep
```

### Usage in CI/CD

**GitHub Actions:**
```yaml
- name: Setup environment
  run: |
    go build -o styx ./cmd/styx
    ./styx sync
    export PATH="$HOME/.styx/bin:$PATH"
```

### Notes

- Requires `styx.lock` to exist
- Downloads binaries (uses cache if available)
- Verifies SHA256 checksums against lock
- Creates symlinks in `~/.styx/bin/`
- Fails fast if any tool verification fails

---

## Future Commands (Phase 2+)

### styx update

Update specific tool version (planned).

```bash
styx update golang@1.24.0
# Updates .styx/styx.toml and styx.lock
```

### styx list

Show installed tools (planned).

```bash
styx list
# ripgrep   14.1.0  ~/.styx/bin/ripgrep
# golang    1.23.1  ~/.styx/bin/golang
# node      20.10.0 ~/.styx/bin/node
```

### styx info

Show tool metadata (planned).

```bash
styx info golang
# Name: golang
# Versions: 1.23.1, 1.22.0, 1.21.0
# Stability: stable
```

### styx search

Search registry (planned).

```bash
styx search "search term"
# Find tools matching term
```

---

## Global Flags

All commands support:

```bash
styx [command] --help    # Show help for command
styx --version           # Show Styx version
```

---

## Configuration Files

### Global Config (Optional)

`~/.local/share/styx/styx.toml` - System-wide defaults

```toml
[tools]
golang = "1.23.1"
node = "20.10.0"

[env]
RUST_BACKTRACE = "1"
```

### Local Config (Project-Specific)

`.styx/styx.toml` or `./styx.toml` - Per-project overrides

```toml
[tools]
golang = "1.22.0"  # Override global
postgres = "15.3"  # New tool

[env]
DATABASE_URL = "postgresql://localhost/dev"
```

**Merge behavior:**
- Local tools override global (by name)
- Local env vars override global (by key)
- Tools/vars in global but not local are inherited

---

## Storage & Directories

Styx creates the following directories on first run:

- `~/.styx/store/` - Content-addressable storage (tools by SHA256)
- `~/.styx/cache/` - Download cache
- `~/.styx/bin/` - Symlinks to installed tools

Example structure:
```
~/.styx/
├── store/
│   ├── abcd1234.../bin/rg              (ripgrep)
│   ├── efgh5678.../bin/go              (golang)
│   └── ijkl9012.../bin/node            (node)
├── cache/
│   ├── ripgrep-14.1.0.tar.gz
│   └── golang-1.23.1.tar.gz
└── bin/                                (symlinks)
    ├── rg -> ../store/abcd1234.../bin/rg
    ├── go -> ../store/efgh5678.../bin/go
    └── node -> ../store/ijkl9012.../bin/node
```

---

## Exit Codes

- `0` - Success
- `1` - General error (config, download, verification failed)
- `2` - Tool not found in registry
- `3` - Checksum mismatch

---

## Environment

Styx respects:

- `HOME` - User home directory (default: `~`)
- `PATH` - Used to find `styx` binary itself

Add `~/.styx/bin` to PATH to use installed tools:

```bash
export PATH="$HOME/.styx/bin:$PATH"
```

---

**For more examples, see [QUICKSTART.md](QUICKSTART.md) and the [examples/](examples/) directory.**
