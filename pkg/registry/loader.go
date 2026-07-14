package registry

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// TomlToolSpec is the TOML representation of a ToolSpec.
type TomlToolSpec struct {
	Name       string          `toml:"name"`
	Repository string          `toml:"repository"`
	Versions   []TomlVersion   `toml:"versions"`
}

// TomlVersion is the TOML representation of a VersionSpec.
type TomlVersion struct {
	Version   string                `toml:"version"`
	Released  string                `toml:"released"`
	Stability string                `toml:"stability"`
	Methods   map[string]TomlMethod `toml:"methods"`
}

// TomlMethod is the TOML representation of an InstallMethod.
type TomlMethod struct {
	Type       string            `toml:"type"`
	Executable string            `toml:"executable"`
	Platforms  map[string]string `toml:"platforms"`
	Checksums  map[string]string `toml:"checksums"`
}

// ParseToolSpecFromTOML parses a TOML string into a ToolSpec.
func ParseToolSpecFromTOML(data string) (*ToolSpec, error) {
	var tomlSpec TomlToolSpec
	if err := toml.Unmarshal([]byte(data), &tomlSpec); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Convert TOML representation to ToolSpec
	spec := &ToolSpec{
		Name:       tomlSpec.Name,
		Repository: tomlSpec.Repository,
		Versions:   make(map[string]VersionSpec),
	}

	for _, tomlVer := range tomlSpec.Versions {
		methods := []InstallMethod{}

		// Convert methods from map to slice
		for methodName, tomlMethod := range tomlVer.Methods {
			if tomlMethod.Type == "" {
				tomlMethod.Type = methodName
			}

			method := InstallMethod{
				Type:       tomlMethod.Type,
				Platforms:  tomlMethod.Platforms,
				Checksums:  tomlMethod.Checksums,
				Executable: tomlMethod.Executable,
			}
			methods = append(methods, method)
		}

		spec.Versions[tomlVer.Version] = VersionSpec{
			Released:  tomlVer.Released,
			Stability: tomlVer.Stability,
			Methods:   methods,
		}
	}

	return spec, nil
}

// ValidateToolSpec checks that a ToolSpec is well-formed.
func ValidateToolSpec(spec *ToolSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("tool spec missing name")
	}

	if len(spec.Versions) == 0 {
		return fmt.Errorf("tool spec for %q has no versions", spec.Name)
	}

	for version, versionSpec := range spec.Versions {
		if version == "" {
			return fmt.Errorf("tool %q has empty version string", spec.Name)
		}

		if len(versionSpec.Methods) == 0 {
			return fmt.Errorf("tool %q version %q has no install methods", spec.Name, version)
		}

		for _, method := range versionSpec.Methods {
			if method.Type == "" {
				return fmt.Errorf("tool %q version %q has method with no type", spec.Name, version)
			}

			if len(method.Platforms) == 0 {
				return fmt.Errorf("tool %q version %q method %q has no platforms", spec.Name, version, method.Type)
			}

			if len(method.Checksums) == 0 {
				return fmt.Errorf("tool %q version %q method %q has no checksums", spec.Name, version, method.Type)
			}

			if method.Executable == "" {
				return fmt.Errorf("tool %q version %q method %q missing executable", spec.Name, version, method.Type)
			}

			// Validate platform/checksum pairs match
			for platform := range method.Platforms {
				if _, ok := method.Checksums[platform]; !ok {
					return fmt.Errorf("tool %q version %q platform %q missing checksum", spec.Name, version, platform)
				}
			}

			for platform := range method.Checksums {
				if _, ok := method.Platforms[platform]; !ok {
					return fmt.Errorf("tool %q version %q checksum for unknown platform %q", spec.Name, version, platform)
				}
			}
		}
	}

	return nil
}

// VersionInfo holds version and released date for sorting.
type VersionInfo struct {
	Version  string
	Released string
}

// GetLatestVersionInfo returns the latest version based on semantic versioning.
func GetLatestVersionInfo(spec *ToolSpec) (VersionInfo, error) {
	if len(spec.Versions) == 0 {
		return VersionInfo{}, fmt.Errorf("no versions available")
	}

	var latest VersionInfo
	for version, versionSpec := range spec.Versions {
		info := VersionInfo{
			Version:  version,
			Released: versionSpec.Released,
		}

		if latest.Version == "" || isGreaterVersion(version, latest.Version) {
			latest = info
		}
	}

	return latest, nil
}

// isGreaterVersion does simple semantic version comparison.
func isGreaterVersion(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 > n2 {
			return true
		}
		if n1 < n2 {
			return false
		}
	}

	return false
}
