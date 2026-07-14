package platform

import (
	"fmt"
	"os"
	"runtime"
)

type Platform struct {
	OS   string
	Arch string
}

func (p Platform) String() string {
	return fmt.Sprintf("%s-%s", p.OS, p.Arch)
}

var supportedPlatforms = []Platform{
	{OS: "linux", Arch: "x86_64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "darwin", Arch: "x86_64"},
}

func Detect() Platform {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	if osName == "linux" {
		return Platform{OS: "linux", Arch: "x86_64"}
	}
	if osName == "darwin" {
		if archName == "arm64" {
			return Platform{OS: "darwin", Arch: "arm64"}
		}
		return Platform{OS: "darwin", Arch: "x86_64"}
	}

	return Platform{OS: osName, Arch: archName}
}

func (p Platform) IsSupported() bool {
	for _, supported := range supportedPlatforms {
		if p == supported {
			return true
		}
	}
	return false
}

func Override(override string) (Platform, error) {
	if override == "" {
		return Detect(), nil
	}

	for _, p := range supportedPlatforms {
		if p.String() == override {
			return p, nil
		}
	}

	return Platform{}, fmt.Errorf("unsupported platform: %s", override)
}

func SupportedPlatforms() []Platform {
	return supportedPlatforms
}

func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}
