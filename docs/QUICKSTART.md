# Styx Quick Start

Get a reproducible dev environment running in 5 minutes.

## Installation

```bash
git clone https://github.com/killallservers/styx
cd styx
go build -o styx ./cmd/styx
sudo mv styx /usr/local/bin/  # or add to PATH
```

## Setup (First Time)

**1. Create a project config:**

```bash
cd ~/my-project
mkdir -p .styx
```

**2. Add tools you need:**

```toml
# .styx/styx.toml
[tools]
ripgrep = "14.1.0"    # Fast grep
golang = "1.23.1"     # Go compiler
node = "20.10.0"      # Node runtime

[env]
DATABASE_URL = "postgresql://localhost/dev"
NODE_ENV = "development"
```

**3. Install:**

```bash
styx install
```

Output:
```
✓ ripgrep@14.1.0 installed to /home/you/.styx/bin/ripgrep
✓ golang@1.23.1 installed to /home/you/.styx/bin/golang
✓ node@20.10.0 installed to /home/you/.styx/bin/node
```

**4. Generate lock file (commit this!):**

```bash
styx lock
git add .styx/styx.toml styx.lock
git commit -m "Add reproducible dev environment"
```

## Usage

### Show loaded environment

```bash
styx env
```

Output:
```
DATABASE_URL=postgresql://localhost/dev
GOLANG_PATH=/home/you/.styx/bin/golang
NODE_ENV=development
NODE_PATH=/home/you/.styx/bin/node
RIPGREP_PATH=/home/you/.styx/bin/ripgrep
```

### Verify tools are correct

```bash
styx verify
```

Output:
```
✓ All 3 tools verified successfully
```

## For Your Team

When a teammate clones the repo:

```bash
git clone ~/my-project
cd my-project
styx install  # or: styx sync to use lock file
```

Everyone gets **exactly the same tool versions**, verified with SHA256 checksums.

## Available Tools

Styx ships with 8 pre-configured tools:

- **ripgrep** (14.1.0) - Fast grep
- **fd** (10.0.0) - Fast find
- **bat** (0.24.0) - Cat with syntax highlight
- **eza** (0.18.0) - Modern ls
- **golang** (1.23.1) - Go compiler
- **rust** (1.81.0) - Rust compiler
- **node** (20.10.0) - Node runtime
- **python** (3.12.0) - Python runtime

Request additional tools in GitHub issues.

## Config Format

### Tools

List any tool + version from the registry:

```toml
[tools]
golang = "1.23.1"
node = "20.10.0"
ripgrep = "14.1.0"
```

### Environment Variables

Set env vars that should be available in your shell:

```toml
[env]
DATABASE_URL = "postgresql://localhost/dev"
NODE_ENV = "development"
RUST_BACKTRACE = "1"
MY_VAR = "my_value"
```

Tool paths are **automatically added** (e.g., `GOLANG_PATH=/home/you/.styx/bin/golang`).

## Next Steps

- **Global config:** Set defaults in `~/.local/share/styx/styx.toml` (optional)
- **Update tools:** Edit `.styx/styx.toml` and run `styx lock`
- **More tools:** Request in GitHub or wait for Phase 2

## Troubleshooting

**"Tool X not found"**
- Make sure tool is in the registry: `styx env`
- Check spelling and version

**"Checksum mismatch"**
- Run `styx verify` to check all tools
- Delete `~/.styx/store` and reinstall if corrupted

**"Command not found after install"**
- Add `~/.styx/bin` to your PATH:
  ```bash
  export PATH="$HOME/.styx/bin:$PATH"
  ```

---

**That's it!** One config file, one lock file, identical environments everywhere.
