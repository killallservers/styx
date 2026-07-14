package registry

import (
	"fmt"

	"github.com/killallservers/styx/pkg/config"
)

// ResolveRegistry loads the registry using configured HTTP URL or embedded fallback.
// Respects config.Registries settings from styx.toml.
func ResolveRegistry(cfg config.Merged) (map[string]*ToolSpec, error) {
	// If registries are configured, use the first one
	if len(cfg.Registries) > 0 {
		registryURL := cfg.Registries[0].URL
		if registryURL != "" {
			// Try HTTP registry with fallback
			return LoadWithHTTPFallback(registryURL)
		}
	}

	// No HTTP registry configured, use embedded
	return LoadEmbeddedRegistry()
}

// GetRegistryStatus returns a human-friendly status message about which registry is in use.
func GetRegistryStatus(cfg config.Merged) string {
	if len(cfg.Registries) > 0 && cfg.Registries[0].URL != "" {
		return fmt.Sprintf("using remote registry: %s", cfg.Registries[0].URL)
	}
	return "using embedded registry"
}
