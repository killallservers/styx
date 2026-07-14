package lock

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	lf := New()
	if lf.Version != "1.0" {
		t.Fatalf("Expected version 1.0, got %s", lf.Version)
	}
	if lf.GeneratedAt.IsZero() {
		t.Fatal("GeneratedAt should be set")
	}
	if len(lf.Tools) != 0 {
		t.Fatal("New lock file should have no tools")
	}
}

func TestAddTool(t *testing.T) {
	lf := New()
	entry := ToolEntry{
		Name:         "ripgrep",
		Version:      "14.1.0",
		Executable:   "rg",
		SourceConfig: "global",
	}

	lf.AddTool(entry)
	if len(lf.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(lf.Tools))
	}
	if lf.Tools[0].Name != "ripgrep" {
		t.Fatalf("Expected tool name ripgrep, got %s", lf.Tools[0].Name)
	}
}

func TestFindTool(t *testing.T) {
	lf := New()
	lf.AddTool(ToolEntry{Name: "ripgrep", Version: "14.1.0"})
	lf.AddTool(ToolEntry{Name: "fd", Version: "10.0.0"})

	found := lf.FindTool("ripgrep")
	if found == nil {
		t.Fatal("FindTool should find ripgrep")
	}
	if found.Version != "14.1.0" {
		t.Fatalf("Expected version 14.1.0, got %s", found.Version)
	}

	notFound := lf.FindTool("nonexistent")
	if notFound != nil {
		t.Fatal("FindTool should return nil for non-existent tool")
	}
}

func TestMarshalJSON(t *testing.T) {
	lf := New()
	lf.GeneratedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	lf.AddTool(ToolEntry{Name: "ripgrep", Version: "14.1.0"})

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var parsed LockFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.Version != "1.0" {
		t.Fatalf("Expected version 1.0, got %s", parsed.Version)
	}
	if len(parsed.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(parsed.Tools))
	}
}

func BenchmarkLockFileNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkLockFileAddTool(b *testing.B) {
	lf := New()
	entry := ToolEntry{
		Name:       "ripgrep",
		Version:    "14.1.0",
		Executable: "rg",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lf.AddTool(entry)
	}
}

func BenchmarkLockFileFindTool(b *testing.B) {
	lf := New()
	for i := 0; i < 100; i++ {
		lf.AddTool(ToolEntry{Name: "tool" + string(rune(i)), Version: "1.0.0"})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lf.FindTool("tool" + string(rune(i%100)))
	}
}

func BenchmarkLockFileMarshalJSON(b *testing.B) {
	lf := New()
	for i := 0; i < 50; i++ {
		lf.AddTool(ToolEntry{Name: "tool" + string(rune(i)), Version: "1.0.0"})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(lf)
	}
}

func BenchmarkLockFileUnmarshalJSON(b *testing.B) {
	lf := New()
	for i := 0; i < 50; i++ {
		lf.AddTool(ToolEntry{Name: "tool" + string(rune(i)), Version: "1.0.0"})
	}
	data, _ := json.Marshal(lf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var parsed LockFile
		json.Unmarshal(data, &parsed)
	}
}
