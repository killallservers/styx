package registry

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
)

//go:embed specs/*.toml
var specsFS embed.FS

// LoadEmbeddedRegistry loads and parses all TOML specs from embedded filesystem.
func LoadEmbeddedRegistry() (map[string]*ToolSpec, error) {
	registry := make(map[string]*ToolSpec)

	// List all TOML files in specs directory
	entries, err := specsFS.ReadDir("specs")
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Load each spec file
	var toolNames []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		toolName := entry.Name()[:len(entry.Name())-5] // Remove .toml
		toolNames = append(toolNames, toolName)
	}

	// Sort for consistent ordering
	sort.Strings(toolNames)

	// Parse each spec
	for _, toolName := range toolNames {
		specPath := filepath.Join("specs", toolName+".toml")
		data, err := specsFS.ReadFile(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec %q: %w", toolName, err)
		}

		spec, err := ParseToolSpecFromTOML(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec %q: %w", toolName, err)
		}

		// Validate spec structure
		if err := ValidateToolSpec(spec); err != nil {
			return nil, fmt.Errorf("invalid spec %q: %w", toolName, err)
		}

		registry[toolName] = spec
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
