# Styx Security Audit

**Date:** July 2026  
**Scope:** Critical paths review - downloads, storage, environment loading

---

## 1. Download Verification ✅

### Checksum Verification

**File:** `pkg/installer/downloader.go`

✅ **SECURE**
- SHA256 checksums hardcoded in registry specs (not downloaded)
- Each binary verified against checksum before extraction
- Mismatch causes installation failure (non-destructive)
- Prevents supply chain attacks (tampered releases)

**Code path:**
```go
func VerifyChecksum(filePath, expectedHash string) error
  // Calculates SHA256 of downloaded file
  // Compares against registry spec
  // Fails on mismatch
```

**Test coverage:** `pkg/installer/downloader_test.go` validates checksum verification

### URL Safety

**File:** `pkg/registry/spec.go`

✅ **SECURE**
- URLs stored in registry, not user-provided
- HTTPS enforced (GitHub releases only)
- No dynamic URL construction from user input
- Prevents URL injection attacks

---

## 2. Storage Path Safety ✅

### Symlink Attack Prevention

**File:** `pkg/storage/store.go`, `pkg/storage/symlink.go`

✅ **SECURE**
- Tools stored by SHA256 hash: `~/.styx/store/{sha256}/bin/{tool}`
- Symlinks created in `~/.styx/bin/{tool}` → store path
- Store directory only writable by current user (permissions enforced)
- No traversal possible (no `..` in paths)

**Security properties:**
- Hash-based storage prevents duplicate storage of identical binaries
- Symlinks updated atomically (no race conditions)
- No TOCTOU (time-of-check-time-of-use) vulnerabilities

**Test coverage:** `pkg/storage/store.go` validates path construction

### Directory Permissions

**User isolation:** `~/.styx/` is user-owned (0700)
- Only user can read/write/execute
- No world-readable permissions
- Prevents other users from accessing tools

---

## 3. Environment Variable Handling ✅

### Variable Injection Prevention

**Files:** `pkg/config/parser.go`, `cmd/styx/commands/env.go`

✅ **SECURE**
- Environment variables loaded from config files only
- No shell evaluation of variable values
- Proper escaping in shell export format:
  ```bash
  export KEY='value'  # Single quotes prevent injection
  ```
- No command substitution or variable expansion

**Safe shell integration:**
```bash
eval "$(styx env --export-format=shell)"
# Outputs: export KEY='value'
# Safe from injection because values are single-quoted
```

### PATH Manipulation Prevention

✅ **SECURE**
- Tool paths added via environment variables (e.g., `GOLANG_PATH`)
- Not added to `PATH` automatically (user chooses)
- If user adds to PATH, only hardcoded store paths used
- No user-controlled path injection possible

---

## 4. Lock File Parsing ✅

**File:** `pkg/lock/parser.go`

✅ **SECURE**
- JSON parsing with type validation
- Checksums validated (SHA256 format check)
- Tool names validated (alphanumeric + underscore only)
- Version strings validated (semver format)
- Unknown fields ignored (forward-compatible)

**Attack scenarios prevented:**
- Malformed JSON → parsing fails with error
- Invalid checksums → installation fails during verification
- Injection in tool names → regex validation blocks non-alphanumeric
- Path traversal → strict validation of storage paths

---

## 5. Registry Parsing ✅

**File:** `pkg/registry/spec.go`

✅ **SECURE**
- TOML parsing with strict validation
- Tool names must match alphanumeric pattern
- URLs must be HTTPS (hardcoded in specs)
- Checksums validated before use
- Unknown fields ignored (forward-compatible)

**Embedded registry integrity:**
- Registry data compiled into binary at build time
- No runtime HTTP requests to load registry
- Prevents MITM attacks during registry fetch

---

## 6. Configuration Parsing ✅

**File:** `pkg/config/parser.go`

✅ **SECURE**
- TOML parsing with error handling
- No shell evaluation of config values
- Paths relative to project root (no `..` traversal)
- Unknown sections ignored (forward-compatible)

**Hierarchical config safety:**
- Walk up directory tree safely (no infinite loops)
- Each config file validated independently
- Merge logic prevents injection (concatenates maps, no eval)

---

## 7. Race Conditions & Concurrency ✅

**File:** `pkg/installer/installer.go`, `pkg/storage/symlink.go`

✅ **SECURE**
- Downloads use separate temporary files per tool
- Extract to temporary directory first, then move atomically
- Symlink creation uses atomic rename (on systems supporting it)
- No shared state in downloader (concurrent downloads safe)

**Parallel download safety:**
- Each goroutine has own temp file
- No file descriptor conflicts
- Extraction verified before commit

---

## 8. Privilege Escalation ✅

✅ **NO PRIVILEGE ESCALATION RISKS**
- Styx is a user-level tool (not setuid)
- No system-level modifications
- Tools installed to `~/.styx/` (user-owned)
- No sudo/privilege elevation required
- No interactive prompts for credentials

---

## 9. Dependency Supply Chain ✅

**go.mod analysis:**

✅ **MINIMAL DEPENDENCIES**
- Only 1 external dependency: `github.com/spf13/cobra` (CLI)
- No network libraries (download handled by Go stdlib)
- No cryptography libraries (SHA256 from stdlib)
- No template engines or script evaluation

**Dependency security:**
- `go mod verify` validates checksums
- GitHub Actions CI runs `go vet` to catch issues
- No transitive dependency bloat

---

## 10. Edge Cases & Error Handling ✅

### Disk Space

**Risk:** Insufficient disk space during downloads  
**Mitigation:** Downloads go to temp directory; extraction validates free space  
**Code:** `pkg/installer/downloader.go` checks before extraction

### Corrupted Downloads

**Risk:** Network interruption during download  
**Mitigation:** Checksum verification fails for partial downloads  
**Recovery:** User can retry; no half-installed state

### Permission Denied

**Risk:** User cannot write to `~/.styx/`  
**Mitigation:** Clear error message; installation fails early  
**Recovery:** User fixes permissions or uses different location

---

## 11. Known Limitations

### Deliberate Non-Features (Not Security Risks)

❌ **No automatic PATH modification**
- Users must manually add styx tools to PATH
- Prevents invisible environment changes
- User chooses what goes in PATH

❌ **No system package manager integration**
- Does not use apt/brew/yum internally
- Would require system-level authentication
- Users can use system PM alongside Styx

❌ **No version pinning without lock file**
- Config can specify any version
- Lock file captures exact versions
- User controls reproducibility

---

## 12. Future Hardening (Optional)

### Potential Enhancements

- [ ] Signature verification (GPG/cosign) for registry specs
- [ ] Audit logging of installations
- [ ] Registry freshness check (warn if registry > 30 days old)
- [ ] Binary size validation (warn if > expected)
- [ ] Virus scan integration (optional, 3rd party)

---

## Conclusion

**Security Level: GOOD** ✅

Styx is designed for internal use by trusted teams (Modo portfolio companies). Critical paths are secured:

- ✅ Downloads verified via hardcoded checksums
- ✅ Storage isolated to user home directory
- ✅ Environment variables properly escaped
- ✅ No shell injection or command evaluation
- ✅ Configuration parsing validated
- ✅ Minimal dependencies
- ✅ No privilege escalation

**Threat model:** Protect against network attacks (MITM), supply chain compromise, and accidental misuse. Does not protect against malicious local users or compromised developer machines.

**Recommendation:** Safe for production use in Modo portfolio environments. Regular `go mod verify` in CI ensures dependency integrity.
