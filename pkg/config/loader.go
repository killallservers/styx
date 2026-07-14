package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadHierarchical loads and merges all styx.toml files from current directory up to root,
// plus the global config at ~/.local/share/styx/styx.toml.
//
// Merge order (lowest to highest priority):
// 1. Global config (~/.local/share/styx/styx.toml)
// 2. Root-most .styx/styx.toml (if in nested dirs)
// 3. ...intermediate directories...
// 4. Current directory .styx/styx.toml (highest priority - closest to user)
//
// Each level can override tools and env vars from previous levels.
func LoadHierarchical() (Merged, error) {
	// Start with global config
	globalConfig, err := LoadGlobal()
	if err != nil && !os.IsNotExist(err) {
		return Merged{}, fmt.Errorf("failed to load global config: %w", err)
	}
	if globalConfig == nil {
		globalConfig = NewConfig()
	}

	// Collect all .styx/styx.toml files from current dir to root
	configs := []*Config{globalConfig}

	paths, err := findConfigsUpward()
	if err != nil {
		// Log but don't fail - missing configs are OK
		paths = []string{}
	}

	// Load each config file
	for _, path := range paths {
		cfg, err := loadFile(path)
		if err != nil {
			return Merged{}, fmt.Errorf("failed to load config at %s: %w", path, err)
		}
		if cfg != nil {
			configs = append(configs, cfg)
		}
	}

	// Merge all configs in order (global first, current dir last)
	merged := Merged{
		Tools: make(map[string]string),
		Env:   make(map[string]string),
	}

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		// Merge tools: later configs override earlier ones
		for k, v := range cfg.Tools {
			merged.Tools[k] = v
		}
		// Merge env: later configs override earlier ones
		for k, v := range cfg.Env {
			merged.Env[k] = v
		}
	}

	return merged, nil
}

// findConfigsUpward walks from current directory to root, collecting all .styx/styx.toml file paths.
// Returns paths in order from root to current (farthest to closest).
func findConfigsUpward() ([]string, error) {
	var paths []string

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	current := cwd

	// Walk up from current directory to filesystem root
	for {
		configPath := filepath.Join(current, ".styx", "styx.toml")

		// Check if config exists at this level
		if _, err := os.Stat(configPath); err == nil {
			paths = append(paths, configPath)
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			break
		}
		current = parent
	}

	// Reverse so farthest (root-most) is first
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}

	return paths, nil
}

// LoadLocalWithFallback loads hierarchical configs, with fallback to direct path if provided.
// Used for backward compatibility during migration.
func LoadLocalWithFallback(directPath string) (Merged, error) {
	// Try hierarchical loading first
	merged, err := LoadHierarchical()
	if err == nil && (len(merged.Tools) > 0 || len(merged.Env) > 0) {
		return merged, nil
	}

	// Fallback to direct path if specified
	if directPath != "" {
		cfg, err := LoadFromPath(directPath)
		if err != nil && !os.IsNotExist(err) {
			return Merged{}, fmt.Errorf("failed to load config from path: %w", err)
		}
		if cfg != nil {
			return MergeConfigs(nil, cfg), nil
		}
	}

	// Return empty merged config if nothing found
	return Merged{
		Tools: make(map[string]string),
		Env:   make(map[string]string),
	}, nil
}
