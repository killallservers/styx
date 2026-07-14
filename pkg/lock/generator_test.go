package lock

import (
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	cfg := &GeneratorConfig{
		RegistryURL:     "https://example.com/registry",
		RegistryVersion: "1.0.0",
		Tools: map[string]InstallationRecord{
			"ripgrep": {
				Name:         "ripgrep",
				Version:      "14.1.0",
				Method:       "binary",
				StorePath:    "/home/user/.styx/store/ripgrep/14.1.0",
				BinaryHash:   "abc123def456",
				Executable:   "rg",
				SourceConfig: "config",
			},
		},
		Env: map[string]string{
			"RUST_BACKTRACE": "1",
		},
	}

	lf := Generate(cfg)

	if lf.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", lf.Version)
	}

	if lf.RegistrySnapshot.URL != "https://example.com/registry" {
		t.Errorf("expected registry URL, got %s", lf.RegistrySnapshot.URL)
	}

	if lf.RegistrySnapshot.Version != "1.0.0" {
		t.Errorf("expected registry version 1.0.0, got %s", lf.RegistrySnapshot.Version)
	}

	if len(lf.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(lf.Tools))
	}

	if lf.Tools[0].Name != "ripgrep" {
		t.Errorf("expected tool name ripgrep, got %s", lf.Tools[0].Name)
	}

	if len(lf.Env) != 1 {
		t.Errorf("expected 1 env var, got %d", len(lf.Env))
	}

	if lf.Env["RUST_BACKTRACE"] != "1" {
		t.Errorf("expected RUST_BACKTRACE=1, got %s", lf.Env["RUST_BACKTRACE"])
	}
}

func TestGenerateMultipleTools(t *testing.T) {
	cfg := &GeneratorConfig{
		RegistryURL:     "embedded",
		RegistryVersion: "0.1.0",
		Tools: map[string]InstallationRecord{
			"ripgrep": {
				Name:         "ripgrep",
				Version:      "14.1.0",
				Method:       "binary",
				StorePath:    "/path/to/ripgrep",
				BinaryHash:   "hash1",
				Executable:   "rg",
				SourceConfig: "global",
			},
			"fd": {
				Name:         "fd",
				Version:      "10.1.0",
				Method:       "binary",
				StorePath:    "/path/to/fd",
				BinaryHash:   "hash2",
				Executable:   "fd",
				SourceConfig: "local",
			},
			"golang": {
				Name:         "golang",
				Version:      "1.22.0",
				Method:       "binary",
				StorePath:    "/path/to/golang",
				BinaryHash:   "hash3",
				Executable:   "go",
				SourceConfig: "global",
			},
		},
		Env: map[string]string{},
	}

	lf := Generate(cfg)

	if len(lf.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(lf.Tools))
	}

	// Verify all tools are present
	toolMap := make(map[string]*ToolEntry)
	for i := range lf.Tools {
		toolMap[lf.Tools[i].Name] = &lf.Tools[i]
	}

	for _, name := range []string{"ripgrep", "fd", "golang"} {
		if _, ok := toolMap[name]; !ok {
			t.Errorf("expected tool %s not found", name)
		}
	}
}

func TestGenerateTimestamp(t *testing.T) {
	cfg := &GeneratorConfig{
		RegistryURL:     "test",
		RegistryVersion: "1.0",
		Tools:           make(map[string]InstallationRecord),
		Env:             make(map[string]string),
	}

	before := time.Now()
	lf := Generate(cfg)
	after := time.Now()

	if lf.GeneratedAt.Before(before) || lf.GeneratedAt.After(after) {
		t.Errorf("GeneratedAt timestamp out of range")
	}
}

func TestVerifyAgainstLock(t *testing.T) {
	lf := &LockFile{
		Version: "1.0",
		Tools: []ToolEntry{
			{
				Name:       "ripgrep",
				Version:    "14.1.0",
				BinaryHash: "abc123",
			},
			{
				Name:       "fd",
				Version:    "10.1.0",
				BinaryHash: "def456",
			},
		},
		Env: make(map[string]string),
	}

	// Test matching hashes
	actual := map[string]string{
		"ripgrep": "abc123",
		"fd":      "def456",
	}

	mismatches, err := VerifyAgainstLock(lf, actual)
	if err != nil {
		t.Fatalf("VerifyAgainstLock failed: %v", err)
	}

	if len(mismatches) != 0 {
		t.Errorf("expected no mismatches, got %d: %v", len(mismatches), mismatches)
	}
}

func TestVerifyAgainstLockMissing(t *testing.T) {
	lf := &LockFile{
		Version: "1.0",
		Tools: []ToolEntry{
			{
				Name:       "ripgrep",
				Version:    "14.1.0",
				BinaryHash: "abc123",
			},
		},
		Env: make(map[string]string),
	}

	// Tool is missing in actual
	actual := map[string]string{}

	mismatches, err := VerifyAgainstLock(lf, actual)
	if err != nil {
		t.Fatalf("VerifyAgainstLock failed: %v", err)
	}

	if len(mismatches) != 1 {
		t.Errorf("expected 1 mismatch, got %d", len(mismatches))
	}
}

func TestVerifyAgainstLockHashMismatch(t *testing.T) {
	lf := &LockFile{
		Version: "1.0",
		Tools: []ToolEntry{
			{
				Name:       "ripgrep",
				Version:    "14.1.0",
				BinaryHash: "abc123",
			},
		},
		Env: make(map[string]string),
	}

	// Hash doesn't match
	actual := map[string]string{
		"ripgrep": "wronghash",
	}

	mismatches, err := VerifyAgainstLock(lf, actual)
	if err != nil {
		t.Fatalf("VerifyAgainstLock failed: %v", err)
	}

	if len(mismatches) != 1 {
		t.Errorf("expected 1 mismatch, got %d", len(mismatches))
	}
}

func TestVerifyAgainstLockMultipleMismatches(t *testing.T) {
	lf := &LockFile{
		Version: "1.0",
		Tools: []ToolEntry{
			{
				Name:       "ripgrep",
				Version:    "14.1.0",
				BinaryHash: "abc123",
			},
			{
				Name:       "fd",
				Version:    "10.1.0",
				BinaryHash: "def456",
			},
			{
				Name:       "bat",
				Version:    "0.24.0",
				BinaryHash: "ghi789",
			},
		},
		Env: make(map[string]string),
	}

	// One missing, one mismatched
	actual := map[string]string{
		"ripgrep": "abc123",
		"fd":      "wronghash",
	}

	mismatches, err := VerifyAgainstLock(lf, actual)
	if err != nil {
		t.Fatalf("VerifyAgainstLock failed: %v", err)
	}

	if len(mismatches) != 2 {
		t.Errorf("expected 2 mismatches, got %d", len(mismatches))
	}
}
