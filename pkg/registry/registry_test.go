package registry

import (
	"testing"
)

func TestLoadEmbeddedRegistry(t *testing.T) {
	reg, err := LoadEmbeddedRegistry()
	if err != nil {
		t.Fatalf("LoadEmbeddedRegistry failed: %v", err)
	}

	expectedTools := []string{
		"ripgrep", "fd", "bat", "eza", "just",
		"golang", "rust", "node", "python",
		"postgres", "redis", "docker",
		"git", "curl", "jq", "tmux",
		"kubectl", "terraform", "aws-cli", "gcloud",
		"protoc", "grpcurl", "neovim", "buf",
		"docker-compose", "helm",
	}
	for _, toolName := range expectedTools {
		if _, ok := reg[toolName]; !ok {
			t.Errorf("expected tool %q not found in registry", toolName)
		}
	}

	if len(reg) != len(expectedTools) {
		t.Errorf("expected %d tools in registry, got %d", len(expectedTools), len(reg))
	}
}

func TestGetTool(t *testing.T) {
	reg, _ := LoadEmbeddedRegistry()

	tests := []struct {
		name     string
		toolName string
		wantErr  bool
	}{
		{"ripgrep exists", "ripgrep", false},
		{"node exists", "node", false},
		{"nonexistent tool", "nonexistent", true},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := GetTool(reg, tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTool(%q) error = %v, wantErr %v", tt.toolName, err, tt.wantErr)
			}
			if !tt.wantErr && spec == nil {
				t.Errorf("GetTool(%q) returned nil spec", tt.toolName)
			}
		})
	}
}

func TestToolSpecGetVersion(t *testing.T) {
	reg, _ := LoadEmbeddedRegistry()
	ripgrep := reg["ripgrep"]

	version := ripgrep.GetVersion("14.1.0")
	if version == nil {
		t.Error("expected to find version 14.1.0")
	}

	missingVersion := ripgrep.GetVersion("99.99.99")
	if missingVersion != nil {
		t.Error("expected nil for missing version")
	}
}

func TestVersionSpecGetMethodForPlatform(t *testing.T) {
	reg, _ := LoadEmbeddedRegistry()
	ripgrep := reg["ripgrep"]
	version := ripgrep.GetVersion("14.1.0")

	tests := []struct {
		name       string
		platform   string
		methodType string
		wantFound  bool
	}{
		{"linux binary", "linux-x86_64", "binary", true},
		{"darwin binary", "darwin-arm64", "binary", true},
		{"missing platform", "windows-x86_64", "binary", false},
		{"missing method type", "linux-x86_64", "source", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := version.GetMethodForPlatform(tt.platform, tt.methodType)
			if (method != nil) != tt.wantFound {
				t.Errorf("GetMethodForPlatform(%q, %q) found=%v, wantFound=%v",
					tt.platform, tt.methodType, method != nil, tt.wantFound)
			}
		})
	}
}

func TestEmbeddedRegistryIntegrity(t *testing.T) {
	reg, _ := LoadEmbeddedRegistry()

	// Each tool should have versions
	for toolName, spec := range reg {
		if len(spec.Versions) == 0 {
			t.Errorf("tool %q has no versions", toolName)
		}

		// Each version should have methods
		for version, versionSpec := range spec.Versions {
			if len(versionSpec.Methods) == 0 {
				t.Errorf("tool %q version %q has no methods", toolName, version)
			}

			// Each method should have platforms and checksums
			for _, method := range versionSpec.Methods {
				if len(method.Platforms) == 0 {
					t.Errorf("tool %q version %q method %q has no platforms",
						toolName, version, method.Type)
				}
				if len(method.Checksums) == 0 {
					t.Errorf("tool %q version %q method %q has no checksums",
						toolName, version, method.Type)
				}
				// Platforms and checksums should match
				if len(method.Platforms) != len(method.Checksums) {
					t.Errorf("tool %q version %q method %q has mismatched platform/checksum counts",
						toolName, version, method.Type)
				}
			}
		}
	}
}

func BenchmarkLoadEmbeddedRegistry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadEmbeddedRegistry()
	}
}

func BenchmarkGetTool(b *testing.B) {
	reg, _ := LoadEmbeddedRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTool(reg, "ripgrep")
	}
}

func BenchmarkGetVersion(b *testing.B) {
	reg, _ := LoadEmbeddedRegistry()
	ripgrep := reg["ripgrep"]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ripgrep.GetVersion("14.1.0")
	}
}

func BenchmarkGetMethodForPlatform(b *testing.B) {
	reg, _ := LoadEmbeddedRegistry()
	ripgrep := reg["ripgrep"]
	version := ripgrep.GetVersion("14.1.0")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version.GetMethodForPlatform("linux-x86_64", "binary")
	}
}
