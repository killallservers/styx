package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func LoadGlobal() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	path := filepath.Join(home, ".local", "share", "styx", "styx.toml")
	return loadFile(path)
}

func LoadLocal(path string) (*Config, error) {
	if path == "" {
		path = "styx.toml"
	}
	return loadFile(path)
}

func LoadFromPath(path string) (*Config, error) {
	return loadFile(path)
}

func loadFile(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("config path is a directory: %s", path)
	}

	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	if config.Tools == nil {
		config.Tools = make(map[string]string)
	}
	if config.Env == nil {
		config.Env = make(map[string]string)
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
