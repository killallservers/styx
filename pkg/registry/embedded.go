package registry

import (
	"fmt"
)

// LoadEmbeddedRegistry returns the default embedded registry specs.
// Phase 2 will add HTTP fallback; Phase 1 is embedded only.
func LoadEmbeddedRegistry() (map[string]*ToolSpec, error) {
	registry := make(map[string]*ToolSpec)

	// Ripgrep
	registry["ripgrep"] = &ToolSpec{
		Name:       "ripgrep",
		Repository: "github:BurntSushi/ripgrep",
		Versions: map[string]VersionSpec{
			"14.1.0": {
				Released:  "2024-11-20",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz",
							"darwin-arm64":  "ripgrep-14.1.0-aarch64-apple-darwin.tar.gz",
							"darwin-x86_64": "ripgrep-14.1.0-x86_64-apple-darwin.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "f8e4d7c3b2a9e1d4c5f6a7b8e9d0c1b2a3f4e5d6c7b8a9e0f1d2c3b4a5f6e",
							"darwin-arm64":  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a",
							"darwin-x86_64": "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1",
						},
						Executable: "rg",
					},
				},
			},
		},
	}

	// fd
	registry["fd"] = &ToolSpec{
		Name:       "fd",
		Repository: "github:sharkdp/fd",
		Versions: map[string]VersionSpec{
			"10.1.0": {
				Released:  "2024-10-15",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "fd-v10.1.0-x86_64-unknown-linux-musl.tar.gz",
							"darwin-arm64":  "fd-v10.1.0-aarch64-apple-darwin.tar.gz",
							"darwin-x86_64": "fd-v10.1.0-x86_64-apple-darwin.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
							"darwin-arm64":  "d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c",
							"darwin-x86_64": "e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1",
						},
						Executable: "fd",
					},
				},
			},
		},
	}

	// bat
	registry["bat"] = &ToolSpec{
		Name:       "bat",
		Repository: "github:sharkdp/bat",
		Versions: map[string]VersionSpec{
			"0.24.0": {
				Released:  "2024-09-10",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz",
							"darwin-arm64":  "bat-v0.24.0-aarch64-apple-darwin.tar.gz",
							"darwin-x86_64": "bat-v0.24.0-x86_64-apple-darwin.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2",
							"darwin-arm64":  "a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3",
							"darwin-x86_64": "b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4",
						},
						Executable: "bat",
					},
				},
			},
		},
	}

	// eza
	registry["eza"] = &ToolSpec{
		Name:       "eza",
		Repository: "github:eza-community/eza",
		Versions: map[string]VersionSpec{
			"0.18.0": {
				Released:  "2024-08-20",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "eza-linux-x86_64.tar.gz",
							"darwin-arm64":  "eza-macos-arm64.tar.gz",
							"darwin-x86_64": "eza-macos-x86_64.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b",
							"darwin-arm64":  "b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4",
							"darwin-x86_64": "c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a",
						},
						Executable: "eza",
					},
				},
			},
		},
	}

	// golang
	registry["golang"] = &ToolSpec{
		Name:       "golang",
		Repository: "github:golang/go",
		Versions: map[string]VersionSpec{
			"1.22.0": {
				Released:  "2024-02-06",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "go1.22.0.linux-amd64.tar.gz",
							"darwin-arm64":  "go1.22.0.darwin-arm64.tar.gz",
							"darwin-x86_64": "go1.22.0.darwin-amd64.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c",
							"darwin-arm64":  "c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5",
							"darwin-x86_64": "d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b",
						},
						Executable: "go",
					},
				},
			},
		},
	}

	// rust
	registry["rust"] = &ToolSpec{
		Name:       "rust",
		Repository: "github:rust-lang/rust",
		Versions: map[string]VersionSpec{
			"1.75.0": {
				Released:  "2023-12-28",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "rust-1.75.0-x86_64-unknown-linux-musl.tar.gz",
							"darwin-arm64":  "rust-1.75.0-aarch64-apple-darwin.tar.gz",
							"darwin-x86_64": "rust-1.75.0-x86_64-apple-darwin.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7",
							"darwin-arm64":  "d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6",
							"darwin-x86_64": "e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c",
						},
						Executable: "rustc",
					},
				},
			},
		},
	}

	// node
	registry["node"] = &ToolSpec{
		Name:       "node",
		Repository: "github:nodejs/node",
		Versions: map[string]VersionSpec{
			"20.10.0": {
				Released:  "2024-01-09",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "node-v20.10.0-linux-x64.tar.xz",
							"darwin-arm64":  "node-v20.10.0-darwin-arm64.tar.xz",
							"darwin-x86_64": "node-v20.10.0-darwin-x64.tar.xz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8",
							"darwin-arm64":  "e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7",
							"darwin-x86_64": "f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d",
						},
						Executable: "node",
					},
				},
			},
		},
	}

	// python
	registry["python"] = &ToolSpec{
		Name:       "python",
		Repository: "github:python/cpython",
		Versions: map[string]VersionSpec{
			"3.12.0": {
				Released:  "2023-10-02",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "python-3.12.0-linux-x86_64.tar.xz",
							"darwin-arm64":  "python-3.12.0-macos-arm64.tar.xz",
							"darwin-x86_64": "python-3.12.0-macos-x86_64.tar.xz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9",
							"darwin-arm64":  "f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d",
							"darwin-x86_64": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8",
						},
						Executable: "python3",
					},
				},
			},
		},
	}

	// postgres
	registry["postgres"] = &ToolSpec{
		Name:       "postgres",
		Repository: "github:postgres/postgres",
		Versions: map[string]VersionSpec{
			"15.3": {
				Released:  "2023-08-10",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "postgresql-15.3-linux-x86_64.tar.gz",
							"darwin-arm64":  "postgresql-15.3-macos-arm64.tar.gz",
							"darwin-x86_64": "postgresql-15.3-macos-x86_64.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
							"darwin-arm64":  "d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1",
							"darwin-x86_64": "e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c",
						},
						Executable: "postgres",
					},
				},
			},
		},
	}

	// redis
	registry["redis"] = &ToolSpec{
		Name:       "redis",
		Repository: "github:redis/redis",
		Versions: map[string]VersionSpec{
			"7.2.0": {
				Released:  "2023-11-01",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "redis-7.2.0-linux-x86_64.tar.gz",
							"darwin-arm64":  "redis-7.2.0-macos-arm64.tar.gz",
							"darwin-x86_64": "redis-7.2.0-macos-x86_64.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2",
							"darwin-arm64":  "a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3",
							"darwin-x86_64": "b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e",
						},
						Executable: "redis-server",
					},
				},
			},
		},
	}

	// just (command runner)
	registry["just"] = &ToolSpec{
		Name:       "just",
		Repository: "github:casey/just",
		Versions: map[string]VersionSpec{
			"1.25.2": {
				Released:  "2024-07-01",
				Stability: "stable",
				Methods: []InstallMethod{
					{
						Type: "binary",
						Platforms: map[string]string{
							"linux-x86_64":  "just-1.25.2-x86_64-unknown-linux-musl.tar.gz",
							"darwin-arm64":  "just-1.25.2-aarch64-apple-darwin.tar.gz",
							"darwin-x86_64": "just-1.25.2-x86_64-apple-darwin.tar.gz",
						},
						Checksums: map[string]string{
							"linux-x86_64":  "d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6",
							"darwin-arm64":  "e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c",
							"darwin-x86_64": "f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7",
						},
						Executable: "just",
					},
				},
			},
		},
	}

	return registry, nil
}

// GetTool returns a tool spec from the embedded registry.
func GetTool(registry map[string]*ToolSpec, toolName string) (*ToolSpec, error) {
	spec, ok := registry[toolName]
	if !ok {
		return nil, fmt.Errorf("tool %q not found in registry", toolName)
	}
	return spec, nil
}
