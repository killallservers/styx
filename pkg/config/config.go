package config

type Config struct {
	Tools      map[string]string `toml:"tools"`
	Env        map[string]string `toml:"env"`
	Registries []Registry        `toml:"registries"`
}

type Registry struct {
	URL     string `toml:"url"`
	Version string `toml:"version"`
}

type Merged struct {
	Tools      map[string]string
	Env        map[string]string
	Registries []Registry
}

func NewConfig() *Config {
	return &Config{
		Tools: make(map[string]string),
		Env:   make(map[string]string),
	}
}

func MergeConfigs(global, local *Config) Merged {
	merged := Merged{
		Tools:      make(map[string]string),
		Env:        make(map[string]string),
		Registries: []Registry{},
	}

	if global != nil {
		for k, v := range global.Tools {
			merged.Tools[k] = v
		}
		for k, v := range global.Env {
			merged.Env[k] = v
		}
		merged.Registries = append(merged.Registries, global.Registries...)
	}

	if local != nil {
		for k, v := range local.Tools {
			merged.Tools[k] = v
		}
		for k, v := range local.Env {
			merged.Env[k] = v
		}
		// Local registries override global (by URL)
		registryMap := make(map[string]Registry)
		for _, reg := range merged.Registries {
			registryMap[reg.URL] = reg
		}
		for _, reg := range local.Registries {
			registryMap[reg.URL] = reg
		}
		merged.Registries = make([]Registry, 0, len(registryMap))
		for _, reg := range registryMap {
			merged.Registries = append(merged.Registries, reg)
		}
	}

	return merged
}
