package registry

type ToolSpec struct {
	Name       string                 `toml:"name"`
	Repository string                 `toml:"repository"`
	Versions   map[string]VersionSpec `toml:"versions"`
}

type VersionSpec struct {
	Released  string                `toml:"released"`
	Stability string                `toml:"stability"`
	Methods   []InstallMethod       `toml:"methods"`
	LLMReason *LLMReasoningMetadata `toml:"llm_reasoning"`
}

type InstallMethod struct {
	Type       string            `toml:"type"`
	Platforms  map[string]string `toml:"platforms"`
	Checksums  map[string]string `toml:"checksums"`
	Executable string            `toml:"executable"`
}

type LLMReasoningMetadata struct {
	Timestamp string `toml:"timestamp"`
	Model     string `toml:"model"`
	Reasoning string `toml:"reasoning"`
}

func (s *ToolSpec) GetVersion(version string) *VersionSpec {
	if v, ok := s.Versions[version]; ok {
		return &v
	}
	return nil
}

func (v *VersionSpec) GetMethodForPlatform(platformStr string, methodType string) *InstallMethod {
	for _, m := range v.Methods {
		if m.Type != methodType {
			continue
		}
		if _, ok := m.Platforms[platformStr]; ok {
			return &m
		}
	}
	return nil
}
