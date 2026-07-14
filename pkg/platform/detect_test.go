package platform

import (
	"testing"
)

func TestDetect(t *testing.T) {
	p := Detect()
	if p.OS == "" || p.Arch == "" {
		t.Fatalf("Detect() returned empty platform: %v", p)
	}
	if !p.IsSupported() {
		t.Fatalf("Detected platform not supported: %s", p)
	}
}

func TestSupportedPlatforms(t *testing.T) {
	platforms := SupportedPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No supported platforms defined")
	}
	if len(platforms) != 3 {
		t.Fatalf("Expected 3 supported platforms, got %d", len(platforms))
	}
}

func TestOverride(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{"linux-x86_64", false},
		{"darwin-arm64", false},
		{"darwin-x86_64", false},
		{"windows-x86_64", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := Override(tt.input)
			if (err != nil) != tt.shouldError {
				t.Fatalf("Override(%q) error = %v, shouldError = %v", tt.input, err, tt.shouldError)
			}
			if err == nil && !p.IsSupported() {
				t.Fatalf("Override(%q) returned unsupported platform: %s", tt.input, p)
			}
		})
	}
}

func TestString(t *testing.T) {
	p := Platform{OS: "linux", Arch: "x86_64"}
	expected := "linux-x86_64"
	if p.String() != expected {
		t.Fatalf("Platform.String() = %q, want %q", p.String(), expected)
	}
}

func BenchmarkDetect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Detect()
	}
}

func BenchmarkOverride(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Override("linux-x86_64")
	}
}

func BenchmarkIsSupported(b *testing.B) {
	p := Platform{OS: "linux", Arch: "x86_64"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.IsSupported()
	}
}
