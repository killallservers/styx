package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfigsUpward(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create nested directories with .styx/styx.toml files
	level1 := filepath.Join(tmpDir, "level1")
	level2 := filepath.Join(level1, "level2")
	level3 := filepath.Join(level2, "level3")

	for _, dir := range []string{level1, level2, level3} {
		if err := os.MkdirAll(filepath.Join(dir, ".styx"), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// Create config files at each level
	for _, dir := range []string{level1, level2, level3} {
		configPath := filepath.Join(dir, ".styx", "styx.toml")
		if err := os.WriteFile(configPath, []byte("[tools]\n"), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
	}

	// Change to level3 directory
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(level3)

	paths, err := findConfigsUpward()
	if err != nil {
		t.Fatalf("findConfigsUpward failed: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("Expected 3 config paths, got %d", len(paths))
	}

	// Verify order: should be from root-most to current
	expected := []string{
		filepath.Join(level1, ".styx", "styx.toml"),
		filepath.Join(level2, ".styx", "styx.toml"),
		filepath.Join(level3, ".styx", "styx.toml"),
	}

	for i, exp := range expected {
		if i >= len(paths) {
			t.Errorf("Missing path %d: %s", i, exp)
			continue
		}
		if paths[i] != exp {
			t.Errorf("Path %d: expected %s, got %s", i, exp, paths[i])
		}
	}
}

func TestFindConfigsUpwardPartial(t *testing.T) {
	// Create a temporary directory structure with configs only at some levels
	tmpDir := t.TempDir()

	level1 := filepath.Join(tmpDir, "level1")
	level2 := filepath.Join(level1, "level2")
	level3 := filepath.Join(level2, "level3")

	// Only create level1 and level3
	for _, dir := range []string{level1, level3} {
		if err := os.MkdirAll(filepath.Join(dir, ".styx"), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		configPath := filepath.Join(dir, ".styx", "styx.toml")
		if err := os.WriteFile(configPath, []byte("[tools]\n"), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
	}

	// Create level2 and level3 dirs but no config at level2
	for _, dir := range []string{level2, level3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(level3)

	paths, err := findConfigsUpward()
	if err != nil {
		t.Fatalf("findConfigsUpward failed: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("Expected 2 config paths (skipping level2), got %d", len(paths))
	}
}

func TestLoadHierarchicalMerge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two levels
	level1 := filepath.Join(tmpDir, "level1")
	level2 := filepath.Join(level1, "level2")

	for _, dir := range []string{level1, level2} {
		if err := os.MkdirAll(filepath.Join(dir, ".styx"), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// Level 1 config: golang@1.22, RUST_BACKTRACE=1
	level1Config := `[tools]
golang = "1.22.0"
node = "20.0.0"

[env]
RUST_BACKTRACE = "1"
`
	if err := os.WriteFile(filepath.Join(level1, ".styx", "styx.toml"), []byte(level1Config), 0644); err != nil {
		t.Fatalf("Failed to write level1 config: %v", err)
	}

	// Level 2 config: override golang to 1.21, add postgres, add DATABASE_URL
	level2Config := `[tools]
golang = "1.21.0"
postgres = "15.3"

[env]
DATABASE_URL = "postgres://localhost"
`
	if err := os.WriteFile(filepath.Join(level2, ".styx", "styx.toml"), []byte(level2Config), 0644); err != nil {
		t.Fatalf("Failed to write level2 config: %v", err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(level2)

	merged, err := LoadHierarchical()
	if err != nil {
		t.Fatalf("LoadHierarchical failed: %v", err)
	}

	// Verify tool merging: level2 overrides level1's golang, inherits node, adds postgres
	if merged.Tools["golang"] != "1.21.0" {
		t.Errorf("Expected golang=1.21.0 (level2 override), got %s", merged.Tools["golang"])
	}
	if merged.Tools["node"] != "20.0.0" {
		t.Errorf("Expected node=20.0.0 (inherited from level1), got %s", merged.Tools["node"])
	}
	if merged.Tools["postgres"] != "15.3" {
		t.Errorf("Expected postgres=15.3 (new in level2), got %s", merged.Tools["postgres"])
	}

	// Verify env merging: both RUST_BACKTRACE and DATABASE_URL present
	if merged.Env["RUST_BACKTRACE"] != "1" {
		t.Errorf("Expected RUST_BACKTRACE=1 (from level1), got %s", merged.Env["RUST_BACKTRACE"])
	}
	if merged.Env["DATABASE_URL"] != "postgres://localhost" {
		t.Errorf("Expected DATABASE_URL (from level2), got %s", merged.Env["DATABASE_URL"])
	}
}

func TestFindConfigsUpwardNoConfigs(t *testing.T) {
	// Create a directory with no .styx/styx.toml
	tmpDir := t.TempDir()

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	paths, err := findConfigsUpward()
	if err != nil {
		t.Fatalf("findConfigsUpward failed: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("Expected 0 config paths when none exist, got %d", len(paths))
	}
}
