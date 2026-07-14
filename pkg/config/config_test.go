package config

import (
	"fmt"
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	global := &Config{
		Tools: map[string]string{
			"golang": "1.22.0",
			"node":   "20.10.0",
		},
		Env: map[string]string{
			"RUST_BACKTRACE": "1",
		},
	}

	local := &Config{
		Tools: map[string]string{
			"golang":   "1.21.0",
			"postgres": "15.3",
		},
		Env: map[string]string{
			"DATABASE_URL": "postgresql://localhost/mydb",
		},
	}

	merged := MergeConfigs(global, local)

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"golang override", "golang", "1.21.0"},
		{"node inherit", "node", "20.10.0"},
		{"postgres new", "postgres", "15.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if merged.Tools[tt.version] != tt.expected {
				t.Fatalf("merged.Tools[%q] = %q, want %q", tt.version, merged.Tools[tt.version], tt.expected)
			}
		})
	}

	if merged.Env["RUST_BACKTRACE"] != "1" {
		t.Fatal("Global env var not inherited")
	}
	if merged.Env["DATABASE_URL"] != "postgresql://localhost/mydb" {
		t.Fatal("Local env var not merged")
	}
}

func TestMergeConfigsNil(t *testing.T) {
	merged := MergeConfigs(nil, nil)
	if len(merged.Tools) != 0 || len(merged.Env) != 0 {
		t.Fatal("Merging nil configs should produce empty config")
	}
}

func BenchmarkMergeConfigs(b *testing.B) {
	global := &Config{
		Tools: make(map[string]string),
		Env:   make(map[string]string),
	}
	// Simulate 50 tools and 30 env vars
	for i := 0; i < 50; i++ {
		global.Tools[fmt.Sprintf("tool%d", i)] = fmt.Sprintf("1.%d.0", i)
	}
	for i := 0; i < 30; i++ {
		global.Env[fmt.Sprintf("VAR%d", i)] = fmt.Sprintf("value%d", i)
	}

	local := &Config{
		Tools: make(map[string]string),
		Env:   make(map[string]string),
	}
	// Override 10 tools and add 5 new ones
	for i := 0; i < 10; i++ {
		local.Tools[fmt.Sprintf("tool%d", i)] = fmt.Sprintf("2.%d.0", i)
	}
	for i := 50; i < 55; i++ {
		local.Tools[fmt.Sprintf("tool%d", i)] = fmt.Sprintf("1.%d.0", i)
	}
	for i := 0; i < 5; i++ {
		local.Env[fmt.Sprintf("NEW_VAR%d", i)] = fmt.Sprintf("new_value%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeConfigs(global, local)
	}
}
