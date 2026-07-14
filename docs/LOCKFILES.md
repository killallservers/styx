# Styx Lock Files

Understanding and using `styx.lock` for reproducible environments.

## What is a Lock File?

A lock file is a JSON snapshot of exact tool versions and checksums at a specific moment in time.

When you run `styx lock`, it creates `styx.lock` that records:
- Exact version of each tool
- Download URL for each tool
- SHA256 checksum of each binary
- Environment variables
- Registry version used

**Key concept:** Anyone with `styx.lock` can reproduce the identical environment.

## Creating a Lock File

### Generate Initial Lock

```bash
styx lock
```

Output:
```
✓ Lock file generated: styx.lock (3 tools)
```

This creates `styx.lock` in the current directory.

### Update Existing Lock

Edit your config:

```toml
[tools]
golang = "1.24.0"  # Updated from 1.23.1
```

Then regenerate:

```bash
styx lock
```

The lock file updates with new version info.

## Lock File Format

```json
{
  "version": "1.0",
  "generated_at": "2026-07-13T20:30:00Z",
  "registry_snapshot": {
    "url": "embedded",
    "version": "0.1.0"
  },
  "tools": [
    {
      "name": "golang",
      "version": "1.23.1",
      "install_method": "binary",
      "storage_path": "~/.styx/store/abc123.../bin/go",
      "binary_hash_sha256": "abc123def456ghi789jkl012mno345pqr678stu9vw",
      "executable": "go",
      "source_config": "local"
    },
    {
      "name": "postgres",
      "version": "15.3",
      "install_method": "binary",
      "storage_path": "~/.styx/store/def456.../bin/postgres",
      "binary_hash_sha256": "def456ghi789jkl012mno345pqr678stu9vw0123xyz",
      "executable": "postgres",
      "source_config": "local"
    }
  ],
  "env": {
    "DATABASE_URL": "postgresql://localhost/dev",
    "DEBUG": "true"
  }
}
```

### Fields Explained

- **version** - Lock file schema version (for forward compatibility)
- **generated_at** - When lock was created (ISO 8601 timestamp)
- **registry_snapshot** - Which registry was used (embedded in Phase 1)
- **tools[].name** - Tool name
- **tools[].version** - Exact version (e.g., "1.23.1")
- **tools[].install_method** - How tool was installed ("binary" in Phase 1)
- **tools[].storage_path** - Where binary is stored (~/.styx/store/HASH/...)
- **tools[].binary_hash_sha256** - SHA256 checksum of downloaded binary
- **tools[].executable** - Binary name after extraction (e.g., "go", "postgres")
- **tools[].source_config** - Whether tool came from "global" or "local" config
- **env** - Merged environment variables

## Using Lock Files

### Reproducing an Environment

Install from lock file:

```bash
styx install
# OR in Phase 2: styx sync
```

Styx verifies each downloaded binary against checksums in lock file.

### Verifying Integrity

Check that installed tools match lock file:

```bash
styx verify
```

Output:
```
✓ All 2 tools verified successfully
```

If mismatch:
```
✗ Verification failed:
  - golang: checksum mismatch
  - postgres: file not found
```

Fix with `styx install` to reinstall.

### Sharing Across Team

Commit lock file to git:

```bash
git add styx.lock .styx/styx.toml
git commit -m "Update dev environment"
git push
```

Teammates pull and install:

```bash
git pull
styx install
```

Everyone gets identical binaries, verified by checksum.

## CI/CD Integration

### GitHub Actions

```yaml
name: Test
on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup dev environment
        run: |
          # Build Styx
          go build -o styx ./cmd/styx
          
          # Install from lock
          ./styx install
          
      - name: Run tests
        run: |
          # Tools from lock are in ~/.styx/bin/
          export PATH="$HOME/.styx/bin:$PATH"
          make test
```

### Docker

```dockerfile
FROM ubuntu:22.04

# Build Styx
COPY . /tmp/styx
RUN cd /tmp/styx && go build -o styx ./cmd/styx

# Copy lock file
COPY styx.lock .styx/ /app/

# Install from lock
RUN cd /app && /tmp/styx/styx install

# Use tools
ENV PATH="${HOME}/.styx/bin:${PATH}"
CMD ["go", "test", "./..."]
```

## Workflow: Config → Lock → Share

1. **Create/edit config:**
   ```bash
   cat > .styx/styx.toml << 'EOF'
   [tools]
   golang = "1.23.1"
   postgres = "15.3"
   
   [env]
   DATABASE_URL = "postgresql://localhost/dev"
   EOF
   ```

2. **Generate lock:**
   ```bash
   styx lock
   ```

3. **Review and commit:**
   ```bash
   git diff styx.lock  # Review changes
   git add .styx/styx.toml styx.lock
   git commit -m "Update dev environment"
   ```

4. **Team pulls and installs:**
   ```bash
   git pull
   styx install  # Verifies against lock
   ```

## Version Updates

To update a tool version:

### Manual Process

1. Edit `.styx/styx.toml`:
   ```bash
   sed -i 's/golang = "1.23.1"/golang = "1.24.0"/' .styx/styx.toml
   ```

2. Regenerate lock:
   ```bash
   styx lock
   ```

3. Review and commit:
   ```bash
   git diff styx.lock  # See what changed
   git add styx.lock .styx/styx.toml
   git commit -m "Upgrade golang to 1.24.0"
   ```

4. Team syncs:
   ```bash
   git pull
   styx install  # Installs new version
   ```

### Automated (Future)

Phase 2+ planned:
```bash
styx update golang@1.24.0
# Updates .styx/styx.toml and styx.lock automatically
```

## Lock File Best Practices

1. **Commit to git** - Enables team reproducibility
   ```bash
   git add styx.lock
   git commit -m "Lock environment"
   ```

2. **Review lock changes** - When updating tools
   ```bash
   git diff styx.lock  # See version/checksum changes
   ```

3. **Don't modify manually** - Always use `styx lock` to regenerate
   - Hand-edited lock files may have wrong checksums
   - Regenerate when config changes

4. **Keep lock updated** - Don't let it drift from config
   ```bash
   # After editing .styx/styx.toml:
   styx lock  # Always regenerate
   ```

5. **One lock per project** - Not per-directory (Phase 1)
   - In Phase 3, directory-specific overrides may be added
   - For now, single lock file per project

## Troubleshooting

### "Lock file not found"

If you haven't created one yet:

```bash
styx lock
```

### "Checksum mismatch during install"

This means a downloaded binary doesn't match the lock file checksum.

**Possible causes:**
- Binary corrupted during download
- Styx binary has a bug
- Someone tampered with download

**Fix:**
```bash
# Reinstall from scratch
rm -rf ~/.styx/cache
styx install
```

Then verify:
```bash
styx verify
```

### "Lock file is out of date"

If you edited config but didn't regenerate lock:

```bash
styx lock  # Regenerate
git diff styx.lock  # Review
git add styx.lock .styx/styx.toml
git commit -m "Update lock"
```

### "Different tools on different machines"

This usually means lock files are different or missing.

Check:
```bash
git diff styx.lock  # Should be clean
git status styx.lock  # Should be committed
```

If uncommitted changes:
```bash
git add styx.lock
git commit -m "Update lock"
git push
```

Then on other machines:
```bash
git pull
styx install  # Uses lock
```

## Security

### Checksum Verification

The lock file includes SHA256 checksums of all binaries.

When you run `styx install` or `styx verify`, Styx:
1. Downloads each binary
2. Computes SHA256 of downloaded file
3. Compares against checksum in lock file
4. Fails if mismatch (corrupted or tampered)

**Why this matters:**
- Detects corrupted downloads
- Prevents use of compromised binaries
- Ensures team uses same binary versions

### Lock File Security

Lock file is JSON (human-readable, git-friendly).

**Treat like source code:**
- Commit to version control
- Review changes when updating versions
- Don't commit with secrets (use config files outside git)

### Offline Verification

Once installed, verify without network:

```bash
styx verify
# Checks local ~/.styx/store/ against lock file checksums
```

## Future: Phase 2+

### Planned Enhancements

**Phase 2:**
- `styx sync` - Install from lock without config (for CI)
- `styx update` - Selective version updates
- HTTP registry fallback

**Phase 3:**
- Directory-specific lock files (hierarchical overrides)
- Automatic lock regeneration in CI
- Integration with shell profile (~/.bashrc)

**Phase 4:**
- Signed lock files (GPG signatures)
- Registry verification CI badges
- Lock file history/changelog

---

**Lock files make reproducibility concrete: one file, exact versions, verified checksums, shareable with your team.**

**Commit `styx.lock` to git. Your whole team gets identical environments.**
