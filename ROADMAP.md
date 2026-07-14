# Styx Roadmap

## Current Release: v1.0.0 ✅

Styx v1.0.0 is production-ready with:
- 11 core commands (install, sync, lock, update, verify, env, list, info, search, init, completion)
- 11 built-in tools (ripgrep, fd, bat, eza, just, golang, rust, node, python, postgres, redis)
- Full lock file reproducibility
- Shell integration with auto-loading
- Docker support (dev + CI)
- 45+ passing tests
- Multi-registry support with caching
- Security audit and benchmarking

**Status:** Deployed to Modo Ventures portfolio companies ✅

---

## v1.1+ Future Enhancements

### High Priority

**Better Error Messages**
- Structured error context (what happened, why, how to fix)
- Helpful suggestions for common mistakes
- Error recovery guides

**Registry Expansion**
- 5-10 additional tools based on Modo feedback (docker, make, postgres CLI tools, etc.)
- Community contribution workflow
- Tool request tracking

**Performance Optimizations**
- Parallel registry loading
- Lazy spec evaluation
- Download progress bars (interactive mode)

### Medium Priority

**Advanced Registry Features**
- Custom/private registries for internal tools
- Registry metadata versioning
- Automatic registry freshness checks
- GPG signature verification for specs

**Shell Integration Enhancements**
- direnv integration improvements
- PowerShell support (currently bash/zsh/fish)
- Custom shell hooks
- Profile-specific tool loading ("dev", "ci", "prod" modes)

**Audit & Governance**
- Installation audit logs
- Tool provenance tracking
- Compliance reporting
- SBOM (Software Bill of Materials) generation

### Nice-to-Have

**UI/UX Improvements**
- Interactive tool selection wizard
- Config validation with suggestions
- TUI (Terminal UI) for browsing registry
- Colored output for better readability

**Distribution Enhancements**
- Arch Linux packaging
- Flatpak support
- Official Debian/Ubuntu packages
- Cross-platform binary signing

**Developer Experience**
- Styx plugin system (custom commands)
- Local tool specs (for internal/experimental tools)
- Config templates for common project types
- IDE integration (VS Code extension?)

### Research / Exploration

**Windows Support**
- Possible if demand exists
- Would require PowerShell integration and different storage strategies
- Deferred: No Windows request from Modo portfolio yet

**Build-from-Source Support**
- Could add for tools where binaries unavailable
- Complex: dependency resolution, build toolchain management
- Deferred: Use system package manager as fallback instead

**AI-Assisted Spec Generation**
- Use Claude to help curate new tool specs
- Verify outputs with CI before merge
- Not at install time (no runtime dependency)

---

## Known Limitations (v1.0.0)

- **No Windows:** Only Linux and macOS supported
- **Binary tools only:** No build-from-source (use system PM as fallback)
- **Registry growth:** Limited to ~20 tools without HTTP registry completion
- **No tool groups:** Can't declare "backend tools" vs "frontend tools"
- **No conditional deps:** All tools always installed (even if unused)

---

## Roadmap by Timeline

### Q3 2026 (Immediate)
- Expand registry based on Modo portfolio feedback
- Error message improvements
- Bug fixes from real usage

### Q4 2026
- Advanced registry features (custom registries, signing)
- Shell integration enhancements
- Audit logging

### Q1+ 2027
- Expand beyond Modo portfolio (if going public)
- Plugin system exploration
- Windows support evaluation

---

## Contributing

Want to suggest a feature or request a tool?

1. **New tool requests:** File a GitHub issue with use case and platform support needed
2. **Feature requests:** Describe the problem you're solving and how it would help
3. **Bug reports:** Include version, platform, config, and error message
4. **Code contributions:** Submit PRs with tests and documentation

See CONTRIBUTING.md (coming soon) for details.

---

## Success Metrics for Future Releases

- **Adoption:** Used by 5+ Modo portfolio companies
- **Registry:** 30+ tools available
- **Reliability:** <1% install failure rate
- **Speed:** Install 10 tools in <30 seconds
- **Feedback:** Positive feedback from portfolio teams
- **Community:** Contributors from outside Modo (if public release)

---

**Status:** Styx v1.0.0 is production-ready. The roadmap above guides future enhancements based on user needs and Modo portfolio feedback.

For current feature details, see [README.md](README.md) and [COMMANDS.md](docs/COMMANDS.md).
